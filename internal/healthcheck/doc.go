// Package healthcheck provides active liveness probing for upstream backends
// used by the cascade-proxy.
//
// A Checker is initialised with one or more target base URLs and a probe
// timeout. Calling Probe issues concurrent GET /health requests to every
// target and records whether each responded with a non-5xx status code.
//
// The results are exposed via Statuses() for programmatic inspection and via
// Handler(), which returns an http.HandlerFunc suitable for mounting on a
// management port (e.g. /__health). The handler responds with:
//
//	200 OK                  — all backends healthy
//	503 Service Unavailable — one or more backends unhealthy
//
// Response body is always JSON:
//
//	{
//	  "healthy": true,
//	  "backends": [
//	    {"target": "http://svc:8080", "healthy": true, "latency": "3ms"}
//	  ]
//	}
//
// Probe is safe for concurrent use. It is the caller's responsibility to
// schedule periodic probing (e.g. via time.Ticker) if continuous monitoring
// is required.
package healthcheck
