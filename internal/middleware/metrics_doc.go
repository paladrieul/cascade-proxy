// Package middleware provides composable HTTP middleware for cascade-proxy.
//
// # Metrics Middleware
//
// NewMetricsMiddleware instruments an HTTP handler and exposes Prometheus
// metrics on a dedicated /metrics endpoint. The following metrics are
// collected for every request that passes through the middleware:
//
//   - cascade_requests_total          – counter partitioned by method and status code
//   - cascade_request_duration_seconds – histogram of end-to-end latency
//   - cascade_errors_total             – counter incremented for 5xx responses
//
// # Usage
//
//	mux := http.NewServeMux()
//	mux.Handle("/metrics", promhttp.Handler())
//
//	handler := middleware.NewMetricsMiddleware(next)
//
// Metrics are registered once against the default Prometheus registry.
// Calling NewMetricsMiddleware more than once in the same process is safe;
// subsequent calls reuse the already-registered collectors.
package middleware
