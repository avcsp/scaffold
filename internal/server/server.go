package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"scaffold/internal/router"
)

const shutdownTimeout = 30 * time.Second

// Server wraps an HTTP server with graceful shutdown and probe endpoints.
type Server struct {
	httpServer *http.Server
	router     *router.Router
	ready      atomic.Bool
	hooks      []func()
}

// New creates a Server that binds the router to the given address
// and registers /probe/live and /probe/ready endpoints.
func New(r *router.Router, addr string) *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: r.Handler(),
		},
		router: r,
	}

	s.ready.Store(true)

	r.Get("/probe/live", s.liveHandler)
	r.Get("/probe/ready", s.readyHandler)

	return s
}

// OnShutdown registers a function to be called during graceful shutdown
// after HTTP request draining completes. Hooks run in registration order.
func (s *Server) OnShutdown(fn func()) {
	s.hooks = append(s.hooks, fn)
}

// Start begins listening and blocks until shutdown completes.
// On SIGTERM/SIGINT it marks the server as not ready, drains in-flight
// requests, runs shutdown hooks, and exits.
func (s *Server) Start() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("server listening on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop
	log.Println("shutdown signal received")

	// Mark not ready so k8s stops routing traffic
	s.ready.Store(false)

	// Drain in-flight requests
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
	log.Println("http server drained")

	// Run shutdown hooks
	for _, fn := range s.hooks {
		fn()
	}
	log.Println("shutdown complete")
}

func (s *Server) liveHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) readyHandler(w http.ResponseWriter, _ *http.Request) {
	if s.ready.Load() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("shutting down"))
}
