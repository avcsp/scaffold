package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	c := New(rec, httptest.NewRequest("GET", "/", nil))

	c.JSON(http.StatusOK, map[string]string{"name": "test"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}
	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if result["name"] != "test" {
		t.Fatalf("expected 'test', got %q", result["name"])
	}
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()
	c := New(rec, httptest.NewRequest("GET", "/", nil))

	c.Error(http.StatusBadRequest, "bad input")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if result["error"] != "bad input" {
		t.Fatalf("expected 'bad input', got %q", result["error"])
	}
}

func TestAbortWithStatusJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	c := New(rec, httptest.NewRequest("GET", "/", nil))

	c.AbortWithStatusJSON(http.StatusForbidden, map[string]string{"error": "forbidden"})

	if !c.IsAborted() {
		t.Fatal("expected context to be aborted")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAbort(t *testing.T) {
	c := New(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if c.IsAborted() {
		t.Fatal("should not be aborted initially")
	}
	c.Abort()
	if !c.IsAborted() {
		t.Fatal("should be aborted after Abort()")
	}
}

func TestBindJSON(t *testing.T) {
	type input struct {
		Name string `json:"name"`
	}

	body := `{"name":"test"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	c := New(httptest.NewRecorder(), req)

	var dst input
	if err := c.BindJSON(&dst); err != nil {
		t.Fatal(err)
	}
	if dst.Name != "test" {
		t.Fatalf("expected 'test', got %q", dst.Name)
	}
}

func TestBindJSONInvalid(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader("not json"))
	c := New(httptest.NewRecorder(), req)

	var dst map[string]string
	if err := c.BindJSON(&dst); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/users/42", nil)
	req.SetPathValue("id", "42")
	c := New(httptest.NewRecorder(), req)

	if v := c.Param("id"); v != "42" {
		t.Fatalf("expected '42', got %q", v)
	}
}

func TestSetAndGet(t *testing.T) {
	c := New(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	c.Set("db", "fake-db-conn")

	val, ok := c.Get("db")
	if !ok {
		t.Fatal("expected key 'db' to exist")
	}
	if val != "fake-db-conn" {
		t.Fatalf("expected 'fake-db-conn', got %v", val)
	}
}

func TestGetMissing(t *testing.T) {
	c := New(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	_, ok := c.Get("nope")
	if ok {
		t.Fatal("expected key 'nope' to not exist")
	}
}

func TestMustGet(t *testing.T) {
	c := New(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	c.Set("key", 42)

	val := c.MustGet("key")
	if val != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestMustGetPanics(t *testing.T) {
	c := New(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for missing key")
		}
	}()

	c.MustGet("missing")
}

func TestQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/search?q=hello&page=2", nil)
	c := New(httptest.NewRecorder(), req)

	if v := c.Query("q"); v != "hello" {
		t.Fatalf("expected 'hello', got %q", v)
	}
	if v := c.Query("page"); v != "2" {
		t.Fatalf("expected '2', got %q", v)
	}
	if v := c.Query("missing"); v != "" {
		t.Fatalf("expected empty, got %q", v)
	}
}

func TestStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	c := New(rec, httptest.NewRequest("GET", "/", nil))

	c.Status(http.StatusNoContent)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	c := New(rec, httptest.NewRequest("GET", "/", nil))

	c.Header("X-Custom", "value")

	if rec.Header().Get("X-Custom") != "value" {
		t.Fatalf("expected 'value', got %q", rec.Header().Get("X-Custom"))
	}
}

func TestString(t *testing.T) {
	rec := httptest.NewRecorder()
	c := New(rec, httptest.NewRequest("GET", "/", nil))

	c.String(http.StatusOK, "hello world")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("expected text/plain, got %q", ct)
	}
	if rec.Body.String() != "hello world" {
		t.Fatalf("expected 'hello world', got %q", rec.Body.String())
	}
}
