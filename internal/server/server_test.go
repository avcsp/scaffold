package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"

	"scaffold/internal/router"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func newTestServer() *Server {
	r := router.New()
	r.Get("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	return New(r, ":0", discardLogger())
}

func TestLiveProbe(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest("GET", "/probe/live", nil)
	rec := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "ok" {
		t.Fatalf("expected 'ok', got %q", body)
	}
}

func TestReadyProbeWhenReady(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest("GET", "/probe/ready", nil)
	rec := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "ok" {
		t.Fatalf("expected 'ok', got %q", body)
	}
}

func TestReadyProbeWhenNotReady(t *testing.T) {
	s := newTestServer()
	s.ready.Store(false)

	req := httptest.NewRequest("GET", "/probe/ready", nil)
	rec := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "shutting down" {
		t.Fatalf("expected 'shutting down', got %q", body)
	}
}

func TestRoutesWorkThroughServer(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	s.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "ok" {
		t.Fatalf("expected 'ok', got %q", body)
	}
}

func TestOnShutdownHooks(t *testing.T) {
	s := newTestServer()

	var order []string
	s.OnShutdown(func() { order = append(order, "first") })
	s.OnShutdown(func() { order = append(order, "second") })

	done := make(chan struct{})
	go func() {
		s.Start()
		close(done)
	}()

	// Give the server a moment to start listening
	time.Sleep(50 * time.Millisecond)

	// Send SIGTERM
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within 5s")
	}

	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Fatalf("expected [first second], got %v", order)
	}
}

func TestGracefulShutdownDrainsRequests(t *testing.T) {
	r := router.New()
	requestStarted := make(chan struct{})
	requestDone := make(chan struct{})

	r.Get("/slow", func(w http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		<-requestDone
		w.Write([]byte("completed"))
	})

	s := New(r, ":0", discardLogger())

	// Use a real listener so http.Server tracks the connection
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()

	shutdownComplete := make(chan struct{})
	go func() {
		s.httpServer.Serve(ln)
		close(shutdownComplete)
	}()

	// Start a slow request over TCP
	respCh := make(chan *http.Response, 1)
	go func() {
		resp, err := http.Get("http://" + addr + "/slow")
		if err != nil {
			t.Logf("GET error: %v", err)
			return
		}
		respCh <- resp
	}()

	// Wait for the request to be in-flight
	<-requestStarted

	// Trigger shutdown while request is in-flight
	s.ready.Store(false)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go s.httpServer.Shutdown(ctx)

	// Ready probe should be false
	if s.ready.Load() {
		t.Fatal("expected ready to be false after shutdown")
	}

	// Let the in-flight request complete
	close(requestDone)

	select {
	case <-shutdownComplete:
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down within 5s")
	}

	resp := <-respCh
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "completed" {
		t.Fatalf("expected 'completed', got %q", string(body))
	}
}

func TestShutdownHooksRunAfterDrain(t *testing.T) {
	r := router.New()
	requestStarted := make(chan struct{})
	requestDone := make(chan struct{})
	var timeline []string

	r.Get("/slow", func(w http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		<-requestDone
		timeline = append(timeline, "request-done")
		w.Write([]byte("ok"))
	})

	s := New(r, ":0", discardLogger())
	s.OnShutdown(func() {
		timeline = append(timeline, "hook-ran")
	})

	// Use a real listener
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()

	serverDone := make(chan struct{})
	go func() {
		s.httpServer.Serve(ln)
	}()

	// Start slow request over TCP
	go func() {
		http.Get("http://" + addr + "/slow")
	}()

	<-requestStarted

	// Trigger shutdown
	hooksDone := make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
		for _, fn := range s.hooks {
			fn()
		}
		close(hooksDone)
		close(serverDone)
	}()

	time.Sleep(50 * time.Millisecond)

	// Release the request
	close(requestDone)

	select {
	case <-hooksDone:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not complete within 5s")
	}

	// Hook must run after request completes
	if len(timeline) != 2 || timeline[0] != "request-done" || timeline[1] != "hook-ran" {
		t.Fatalf("expected [request-done hook-ran], got %v", timeline)
	}
}
