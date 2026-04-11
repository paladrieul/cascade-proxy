package middleware

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// DefaultFailoverConfig returns a FailoverConfig with sensible defaults.
func DefaultFailoverConfig() FailoverConfig {
	return FailoverConfig{
		RetryableStatuses: []int{http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout},
		Timeout:           5 * time.Second,
	}
}

// FailoverConfig controls which backends are tried and under what conditions.
type FailoverConfig struct {
	// Backends is an ordered list of fallback target base URLs.
	// The first backend is primary; subsequent entries are tried on failure.
	Backends []string
	// RetryableStatuses lists HTTP status codes that trigger failover to the next backend.
	RetryableStatuses []int
	// Timeout is the per-backend request deadline.
	Timeout time.Duration
	// Logger is an optional structured logger.
	Logger *slog.Logger
}

// NewFailoverMiddleware returns an http.Handler that forwards requests to the
// primary backend and, on a retryable status, cascades through the fallback
// backends in order. The first successful response is returned to the client.
func NewFailoverMiddleware(cfg FailoverConfig, next http.Handler) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	retryable := make(map[int]bool, len(cfg.RetryableStatuses))
	for _, s := range cfg.RetryableStatuses {
		retryable[s] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i, backend := range cfg.Backends {
			rec := newBufferedRecorder()

			// Rewrite the request target for this backend.
			base, err := url.Parse(backend)
			if err != nil {
				cfg.Logger.Error("failover: invalid backend URL", "backend", backend, "err", err)
				continue
			}

			proxied := r.Clone(r.Context())
			proxied.URL.Scheme = base.Scheme
			proxied.URL.Host = base.Host
			proxied.Host = base.Host

			next.ServeHTTP(rec, proxied)

			if !retryable[rec.status] {
				// Success — flush and return.
				if i > 0 {
					cfg.Logger.Info("failover: succeeded on fallback", "backend", backend, "attempt", i+1)
				}
				rec.flush(w)
				return
			}

			cfg.Logger.Warn("failover: retryable status, trying next backend",
				"backend", backend, "status", rec.status, "attempt", i+1)
		}

		// All backends exhausted.
		cfg.Logger.Error("failover: all backends exhausted")
		http.Error(w, "all backends unavailable", http.StatusBadGateway)
	})
}
