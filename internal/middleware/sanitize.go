package middleware

import (
	"net/http"
	"strings"
)

// SanitizeConfig holds configuration for the sanitize middleware.
type SanitizeConfig struct {
	// StripHeaders is a list of request header names to remove before forwarding.
	StripHeaders []string
	// MaxQueryParams limits the number of query parameters allowed (0 = unlimited).
	MaxQueryParams int
	// AllowedMethods restricts which HTTP methods are permitted (empty = all allowed).
	AllowedMethods []string
}

// DefaultSanitizeConfig returns a SanitizeConfig with sensible defaults.
func DefaultSanitizeConfig() SanitizeConfig {
	return SanitizeConfig{
		StripHeaders:   []string{"X-Forwarded-For", "X-Real-IP"},
		MaxQueryParams: 50,
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions},
	}
}

// NewSanitizeMiddleware returns middleware that strips unwanted headers,
// enforces method allowlists, and limits query parameter count.
func NewSanitizeMiddleware(cfg SanitizeConfig, next http.Handler) http.Handler {
	stripSet := make(map[string]struct{}, len(cfg.StripHeaders))
	for _, h := range cfg.StripHeaders {
		stripSet[http.CanonicalHeaderKey(h)] = struct{}{}
	}

	methodSet := make(map[string]struct{}, len(cfg.AllowedMethods))
	for _, m := range cfg.AllowedMethods {
		methodSet[strings.ToUpper(m)] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Enforce allowed methods.
		if len(methodSet) > 0 {
			if _, ok := methodSet[r.Method]; !ok {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
		}

		// Enforce max query parameter count.
		if cfg.MaxQueryParams > 0 {
			if len(r.URL.Query()) > cfg.MaxQueryParams {
				http.Error(w, "Too Many Query Parameters", http.StatusBadRequest)
				return
			}
		}

		// Strip disallowed request headers.
		for h := range stripSet {
			r.Header.Del(h)
		}

		next.ServeHTTP(w, r)
	})
}
