package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery catches panics in handlers, logs the stack trace,
// and returns a 500 Internal Server Error response.
// Works at the http.Handler level so it wraps the entire mux.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
