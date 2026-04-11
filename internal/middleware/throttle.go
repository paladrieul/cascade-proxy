package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// ThrottleConfig holds configuration for the throttle middleware.
type ThrottleConfig struct {
	// MaxConcurrent is the maximum number of requests processed simultaneously.
	MaxConcurrent int
	// QueueTimeout is how long a request will wait in the queue before being rejected.
	QueueTimeout time.Duration
	// Logger is used to log throttle events.
	Logger *log.Logger
}

// DefaultThrottleConfig returns a ThrottleConfig with sensible defaults.
func DefaultThrottleConfig(logger *log.Logger) ThrottleConfig {
	return ThrottleConfig{
		MaxConcurrent: 10,
		QueueTimeout:  5 * time.Second,
		Logger:        logger,
	}
}

type throttleMiddleware struct {
	cfg    ThrottleConfig
	sem    chan struct{}
	mu     sync.Mutex
	active int
}

// NewThrottleMiddleware returns an http.Handler that limits concurrent requests.
// Requests exceeding MaxConcurrent are queued; if they cannot acquire a slot
// within QueueTimeout a 503 Service Unavailable is returned.
func NewThrottleMiddleware(cfg ThrottleConfig, next http.Handler) http.Handler {
	t := &throttleMiddleware{
		cfg: cfg,
		sem: make(chan struct{}, cfg.MaxConcurrent),
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case t.sem <- struct{}{}:
			defer func() { <-t.sem }()
			next.ServeHTTP(w, r)
		case <-time.After(cfg.QueueTimeout):
			if cfg.Logger != nil {
				cfg.Logger.Printf("throttle: request rejected after %s queue timeout", cfg.QueueTimeout)
			}
			http.Error(w, "503 Service Unavailable: too many concurrent requests", http.StatusServiceUnavailable)
		}
	})
}
