// Package middleware provides a collection of composable HTTP middleware
// components for the cascade-proxy server.
//
// # IP Filter
//
// NewIPFilterMiddleware enforces IP-based access control on incoming requests.
// It supports both an allowlist (AllowedCIDRs) and a denylist (BlockedCIDRs)
// expressed as CIDR notation strings.
//
// Evaluation order:
//  1. If the client IP matches any entry in BlockedCIDRs, the request is
//     immediately rejected with 403 Forbidden.
//  2. If AllowedCIDRs is non-empty and the client IP does not match any
//     entry, the request is rejected with 403 Forbidden.
//  3. Otherwise the request is forwarded to the next handler.
//
// The client IP is resolved from the X-Forwarded-For header (first value)
// when present, falling back to the TCP RemoteAddr of the connection.
//
// Example:
//
//	cfg := middleware.DefaultIPFilterConfig()
//	cfg.AllowedCIDRs = []string{"10.0.0.0/8", "192.168.0.0/16"}
//	cfg.BlockedCIDRs = []string{"10.99.0.0/16"}
//	handler = middleware.NewIPFilterMiddleware(cfg, handler)
package middleware
