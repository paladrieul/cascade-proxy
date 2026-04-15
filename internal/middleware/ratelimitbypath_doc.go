// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Per-Path Rate Limiting
//
// NewPathRateLimitMiddleware enforces independent token-bucket rate limits
// for different URL path prefixes. This is useful when certain API routes
// (e.g. /api/admin) require stricter throttling than general endpoints.
//
// Rules are evaluated in declaration order; the first matching prefix wins.
// An optional Fallback config is applied to requests that do not match any
// rule. If neither a rule nor a fallback matches, the request is allowed.
//
// Example:
//
//	cfg := middleware.DefaultPathRateLimitConfig()
//	cfg.Rules = []middleware.PathRateLimitRule{
//		{Prefix: "/api/admin", Rate: 1, Burst: 5,  TTL: time.Minute},
//		{Prefix: "/api",       Rate: 50, Burst: 100, TTL: time.Minute},
//	}
//	h := middleware.NewPathRateLimitMiddleware(cfg, nextHandler)
package middleware
