package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sctx "scaffold/internal/context"
)

func TestBasicRoute(t *testing.T) {
	r := New()
	r.Get("/health", func(c *sctx.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "ok" {
		t.Fatalf("expected 'ok', got %q", body)
	}
}

func TestAllMethods(t *testing.T) {
	r := New()
	handler := func(c *sctx.Context) {
		c.Status(http.StatusOK)
	}

	r.Get("/r", handler)
	r.Post("/r", handler)
	r.Put("/r", handler)
	r.Patch("/r", handler)
	r.Delete("/r", handler)

	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		req := httptest.NewRequest(method, "/r", nil)
		rec := httptest.NewRecorder()
		r.mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s /r: expected 200, got %d", method, rec.Code)
		}
	}
}

func TestGlobalMiddleware(t *testing.T) {
	r := New()
	r.Use(func(next sctx.HandlerFunc) sctx.HandlerFunc {
		return func(c *sctx.Context) {
			c.Header("X-Global", "applied")
			next(c)
		}
	})
	r.Get("/test", func(c *sctx.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Header().Get("X-Global") != "applied" {
		t.Fatal("global middleware was not applied")
	}
}

func TestMiddlewareOrder(t *testing.T) {
	r := New()
	var order []string

	makeMW := func(name string) sctx.Middleware {
		return func(next sctx.HandlerFunc) sctx.HandlerFunc {
			return func(c *sctx.Context) {
				order = append(order, name+"-before")
				next(c)
				order = append(order, name+"-after")
			}
		}
	}

	r.Use(makeMW("A"))
	r.Use(makeMW("B"))
	r.Get("/order", func(c *sctx.Context) {
		order = append(order, "handler")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/order", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	expected := []string{"A-before", "B-before", "handler", "B-after", "A-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("position %d: expected %q, got %q", i, expected[i], order[i])
		}
	}
}

func TestGroupPrefix(t *testing.T) {
	r := New()
	r.Group("/api/v1", func(g *Group) {
		g.Get("/users", func(c *sctx.Context) {
			c.String(http.StatusOK, "users")
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "users" {
		t.Fatalf("expected 'users', got %q", body)
	}
}

func TestGroupMiddleware(t *testing.T) {
	r := New()
	r.Use(func(next sctx.HandlerFunc) sctx.HandlerFunc {
		return func(c *sctx.Context) {
			c.Header("X-Global", "yes")
			next(c)
		}
	})

	r.Get("/public", func(c *sctx.Context) {
		c.Status(http.StatusOK)
	})

	r.Group("/admin", func(g *Group) {
		g.Use(func(next sctx.HandlerFunc) sctx.HandlerFunc {
			return func(c *sctx.Context) {
				c.Header("X-Auth", "yes")
				next(c)
			}
		})
		g.Get("/dashboard", func(c *sctx.Context) {
			c.Status(http.StatusOK)
		})
	})

	// Public route: global middleware only
	req := httptest.NewRequest("GET", "/public", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)
	if rec.Header().Get("X-Global") != "yes" {
		t.Fatal("public: global middleware missing")
	}
	if rec.Header().Get("X-Auth") != "" {
		t.Fatal("public: group middleware should not apply")
	}

	// Admin route: global + group middleware
	req = httptest.NewRequest("GET", "/admin/dashboard", nil)
	rec = httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)
	if rec.Header().Get("X-Global") != "yes" {
		t.Fatal("admin: global middleware missing")
	}
	if rec.Header().Get("X-Auth") != "yes" {
		t.Fatal("admin: group middleware missing")
	}
}

func TestNestedGroups(t *testing.T) {
	r := New()
	r.Group("/api", func(g *Group) {
		g.Group("/v1", func(g2 *Group) {
			g2.Get("/items", func(c *sctx.Context) {
				c.String(http.StatusOK, "items")
			})
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/items", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "items" {
		t.Fatalf("expected 'items', got %q", body)
	}
}

func TestNestedGroupMiddleware(t *testing.T) {
	var order []string
	makeMW := func(name string) sctx.Middleware {
		return func(next sctx.HandlerFunc) sctx.HandlerFunc {
			return func(c *sctx.Context) {
				order = append(order, name)
				next(c)
			}
		}
	}

	r := New()
	r.Use(makeMW("global"))
	r.Group("/a", func(g *Group) {
		g.Use(makeMW("group-a"))
		g.Group("/b", func(g2 *Group) {
			g2.Use(makeMW("group-b"))
			g2.Get("/c", func(c *sctx.Context) {
				order = append(order, "handler")
				c.Status(http.StatusOK)
			})
		})
	})

	req := httptest.NewRequest("GET", "/a/b/c", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	expected := []string{"global", "group-a", "group-b", "handler"}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("position %d: expected %q, got %q", i, expected[i], order[i])
		}
	}
}

func TestPerRouteMiddleware(t *testing.T) {
	r := New()

	routeMW := func(next sctx.HandlerFunc) sctx.HandlerFunc {
		return func(c *sctx.Context) {
			c.Header("X-Route", "yes")
			next(c)
		}
	}

	r.Get("/with-mw", func(c *sctx.Context) {
		c.Status(http.StatusOK)
	}, routeMW)

	r.Get("/without-mw", func(c *sctx.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/with-mw", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)
	if rec.Header().Get("X-Route") != "yes" {
		t.Fatal("per-route middleware was not applied")
	}

	req = httptest.NewRequest("GET", "/without-mw", nil)
	rec = httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)
	if rec.Header().Get("X-Route") != "" {
		t.Fatal("per-route middleware should not leak to other routes")
	}
}

func TestAllThreeLevels(t *testing.T) {
	var order []string
	makeMW := func(name string) sctx.Middleware {
		return func(next sctx.HandlerFunc) sctx.HandlerFunc {
			return func(c *sctx.Context) {
				order = append(order, name)
				next(c)
			}
		}
	}

	r := New()
	r.Use(makeMW("global"))
	r.Group("/api", func(g *Group) {
		g.Use(makeMW("group"))
		g.Get("/data", func(c *sctx.Context) {
			order = append(order, "handler")
			c.Status(http.StatusOK)
		}, makeMW("route"))
	})

	req := httptest.NewRequest("GET", "/api/data", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	expected := []string{"global", "group", "route", "handler"}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("position %d: expected %q, got %q", i, expected[i], order[i])
		}
	}
}

func TestPathParams(t *testing.T) {
	r := New()
	r.Get("/users/{id}", func(c *sctx.Context) {
		c.String(http.StatusOK, "user:"+c.Param("id"))
	})

	req := httptest.NewRequest("GET", "/users/42", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "user:42" {
		t.Fatalf("expected 'user:42', got %q", body)
	}
}

func TestGroupPathParams(t *testing.T) {
	r := New()
	r.Group("/api", func(g *Group) {
		g.Get("/users/{id}/posts/{postID}", func(c *sctx.Context) {
			c.String(http.StatusOK, "user:"+c.Param("id")+":post:"+c.Param("postID"))
		})
	})

	req := httptest.NewRequest("GET", "/api/users/5/posts/99", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "user:5:post:99" {
		t.Fatalf("expected 'user:5:post:99', got %q", body)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	r := New()
	r.Get("/only-get", func(c *sctx.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/only-get", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestNotFound(t *testing.T) {
	r := New()
	r.Get("/exists", func(c *sctx.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/nope", nil)
	rec := httptest.NewRecorder()
	r.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
