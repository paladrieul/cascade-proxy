package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// UpstreamConfig controls upstream health tracking behaviour.
type UpstreamConfig struct {
	// StatusHeader is the response header written with the upstream latency in ms.
	StatusHeader string
	// SlowThreshold marks a request as slow when its upstream latency exceeds this value.
	SlowThreshold time.Duration
	// Logger is used to emit slow-upstream warnings.
	Logger *slog.Logger
}

// DefaultUpstreamConfig returns sensible defaults.
func DefaultUpstreamConfig() UpstreamConfig {
	return UpstreamConfig{
		StatusHeader:  "X-Upstream-Ms",
		SlowThreshold: 500 * time.Millisecond,
		Logger:        slog.Default(),
	}
}

// NewUpstreamMiddleware records upstream latency, writes it as a response
// header and logs a warning when the configured slow threshold is exceeded.
func NewUpstreamMiddleware(cfg UpstreamConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := NewBufferedResponseRecorder(w)
		next.ServeHTTP(rec, r)
		elapsed := time.Since(start)

		ms := elapsed.Milliseconds()
		w.Header().Set(cfg.StatusHeader, strconv.FormatInt(ms, 10))

		if elapsed > cfg.SlowThreshold {
			cfg.Logger.Warn("slow upstream response",
				"method", r.Method,
				"path", r.URL.Path,
				"latency_ms", ms,
				"threshold_ms", cfg.SlowThreshold.Milliseconds(),
			)
		}

		rec.Flush()
	})
}
