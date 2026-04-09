package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware catches panics in downstream handlers, logs the stack
// trace, and returns a 500 Internal Server Error to the client so that a
// single misbehaving request cannot crash the whole proxy process.
func NewRecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					logger.Error("panic recovered",
						"panic", rec,
						"method", r.Method,
						"path", r.URL.Path,
						"stack", string(stack),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
