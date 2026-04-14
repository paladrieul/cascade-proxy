// Package middleware provides a collection of composable HTTP middleware
// components for cascade-proxy.
//
// # Maintenance Middleware
//
// NewMaintenanceMiddleware enables or disables a global maintenance mode at
// runtime without restarting the proxy. While active, every request that does
// not match an allowed path receives an HTTP 503 response with a JSON body and
// an optional Retry-After header.
//
// The Enabled flag is an *atomic.Bool so it can be toggled safely from a
// separate goroutine — for example, from a management API or a signal handler.
//
// Usage:
//
//	cfg := middleware.DefaultMaintenanceConfig()
//	mw  := middleware.NewMaintenanceMiddleware(cfg)
//
//	// Enable maintenance mode at runtime:
//	cfg.Enabled.Store(true)
//
//	// Disable again:
//	cfg.Enabled.Store(false)
package middleware
