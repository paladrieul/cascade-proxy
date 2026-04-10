// Package balancer provides a round-robin load balancer for distributing
// requests across multiple backend targets.
package balancer

import (
	"errors"
	"net/url"
	"sync/atomic"
)

// ErrNoTargets is returned when no backend targets are available.
var ErrNoTargets = errors.New("balancer: no targets available")

// Config holds configuration for the load balancer.
type Config struct {
	Targets []string
}

// Balancer distributes requests across a set of backend targets
// using a round-robin strategy.
type Balancer struct {
	targets []*url.URL
	counter uint64
}

// New creates a new Balancer from the given config.
// Returns an error if any target URL is invalid or no targets are provided.
func New(cfg Config) (*Balancer, error) {
	if len(cfg.Targets) == 0 {
		return nil, ErrNoTargets
	}

	urls := make([]*url.URL, 0, len(cfg.Targets))
	for _, t := range cfg.Targets {
		u, err := url.Parse(t)
		if err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}

	return &Balancer{targets: urls}, nil
}

// Next returns the next backend target using round-robin selection.
// Returns ErrNoTargets if the balancer has no targets.
func (b *Balancer) Next() (*url.URL, error) {
	if len(b.targets) == 0 {
		return nil, ErrNoTargets
	}
	idx := atomic.AddUint64(&b.counter, 1) - 1
	return b.targets[idx%uint64(len(b.targets))], nil
}

// Len returns the number of registered targets.
func (b *Balancer) Len() int {
	return len(b.targets)
}
