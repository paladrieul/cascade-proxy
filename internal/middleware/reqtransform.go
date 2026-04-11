package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// RequestTransformRule defines a single find-and-replace rule applied to request bodies.
type RequestTransformRule struct {
	Find        string
	Replace     string
	ContentType string // optional; empty means match all
}

// RequestTransformConfig holds configuration for NewRequestTransformMiddleware.
type RequestTransformConfig struct {
	Rules []RequestTransformRule
}

// DefaultRequestTransformConfig returns an empty configuration.
func DefaultRequestTransformConfig() RequestTransformConfig {
	return RequestTransformConfig{}
}

// NewRequestTransformMiddleware rewrites request bodies according to the
// configured find-and-replace rules before forwarding to the next handler.
func NewRequestTransformMiddleware(cfg RequestTransformConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(cfg.Rules) == 0 || r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}

		ct := r.Header.Get("Content-Type")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		modified := string(body)
		for _, rule := range cfg.Rules {
			if rule.ContentType != "" && !strings.Contains(ct, rule.ContentType) {
				continue
			}
			modified = strings.ReplaceAll(modified, rule.Find, rule.Replace)
		}

		newBody := []byte(modified)
		r = r.Clone(r.Context())
		r.Body = io.NopCloser(bytes.NewReader(newBody))
		r.ContentLength = int64(len(newBody))

		next.ServeHTTP(w, r)
	})
}
