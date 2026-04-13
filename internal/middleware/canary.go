package middleware

import (
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// CanaryConfig controls traffic splitting between a stable and canary backend.
type CanaryConfig struct {
	// Weight is the percentage of traffic (0–100) routed to the canary backend.
	Weight int
	// CanaryURL is the upstream URL for canary traffic.
	CanaryURL string
	// HeaderOverride, if set, routes requests carrying this header to the canary
	// regardless of weight.
	HeaderOverride string
}

// DefaultCanaryConfig returns a conservative default: 10 % canary traffic.
func DefaultCanaryConfig() CanaryConfig {
	return CanaryConfig{
		Weight:         10,
		HeaderOverride: "X-Canary",
	}
}

// NewCanaryMiddleware splits traffic between the primary handler and a canary
// backend according to cfg. Requests selected for the canary are reverse-
// proxied directly; all other requests fall through to next.
func NewCanaryMiddleware(cfg CanaryConfig, next http.Handler) (http.Handler, error) {
	if cfg.CanaryURL == "" {
		// No canary configured – pass through transparently.
		return next, nil
	}

	target, err := url.Parse(cfg.CanaryURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isCanaryRequest(r, cfg) {
			w.Header().Set("X-Canary-Routed", "true")
			proxy.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}), nil
}

func isCanaryRequest(r *http.Request, cfg CanaryConfig) bool {
	if cfg.HeaderOverride != "" && r.Header.Get(cfg.HeaderOverride) != "" {
		return true
	}
	if cfg.Weight <= 0 {
		return false
	}
	if cfg.Weight >= 100 {
		return true
	}
	//nolint:gosec // non-cryptographic use is intentional
	return rand.Intn(100) < cfg.Weight
}
