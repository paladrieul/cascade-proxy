package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// MaintenanceConfig controls the maintenance mode middleware behaviour.
type MaintenanceConfig struct {
	// Enabled toggles maintenance mode on or off atomically.
	Enabled *atomic.Bool
	// StatusCode is the HTTP status returned during maintenance (default 503).
	StatusCode int
	// Message is the body returned to clients.
	Message string
	// AllowedPaths are paths that bypass maintenance mode (e.g. health checks).
	AllowedPaths []string
	// RetryAfter sets the Retry-After header value in seconds (0 = omit).
	RetryAfter int
}

// DefaultMaintenanceConfig returns a sensible default configuration.
func DefaultMaintenanceConfig() MaintenanceConfig {
	return MaintenanceConfig{
		Enabled:      &atomic.Bool{},
		StatusCode:   http.StatusServiceUnavailable,
		Message:      "Service is temporarily unavailable. Please try again later.",
		AllowedPaths: []string{"/healthz", "/readyz"},
		RetryAfter:   30,
	}
}

// NewMaintenanceMiddleware returns an HTTP middleware that rejects all requests
// with a 503 while maintenance mode is active. Allowed paths always pass
// through so that health-check endpoints remain reachable.
func NewMaintenanceMiddleware(cfg MaintenanceConfig) func(http.Handler) http.Handler {
	if cfg.Enabled == nil {
		cfg.Enabled = &atomic.Bool{}
	}
	if cfg.StatusCode == 0 {
		cfg.StatusCode = http.StatusServiceUnavailable
	}

	allowed := make(map[string]struct{}, len(cfg.AllowedPaths))
	for _, p := range cfg.AllowedPaths {
		allowed[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled.Load() {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := allowed[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if cfg.RetryAfter > 0 {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", cfg.RetryAfter))
			}
			w.WriteHeader(cfg.StatusCode)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "maintenance",
				"message": cfg.Message,
				"time":    time.Now().UTC().Format(time.RFC3339),
			})
		})
	}
}
