// Package middleware provides a collection of composable HTTP middleware
// components for the cascade-proxy server.
//
// # Access Log Middleware
//
// NewAccessLogMiddleware writes a structured (JSON) access log entry for every
// HTTP request processed by the proxy. Each entry captures:
//
//   - HTTP method and request path
//   - Response status code
//   - Latency in milliseconds
//   - Response body size in bytes
//   - Client remote address
//   - User-Agent header
//
// Log severity is adjusted automatically: 2xx/3xx responses are logged at INFO,
// 4xx responses at WARN, and 5xx responses at ERROR, making it straightforward
// to filter noise in production log aggregators.
//
// Paths that should never appear in access logs (e.g. health-check endpoints)
// can be listed in AccessLogConfig.SkipPaths. The default configuration skips
// /healthz and /readyz.
//
// Usage:
//
//	logger := slog.Default()
//	cfg := middleware.DefaultAccessLogConfig(logger)
//	cfg.SkipPaths = append(cfg.SkipPaths, "/metrics")
//	mw := middleware.NewAccessLogMiddleware(cfg)
//	http.Handle("/", mw(myHandler))
package middleware
