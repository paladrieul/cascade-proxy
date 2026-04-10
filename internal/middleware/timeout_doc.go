// Package middleware provides composable HTTP middleware for cascade-proxy.
//
// # Timeout Middleware
//
// NewTimeoutMiddleware wraps a handler and enforces a maximum duration for
// each request. If the upstream handler does not complete within the
// configured deadline the middleware cancels the request context and writes
// a 504 Gateway Timeout response.
//
// # Configuration
//
//	config := middleware.DefaultTimeoutConfig()
//	config.Timeout = 10 * time.Second   // per-request deadline
//	config.Message = "request timed out" // body sent on timeout
//
// # Usage
//
//	handler := middleware.NewTimeoutMiddleware(next, config)
//
// The middleware relies on context cancellation so downstream handlers
// should respect ctx.Done() for early termination.
package middleware
