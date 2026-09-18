package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	sctx "scaffold/internal/context"
)

func corsHandler(c *sctx.Context) {
	c.String(http.StatusOK, "ok")
}

func serveCORS(origin string) *httptest.ResponseRecorder {
	handler := CORS(corsHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	c := sctx.New(rec, req)
	handler(c)
	return rec
}

func TestCORSDefaultAllowAll(t *testing.T) {
	os.Unsetenv("CORS_ORIGINS")
	rec := serveCORS("http://example.com")

	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "*" {
		t.Fatalf("expected '*', got %q", v)
	}
}

func TestCORSSpecificOriginAllowed(t *testing.T) {
	os.Setenv("CORS_ORIGINS", "http://foo.com,http://bar.com")
	defer os.Unsetenv("CORS_ORIGINS")

	rec := serveCORS("http://bar.com")

	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "http://bar.com" {
		t.Fatalf("expected 'http://bar.com', got %q", v)
	}
}

func TestCORSOriginNotAllowed(t *testing.T) {
	os.Setenv("CORS_ORIGINS", "http://foo.com")
	defer os.Unsetenv("CORS_ORIGINS")

	rec := serveCORS("http://evil.com")

	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "" {
		t.Fatalf("expected empty, got %q", v)
	}
}

func TestCORSPreflight(t *testing.T) {
	os.Unsetenv("CORS_ORIGINS")

	handler := CORS(corsHandler)
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	c := sctx.New(rec, req)
	handler(c)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "*" {
		t.Fatalf("expected '*', got %q", v)
	}
	if v := rec.Header().Get("Access-Control-Allow-Methods"); v == "" {
		t.Fatal("expected Allow-Methods header")
	}
	if v := rec.Header().Get("Access-Control-Allow-Headers"); v == "" {
		t.Fatal("expected Allow-Headers header")
	}
	if v := rec.Header().Get("Access-Control-Max-Age"); v == "" {
		t.Fatal("expected Max-Age header")
	}
	// Body should be empty for preflight
	if rec.Body.String() != "" {
		t.Fatalf("expected empty body for preflight, got %q", rec.Body.String())
	}
}

func TestCORSPreflightDoesNotCallNext(t *testing.T) {
	os.Unsetenv("CORS_ORIGINS")
	called := false

	handler := CORS(func(c *sctx.Context) {
		called = true
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	c := sctx.New(rec, req)
	handler(c)

	if called {
		t.Fatal("preflight should not call next handler")
	}
}

func TestCORSCustomMethods(t *testing.T) {
	os.Setenv("CORS_METHODS", "GET,POST")
	defer os.Unsetenv("CORS_METHODS")

	rec := serveCORS("http://example.com")

	if v := rec.Header().Get("Access-Control-Allow-Methods"); v != "GET,POST" {
		t.Fatalf("expected 'GET,POST', got %q", v)
	}
}

func TestCORSCustomHeaders(t *testing.T) {
	os.Setenv("CORS_HEADERS", "X-Custom,Authorization")
	defer os.Unsetenv("CORS_HEADERS")

	rec := serveCORS("http://example.com")

	if v := rec.Header().Get("Access-Control-Allow-Headers"); v != "X-Custom,Authorization" {
		t.Fatalf("expected 'X-Custom,Authorization', got %q", v)
	}
}

func TestCORSNoOriginHeaderWithWildcard(t *testing.T) {
	os.Unsetenv("CORS_ORIGINS")

	rec := serveCORS("")

	// Wildcard allows everything, even requests without Origin
	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "*" {
		t.Fatalf("expected '*', got %q", v)
	}
}

func TestCORSNoOriginHeaderWithSpecificOrigins(t *testing.T) {
	os.Setenv("CORS_ORIGINS", "http://foo.com")
	defer os.Unsetenv("CORS_ORIGINS")

	rec := serveCORS("")

	// No Origin header with specific origins should not set CORS headers
	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "" {
		t.Fatalf("expected empty, got %q", v)
	}
}

func TestCORSOriginCaseInsensitive(t *testing.T) {
	os.Setenv("CORS_ORIGINS", "http://Foo.Com")
	defer os.Unsetenv("CORS_ORIGINS")

	rec := serveCORS("http://foo.com")

	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "http://foo.com" {
		t.Fatalf("expected 'http://foo.com', got %q", v)
	}
}
