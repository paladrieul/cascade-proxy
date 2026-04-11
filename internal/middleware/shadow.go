package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"
)

// ShadowConfig holds configuration for the shadow proxy middleware.
type ShadowConfig struct {
	// ShadowURL is the base URL of the shadow backend to mirror traffic to.
	ShadowURL string
	// Logger is used to record shadow request outcomes.
	Logger *slog.Logger
	// Timeout is the maximum duration to wait for the shadow backend.
	Timeout time.Duration
}

// DefaultShadowConfig returns a ShadowConfig with sensible defaults.
func DefaultShadowConfig(shadowURL string, logger *slog.Logger) ShadowConfig {
	return ShadowConfig{
		ShadowURL: shadowURL,
		Logger:    logger,
		Timeout:   2 * time.Second,
	}
}

// NewShadowMiddleware mirrors every incoming request to a shadow backend
// asynchronously without affecting the primary response.
func NewShadowMiddleware(cfg ShadowConfig, next http.Handler) http.Handler {
	client := &http.Client{Timeout: cfg.Timeout}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Buffer the request body so both primary and shadow can read it.
		var bodyBytes []byte
		if r.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// Fire shadow request in the background.
		go func() {
			shadowReq, err := http.NewRequest(r.Method, cfg.ShadowURL+r.RequestURI, bytes.NewReader(bodyBytes))
			if err != nil {
				cfg.Logger.Warn("shadow: failed to build request", "error", err)
				return
			}
			for key, vals := range r.Header {
				for _, v := range vals {
					shadowReq.Header.Add(key, v)
				}
			}
			shadowReq.Header.Set("X-Shadow-Request", "true")

			rec := httptest.NewRecorder()
			_ = rec

			start := time.Now()
			resp, err := client.Do(shadowReq)
			latency := time.Since(start)
			if err != nil {
				cfg.Logger.Warn("shadow: request failed", "error", err, "latency", latency)
				return
			}
			defer resp.Body.Close()
			cfg.Logger.Info("shadow: request completed",
				"status", resp.StatusCode,
				"latency", latency,
				"path", r.URL.Path,
			)
		}()

		next.ServeHTTP(w, r)
	})
}
