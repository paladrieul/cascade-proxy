// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Throttle Middleware
//
// NewThrottleMiddleware limits the number of requests handled concurrently by
// the proxy. Requests that arrive when the concurrency ceiling has been reached
// are held in a lightweight channel-based queue. If a queued request cannot
// acquire a processing slot within the configured QueueTimeout it is rejected
// with HTTP 503 Service Unavailable.
//
// # Configuration
//
//	cfg := middleware.ThrottleConfig{
//	    MaxConcurrent: 20,              // allow up to 20 in-flight requests
//	    QueueTimeout:  3 * time.Second, // wait at most 3 s before rejecting
//	    Logger:        myLogger,
//	}
//	handler = middleware.NewThrottleMiddleware(cfg, handler)
//
// Use DefaultThrottleConfig for sensible out-of-the-box settings
// (MaxConcurrent=10, QueueTimeout=5s).
package middleware
