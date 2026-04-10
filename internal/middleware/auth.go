package middleware

import (
	"log"
	"net/http"
	"strings"
)

// AuthConfig holds configuration for the API key auth middleware.
type AuthConfig struct {
	// ValidKeys is the set of accepted bearer tokens / API keys.
	ValidKeys map[string]struct{}
	// Header is the HTTP header to inspect (default: Authorization).
	Header string
	// Logger receives rejection log lines.
	Logger *log.Logger
}

// DefaultAuthConfig returns an AuthConfig with sensible defaults.
func DefaultAuthConfig(keys []string, logger *log.Logger) AuthConfig {
	set := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		set[k] = struct{}{}
	}
	return AuthConfig{
		ValidKeys: set,
		Header:    "Authorization",
		Logger:    logger,
	}
}

// NewAuthMiddleware returns an HTTP middleware that rejects requests whose
// Authorization header does not carry a recognised Bearer token.
// Requests without a matching key receive 401 Unauthorized.
func NewAuthMiddleware(cfg AuthConfig, next http.Handler) http.Handler {
	header := cfg.Header
	if header == "" {
		header = "Authorization"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get(header)
		key := extractBearer(raw)

		if _, ok := cfg.ValidKeys[key]; !ok {
			if cfg.Logger != nil {
				cfg.Logger.Printf("auth: rejected request from %s – invalid or missing key", r.RemoteAddr)
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractBearer strips the "Bearer " prefix from a raw Authorization value.
// If the value does not start with "Bearer ", it is returned as-is so that
// plain API-key headers are also supported.
func extractBearer(raw string) string {
	if strings.HasPrefix(raw, "Bearer ") {
		return strings.TrimPrefix(raw, "Bearer ")
	}
	return raw
}
