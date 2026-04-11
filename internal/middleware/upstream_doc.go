// Package middleware provides composable HTTP middleware for cascade-proxy.
//
// # Upstream Latency Middleware
//
// NewUpstreamMiddleware wraps a handler and measures the time taken by the
// upstream (next) handler to produce a response.  It writes the elapsed time
// in milliseconds to a configurable response header (default: X-Upstream-Ms)
// so that clients and monitoring systems can observe backend latency without
// needing access to server-side metrics.
//
// When the latency exceeds SlowThreshold a structured warning is emitted via
// the configured slog.Logger, making it easy to spot degraded backends in log
// aggregation pipelines.
//
// Usage:
//
//	cfg := middleware.DefaultUpstreamConfig()
//	cfg.SlowThreshold = 200 * time.Millisecond
//	handler := middleware.NewUpstreamMiddleware(cfg, proxy)
package middleware
