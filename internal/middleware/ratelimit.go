package middleware

import (
	"log"
	"net/http"

	"github.com/cascade-proxy/internal/ratelimiter"
)

// RateLimitMiddleware wraps a RateLimiter and integrates it as HTTP middleware
// with optional logging when a request is rejected.
type RateLimitMiddleware struct {
	limiter *ratelimiter.RateLimiter
	logger  *log.Logger
}

// NewRateLimitMiddleware creates a RateLimitMiddleware using the provided
// RateLimiter and logger.
func NewRateLimitMiddleware(rl *ratelimiter.RateLimiter, logger *log.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: rl,
		logger:  logger,
	}
}

// Handler returns an http.Handler that enforces rate limits and logs rejections.
func (m *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientKey(r)
		if !m.limiter.Allow(key) {
			if m.logger != nil {
				m.logger.Printf("rate limit exceeded for %s %s from %s", r.Method, r.URL.Path, key)
			}
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientKey mirrors the key extraction logic from the ratelimiter package.
func clientKey(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	return r.RemoteAddr
}
