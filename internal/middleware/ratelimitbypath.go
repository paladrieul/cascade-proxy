package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/cascade-proxy/internal/ratelimiter"
)

// PathRateLimitRule defines a rate limit rule for a specific path prefix.
type PathRateLimitRule struct {
	Prefix   string
	Rate     float64
	Burst    int
	TTL      time.Duration
}

// PathRateLimitConfig holds configuration for per-path rate limiting.
type PathRateLimitConfig struct {
	Rules      []PathRateLimitRule
	Fallback   *ratelimiter.Config // applied to paths not matching any rule
	Logger     *slog.Logger
}

// DefaultPathRateLimitConfig returns a sensible default configuration.
func DefaultPathRateLimitConfig() PathRateLimitConfig {
	return PathRateLimitConfig{
		Logger: slog.Default(),
	}
}

// NewPathRateLimitMiddleware returns an http.Handler that enforces per-path
// rate limits. Rules are evaluated in order; the first matching prefix wins.
func NewPathRateLimitMiddleware(cfg PathRateLimitConfig, next http.Handler) http.Handler {
	type entry struct {
		prefix string
		limiter *ratelimiter.RateLimiter
	}

	var entries []entry
	for _, r := range cfg.Rules {
		rl := ratelimiter.New(ratelimiter.Config{
			Rate:  r.Rate,
			Burst: r.Burst,
			TTL:   r.TTL,
		})
		entries = append(entries, entry{prefix: r.Prefix, limiter: rl})
	}

	var fallback *ratelimiter.RateLimiter
	if cfg.Fallback != nil {
		fallback = ratelimiter.New(*cfg.Fallback)
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		key := clientKey(r)

		var rl *ratelimiter.RateLimiter
		for _, e := range entries {
			if len(path) >= len(e.prefix) && path[:len(e.prefix)] == e.prefix {
				rl = e.limiter
				break
			}
		}
		if rl == nil {
			rl = fallback
		}

		if rl != nil && !rl.Allow(key) {
			logger.Warn("path rate limit exceeded",
				slog.String("path", path),
				slog.String("client", key),
			)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
