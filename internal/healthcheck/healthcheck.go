// Package healthcheck provides a simple health check handler and backend
// liveness probe used by the proxy to determine upstream availability.
package healthcheck

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status represents the health of a backend target.
type Status struct {
	Target  string `json:"target"`
	Healthy bool   `json:"healthy"`
	Latency string `json:"latency,omitempty"`
}

// Checker probes a set of backend targets and reports their health.
type Checker struct {
	mu      sync.RWMutex
	targets []string
	client  *http.Client
	statuses map[string]Status
}

// New creates a Checker for the given backend target URLs.
func New(targets []string, timeout time.Duration) *Checker {
	return &Checker{
		targets:  targets,
		client:   &http.Client{Timeout: timeout},
		statuses: make(map[string]Status, len(targets)),
	}
}

// Probe performs a synchronous health check against all configured targets.
func (c *Checker) Probe(ctx context.Context) {
	var wg sync.WaitGroup
	for _, t := range c.targets {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			start := time.Now()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, target+"/health", nil)
			healthy := false
			var latency time.Duration
			if err == nil {
				resp, doErr := c.client.Do(req)
				latency = time.Since(start)
				if doErr == nil {
					resp.Body.Close()
					healthy = resp.StatusCode < 500
				}
			}
			c.mu.Lock()
			c.statuses[target] = Status{
				Target:  target,
				Healthy: healthy,
				Latency: latency.String(),
			}
			c.mu.Unlock()
		}(t)
	}
	wg.Wait()
}

// Statuses returns a snapshot of the last probed statuses.
func (c *Checker) Statuses() []Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Status, 0, len(c.statuses))
	for _, s := range c.statuses {
		out = append(out, s)
	}
	return out
}

// Handler returns an http.HandlerFunc that reports backend health as JSON.
func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.Probe(r.Context())
		statuses := c.Statuses()
		allHealthy := true
		for _, s := range statuses {
			if !s.Healthy {
				allHealthy = false
				break
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if !allHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"healthy":  allHealthy,
			"backends": statuses,
		})
	}
}
