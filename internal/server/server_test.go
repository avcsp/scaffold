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

	sctx "scaffold/internal/context"
	"scaffold/internal/router"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func newTestServer() *Server {
	r := router.New()
	r.Get("/test", func(c *sctx.Context) {
		c.String(http.StatusOK, "ok")
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

	time.Sleep(50 * time.Millisecond)

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

	r.Get("/slow", func(c *sctx.Context) {
		close(requestStarted)
		<-requestDone
		c.String(http.StatusOK, "completed")
	})

	s := New(r, ":0", discardLogger())

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

	respCh := make(chan *http.Response, 1)
	go func() {
		resp, err := http.Get("http://" + addr + "/slow")
		if err != nil {
			t.Logf("GET error: %v", err)
			return
		}
		respCh <- resp
	}()

	<-requestStarted

	s.ready.Store(false)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go s.httpServer.Shutdown(ctx)

	if s.ready.Load() {
		t.Fatal("expected ready to be false after shutdown")
	}

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

	r.Get("/slow", func(c *sctx.Context) {
		close(requestStarted)
		<-requestDone
		timeline = append(timeline, "request-done")
		c.String(http.StatusOK, "ok")
	})

	s := New(r, ":0", discardLogger())
	s.OnShutdown(func() {
		timeline = append(timeline, "hook-ran")
	})

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()

	serverDone := make(chan struct{})
	go func() {
		s.httpServer.Serve(ln)
	}()

	go func() {
		http.Get("http://" + addr + "/slow")
	}()

	<-requestStarted

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

	close(requestDone)

	select {
	case <-hooksDone:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not complete within 5s")
	}

	if len(timeline) != 2 || timeline[0] != "request-done" || timeline[1] != "hook-ran" {
		t.Fatalf("expected [request-done hook-ran], got %v", timeline)
	}
}
