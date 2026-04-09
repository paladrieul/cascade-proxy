package ratelimiter

import (
	"net/http"
	"sync"
	"time"
)

// Config holds configuration for the rate limiter.
type Config struct {
	RequestsPerSecond float64
	Burst             int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		RequestsPerSecond: 100,
		Burst:             20,
	}
}

// bucket tracks token state for a single client key.
type bucket struct {
	tokens    float64
	lastRefil time.Time
	mu        sync.Mutex
}

// RateLimiter implements a per-key token bucket rate limiter.
type RateLimiter struct {
	cfg     Config
	buckets map[string]*bucket
	mu      sync.Mutex
}

// New creates a new RateLimiter with the given config.
func New(cfg Config) *RateLimiter {
	return &RateLimiter{
		cfg:     cfg,
		buckets: make(map[string]*bucket),
	}
}

// Allow returns true if the request for the given key is within the rate limit.
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	b, ok := r.buckets[key]
	if !ok {
		b = &bucket{
			tokens:    float64(r.cfg.Burst),
			lastRefil: time.Now(),
		}
		r.buckets[key] = b
	}
	r.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefil).Seconds()
	b.tokens += elapsed * r.cfg.RequestsPerSecond
	if b.tokens > float64(r.cfg.Burst) {
		b.tokens = float64(r.cfg.Burst)
	}
	b.lastRefil = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Middleware returns an HTTP middleware that enforces rate limiting by remote IP.
func (r *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		key := clientKey(req)
		if !r.Allow(key) {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, req)
	})
}

// clientKey extracts a stable key from the request (remote IP).
func clientKey(req *http.Request) string {
	if fwd := req.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	return req.RemoteAddr
}
