// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Retry Middleware
//
// NewRetryMiddleware wraps an HTTP handler and automatically retries requests
// that receive a retryable HTTP status code from the downstream handler.
//
// Configuration:
//
//	type RetryConfig struct {
//		MaxAttempts       int           // total attempts including the first (default: 3)
//		RetryDelay        time.Duration // delay between attempts (default: 100ms)
//		RetryableStatuses []int         // statuses that trigger a retry (default: 502, 503, 504)
//	}
//
// The middleware buffers the response from the inner handler so that it can
// inspect the status code before committing the response to the client.
// Only GET and HEAD requests are retried by default; all methods are retried
// if the inner handler returns a retryable status regardless of method.
//
// Usage:
//
//	cfg := middleware.DefaultRetryConfig()
//	cfg.MaxAttempts = 5
//	handler := middleware.NewRetryMiddleware(cfg, next)
package middleware
