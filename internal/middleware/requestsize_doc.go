// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # RequestSize Middleware
//
// NewRequestSizeMiddleware enforces configurable limits on incoming request
// metadata to protect upstream services from oversized inputs.
//
// Limits enforced:
//   - MaxURLLength:   rejects requests whose raw URI exceeds the limit (HTTP 414).
//   - MaxQueryParams: rejects requests with more query parameters than allowed (HTTP 400).
//   - MaxHeaderBytes: rejects requests whose combined header size exceeds the limit (HTTP 431).
//
// Example usage:
//
//	cfg := middleware.DefaultRequestSizeConfig()
//	cfg.MaxURLLength = 1024
//	cfg.MaxQueryParams = 20
//	handler := middleware.NewRequestSizeMiddleware(cfg, next)
//
// All limits can be disabled individually by setting the corresponding field to 0.
package middleware
