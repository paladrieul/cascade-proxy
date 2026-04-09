package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/cascade-proxy/internal/circuitbreaker"
)

// Config holds proxy configuration.
type Config struct {
	TargetURL      string
	MaxRetries     int
	RetryDelay     time.Duration
	RequestTimeout time.Duration
	CBConfig       circuitbreaker.Config
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxRetries:     3,
		RetryDelay:     200 * time.Millisecond,
		RequestTimeout: 10 * time.Second,
		CBConfig:       circuitbreaker.DefaultConfig(),
	}
}

// Proxy forwards requests to a target URL with retry and circuit breaker support.
type Proxy struct {
	cfg      Config
	target   *url.URL
	cb       *circuitbreaker.CircuitBreaker
	reverseP *httputil.ReverseProxy
	client   *http.Client
}

// New creates a new Proxy instance.
func New(cfg Config) (*Proxy, error) {
	target, err := url.Parse(cfg.TargetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}

	cb := circuitbreaker.New(cfg.CBConfig)
	client := &http.Client{Timeout: cfg.RequestTimeout}

	return &Proxy{
		cfg:    cfg,
		target: target,
		cb:     cb,
		client: client,
	}, nil
}

// ServeHTTP implements http.Handler.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var lastErr error

	for attempt := 0; attempt <= p.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(p.cfg.RetryDelay)
		}

		err := p.cb.Execute(func() error {
			return p.forward(w, r)
		})

		if err == nil {
			return
		}

		lastErr = err
		if err == circuitbreaker.ErrCircuitOpen {
			http.Error(w, "service unavailable (circuit open)", http.StatusServiceUnavailable)
			return
		}
	}

	http.Error(w, fmt.Sprintf("upstream error after %d retries: %v", p.cfg.MaxRetries, lastErr), http.StatusBadGateway)
}

// forward proxies the request to the target and returns an error on non-2xx status.
func (p *Proxy) forward(w http.ResponseWriter, r *http.Request) error {
	outReq, err := buildRequest(r, p.target)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := p.client.Do(outReq)
	if err != nil {
		return fmt.Errorf("upstream request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("upstream returned %d", resp.StatusCode)
	}

	copyResponse(w, resp)
	return nil
}
