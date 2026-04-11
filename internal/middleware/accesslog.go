package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// AccessLogConfig holds configuration for the access log middleware.
type AccessLogConfig struct {
	// Logger is the structured logger to write access log entries to.
	Logger *slog.Logger
	// SkipPaths is a list of URL paths to exclude from access logging.
	SkipPaths []string
}

// DefaultAccessLogConfig returns an AccessLogConfig with sensible defaults.
func DefaultAccessLogConfig(logger *slog.Logger) AccessLogConfig {
	return AccessLogConfig{
		Logger:    logger,
		SkipPaths: []string{"/healthz", "/readyz"},
	}
}

// NewAccessLogMiddleware returns an HTTP middleware that writes a structured
// access log entry for every request, including method, path, status code,
// latency, and response size. Paths listed in SkipPaths are silently passed
// through without logging.
func NewAccessLogMiddleware(cfg AccessLogConfig) func(http.Handler) http.Handler {
	skip := make(map[string]struct{}, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skip[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ignored := skip[r.URL.Path]; ignored {
				next.ServeHTTP(w, r)
				return
			}

			rec := NewBufferedResponseRecorder(w)
			start := time.Now()

			next.ServeHTTP(rec, r)

			latency := time.Since(start)
			status := rec.Status()

			level := slog.LevelInfo
			if status >= 500 {
				level = slog.LevelError
			} else if status >= 400 {
				level = slog.LevelWarn
			}

			cfg.Logger.Log(r.Context(), level, "access",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"latency", fmt.Sprintf("%dms", latency.Milliseconds()),
				"bytes", rec.BytesWritten(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}
