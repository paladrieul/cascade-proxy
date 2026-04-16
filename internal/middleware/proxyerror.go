package middleware

import (
	"log/slog"
	"net/http"
)

// ProxyErrorConfig holds configuration for the proxy error handler middleware.
type ProxyErrorConfig struct {
	// Logger is the structured logger to use. If nil, slog.Default() is used.
	Logger *slog.Logger
	// IncludeDetails controls whether error details are included in the response body.
	IncludeDetails bool
}

// DefaultProxyErrorConfig returns a ProxyErrorConfig with sensible defaults.
func DefaultProxyErrorConfig() ProxyErrorConfig {
	return ProxyErrorConfig{
		Logger:         slog.Default(),
		IncludeDetails: false,
	}
}

// NewProxyErrorMiddleware returns middleware that intercepts 502/503/504 responses
// from upstream and emits a structured log entry with context about the failure.
func NewProxyErrorMiddleware(cfg ProxyErrorConfig, next http.Handler) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := NewBufferedResponseRecorder()
		next.ServeHTTP(rec, r)

		switch rec.Code {
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			cfg.Logger.Error("proxy upstream error",
				"status", rec.Code,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)
			if cfg.IncludeDetails {
				w.Header().Set("X-Proxy-Error", http.StatusText(rec.Code))
			}
		}

		for k, vs := range rec.HeaderMap {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rec.Code)
		_, _ = w.Write(rec.Body.Bytes())
	})
}
