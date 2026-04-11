package middleware

import (
	"net/http"
)

// DefaultBodySizeConfig returns a BodySizeConfig with a 1 MB limit.
func DefaultBodySizeConfig() BodySizeConfig {
	return BodySizeConfig{
		MaxBytes: 1 << 20, // 1 MB
	}
}

// BodySizeConfig holds configuration for the body size limiting middleware.
type BodySizeConfig struct {
	// MaxBytes is the maximum number of bytes allowed in a request body.
	// Requests exceeding this limit will receive a 413 response.
	MaxBytes int64
}

// NewBodySizeMiddleware returns a middleware that rejects requests whose body
// exceeds MaxBytes. It uses http.MaxBytesReader so the limit is enforced
// during reading, protecting downstream handlers from unbounded reads.
func NewBodySizeMiddleware(cfg BodySizeConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil && r.ContentLength > cfg.MaxBytes {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
				return
			}

			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBytes)
			}

			next.ServeHTTP(w, r)
		})
	}
}
