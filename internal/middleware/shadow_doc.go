// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Shadow Middleware
//
// NewShadowMiddleware mirrors every incoming request to a secondary "shadow"
// backend asynchronously, without influencing the response returned to the
// client. This is useful for:
//
//   - Dark-launching new service versions alongside the production backend.
//   - Comparing response parity between old and new implementations.
//   - Load-testing shadow environments with real production traffic shapes.
//
// The shadow request is fired in a background goroutine immediately after the
// primary handler begins executing. The client always receives the primary
// response; shadow failures are logged but never propagated.
//
// # Configuration
//
//	cfg := middleware.DefaultShadowConfig("http://shadow-svc:8080", logger)
//	cfg.Timeout = 1 * time.Second   // cap shadow latency
//
//	handler := middleware.NewShadowMiddleware(cfg, primaryHandler)
//
// The X-Shadow-Request: true header is injected into every mirrored request so
// shadow backends can distinguish mirrored traffic from real traffic.
package middleware
