package middleware

import (
	"context"
	"log"
	"net/http"
	"time"
)

// TimeoutConfig holds configuration for the timeout middleware.
type TimeoutConfig struct {
	// Timeout is the maximum duration allowed for a proxied request.
	Timeout time.Duration
	// Logger is used to log timeout events.
	Logger *log.Logger
}

// DefaultTimeoutConfig returns a TimeoutConfig with sensible defaults.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout: 30 * time.Second,
		Logger:  log.Default(),
	}
}

// NewTimeoutMiddleware returns an HTTP middleware that cancels requests
// exceeding the configured timeout duration and responds with 504 Gateway Timeout.
func NewTimeoutMiddleware(cfg TimeoutConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), cfg.Timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan struct{})
			rw := NewResponseRecorder(w)

			go func() {
				defer close(done)
				next.ServeHTTP(rw, r)
			}()

			select {
			case <-done:
				// Request completed in time; flush recorded response.
				w.WriteHeader(rw.Status())
			case <-ctx.Done():
				cfg.Logger.Printf("timeout: request to %s exceeded %s", r.URL.Path, cfg.Timeout)
				http.Error(w, "504 Gateway Timeout", http.StatusGatewayTimeout)
			}
		})
	}
}
