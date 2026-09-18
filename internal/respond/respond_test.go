package respond

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"name": "shyamin"}

	JSON(rec, http.StatusOK, data)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}

	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if result["name"] != "shyamin" {
		t.Fatalf("expected 'shyamin', got %q", result["name"])
	}
}

func TestJSONWithStruct(t *testing.T) {
	type user struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	rec := httptest.NewRecorder()
	JSON(rec, http.StatusCreated, user{ID: 1, Name: "test"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var result user
	json.NewDecoder(rec.Body).Decode(&result)
	if result.ID != 1 || result.Name != "test" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()
	Error(rec, http.StatusNotFound, "user not found")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if result["error"] != "user not found" {
		t.Fatalf("expected 'user not found', got %q", result["error"])
	}
}

func TestBindJSON(t *testing.T) {
	type input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	body := `{"name":"shyamin","email":"me@shyamin.com"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))

	var dst input
	if err := BindJSON(req, &dst); err != nil {
		t.Fatal(err)
	}

	if dst.Name != "shyamin" || dst.Email != "me@shyamin.com" {
		t.Fatalf("unexpected result: %+v", dst)
	}
}

func TestBindJSONInvalidBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader("not json"))

	var dst map[string]string
	if err := BindJSON(req, &dst); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestBindJSONEmptyBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(""))

	var dst map[string]string
	if err := BindJSON(req, &dst); err == nil {
		t.Fatal("expected error for empty body")
	}
}
