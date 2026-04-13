// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Canary Middleware
//
// NewCanaryMiddleware implements traffic splitting between a stable (primary)
handler and a canary backend URL.
//
// Traffic is split by a configurable weight (0–100 %). A weight of 0 sends
// all traffic to the primary; a weight of 100 sends all traffic to the
// canary. Fractional splits use a uniform random sample per request.
//
// An optional HeaderOverride allows individual clients to force canary routing
// by sending a specific request header (e.g. X-Canary: 1). This is useful for
// internal testing or feature-flag driven rollouts.
//
// Requests routed to the canary receive an X-Canary-Routed: true response
// header so that callers can observe which backend served them.
//
// Example:
//
//	cfg := middleware.DefaultCanaryConfig()
//	cfg.CanaryURL = "http://canary-svc:8080"
//	cfg.Weight = 20 // 20 % of traffic
//	h, err := middleware.NewCanaryMiddleware(cfg, primaryHandler)
package middleware
