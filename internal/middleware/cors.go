package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed.
	// Use ["*"] to allow all origins.
	AllowedOrigins []string
	// AllowedMethods is a list of HTTP methods allowed for CORS requests.
	AllowedMethods []string
	// AllowedHeaders is a list of HTTP headers allowed for CORS requests.
	AllowedHeaders []string
	// AllowCredentials indicates whether the request can include credentials.
	AllowCredentials bool
	// MaxAge sets the value of the Access-Control-Max-Age header in seconds.
	MaxAge string
}

// DefaultCORSConfig returns a permissive CORS configuration suitable for development.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Authorization", "Accept"},
		AllowCredentials: false,
		MaxAge:           "86400",
	}
}

// NewCORSMiddleware returns an HTTP middleware that applies CORS headers based on cfg.
func NewCORSMiddleware(cfg CORSConfig) func(http.Handler) http.Handler {
	allowedOriginSet := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		allowedOriginSet[o] = struct{}{}
	}

	allowedMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowedHeaders := strings.Join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			_, allowAll := allowedOriginSet["*"]
			_, allowOrigin := allowedOriginSet[origin]

			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if allowOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests.
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
				if cfg.MaxAge != "" {
					w.Header().Set("Access-Control-Max-Age", cfg.MaxAge)
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
