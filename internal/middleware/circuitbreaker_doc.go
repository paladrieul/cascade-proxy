// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Circuit Breaker Middleware
//
// NewCircuitBreakerMiddleware wraps an HTTP handler with a circuit breaker
// that monitors downstream failures and short-circuits requests when the
// failure rate exceeds a configured threshold.
//
// States:
//
//   - Closed: requests pass through normally; failures are counted.
//   - Open: requests are immediately rejected with 503 Service Unavailable;
//     the circuit reopens after a configurable timeout.
//   - Half-Open: a single probe request is allowed through; success closes
//     the circuit, failure reopens it.
//
// Usage:
//
//	cb := circuitbreaker.New(circuitbreaker.DefaultConfig())
//	handler := middleware.NewCircuitBreakerMiddleware(cb, next)
//
// The middleware records a failure for any response with status >= 500 and
// a success for any response with status < 500.
package middleware
