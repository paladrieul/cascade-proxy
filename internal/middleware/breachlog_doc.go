// Package middleware provides a collection of composable HTTP middleware
// components for the cascade-proxy server.
//
// # BreachLog Middleware
//
// NewBreachLogMiddleware instruments a handler to detect and log when the
// error rate for a given key (e.g. URL path, client IP) exceeds a configured
// threshold over a sliding window of recent responses.
//
// This is useful for early-warning alerting before a circuit breaker opens,
// giving operators visibility into degrading backends.
//
// Configuration:
//
//	cfg := middleware.DefaultBreachLogConfig(logger)
//	cfg.WindowSize = 50          // number of recent requests to consider
//	cfg.ErrorThreshold = 0.4     // log when 40%+ of responses are 5xx
//	cfg.KeyFunc = func(r *http.Request) string {
//	    return r.Header.Get("X-Service-Name") // group by downstream service
//	}
//	mw := middleware.NewBreachLogMiddleware(cfg)
//
// A warning-level log entry is emitted at most once every 5 seconds per key
// to prevent log flooding during sustained outages.
package middleware
