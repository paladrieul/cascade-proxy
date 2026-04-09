// Package ratelimiter provides a per-key token bucket rate limiter for use
// in HTTP proxy pipelines.
//
// Each unique client key (derived from X-Forwarded-For or RemoteAddr) gets
// its own token bucket. Tokens are refilled continuously at the configured
// rate up to the burst cap.
//
// Basic usage:
//
//	rl := ratelimiter.New(ratelimiter.DefaultConfig())
//
//	// As standalone check:
//	if !rl.Allow(clientIP) {
//	    http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
//	    return
//	}
//
//	// As middleware:
//	http.Handle("/", rl.Middleware(myHandler))
//
// The RateLimiter is safe for concurrent use.
package ratelimiter
