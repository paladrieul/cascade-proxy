// Package middleware provides composable HTTP middleware for cascade-proxy.
//
// # Failover Middleware
//
// NewFailoverMiddleware wraps a handler and retries the request against an
// ordered list of backend URLs whenever the upstream returns a configurable
// retryable HTTP status code (default: 502, 503, 504).
//
// Backends are tried in declaration order. The first backend to return a
// non-retryable status code wins, and its response is forwarded to the client.
// If every backend returns a retryable status, the middleware responds with
// 502 Bad Gateway.
//
// Example:
//
//	cfg := middleware.DefaultFailoverConfig()
//	cfg.Backends = []string{
//		"https://primary.example.com",
//		"https://secondary.example.com",
//		"https://tertiary.example.com",
//	}
//	h := middleware.NewFailoverMiddleware(cfg, proxyHandler)
package middleware
