package middleware

import (
	"log"
	"net/http"

	"github.com/cascade-proxy/internal/circuitbreaker"
)

// CircuitBreakerMiddleware wraps an HTTP handler with circuit breaker protection.
type CircuitBreakerMiddleware struct {
	cb     *circuitbreaker.CircuitBreaker
	logger *log.Logger
}

// NewCircuitBreakerMiddleware creates a new circuit breaker middleware using
// the provided CircuitBreaker instance and logger.
func NewCircuitBreakerMiddleware(cb *circuitbreaker.CircuitBreaker, logger *log.Logger) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		cb:     cb,
		logger: logger,
	}
}

// Wrap returns an http.Handler that gates requests through the circuit breaker.
// When the circuit is open, it responds immediately with 503 Service Unavailable.
func (m *CircuitBreakerMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := m.cb.Execute(func() error {
			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)
			if rec.Status() >= 500 {
				return &upstreamError{status: rec.Status()}
			}
			return nil
		})

		if err != nil {
			if isOpen(err) {
				m.logger.Printf("circuit breaker open: rejecting request %s %s", r.Method, r.URL.Path)
				http.Error(w, "service unavailable: circuit breaker open", http.StatusServiceUnavailable)
				return
			}
			// upstream error already written by the recorder
		}
	})
}

// upstreamError represents a non-2xx response from the upstream handler.
type upstreamError struct {
	status int
}

func (e *upstreamError) Error() string {
	return http.StatusText(e.status)
}

// isOpen returns true when the error signals the circuit breaker is open.
func isOpen(err error) bool {
	return err == circuitbreaker.ErrCircuitOpen
}
