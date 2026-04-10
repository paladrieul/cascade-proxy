// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Deduplication Middleware
//
// NewDedupeMiddleware prevents duplicate mutating requests from reaching the
// backend within a configurable time window.  It is designed to work with
// idempotency keys sent by clients (Idempotency-Key header) so that retried
// POST / PUT / PATCH requests do not cause duplicate side-effects.
//
// When a request fingerprint is seen for the first time the response is
// forwarded normally and the result is stored.  Subsequent requests with the
// same fingerprint that arrive before the TTL expires receive the cached
// response immediately together with the X-Dedupe-Hit: true header.
//
// Fingerprinting combines the HTTP method, request URL and, when present, the
// Idempotency-Key header.  Only methods listed in DedupeConfig.Methods are
// subject to deduplication; all other methods pass through unchanged.
//
// Example:
//
//	cfg := middleware.DefaultDedupeConfig()
//	cfg.TTL = time.Second
//	handler = middleware.NewDedupeMiddleware(cfg, handler)
package middleware
