package middleware

import (
	"net/http"
	"strings"

	"scaffold/internal/config"
	sctx "scaffold/internal/context"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
//
// Configuration via env vars:
//   - CORS_ORIGINS: comma-separated allowed origins (default "*")
//   - CORS_METHODS: comma-separated allowed methods (default "GET,POST,PUT,PATCH,DELETE,OPTIONS")
//   - CORS_HEADERS: comma-separated allowed headers (default "Content-Type,Authorization")
//   - CORS_MAX_AGE: preflight cache duration in seconds (default "86400")
func CORS(next sctx.HandlerFunc) sctx.HandlerFunc {
	origins := config.Getenv("CORS_ORIGINS", "*")
	methods := config.Getenv("CORS_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	headers := config.Getenv("CORS_HEADERS", "Content-Type,Authorization")
	maxAge := config.Getenv("CORS_MAX_AGE", "86400")

	allowedOrigins := strings.Split(origins, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	return func(c *sctx.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := matchOrigin(origin, allowedOrigins)

		if allowed != "" {
			c.Header("Access-Control-Allow-Origin", allowed)
			c.Header("Access-Control-Allow-Methods", methods)
			c.Header("Access-Control-Allow-Headers", headers)
			c.Header("Access-Control-Max-Age", maxAge)
		}

		// Handle preflight
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			return
		}

		next(c)
	}
}

func matchOrigin(origin string, allowed []string) string {
	for _, a := range allowed {
		if a == "*" {
			return "*"
		}
		if strings.EqualFold(a, origin) {
			return origin
		}
	}
	return ""
}
