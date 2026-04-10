package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/user/cascade-proxy/internal/balancer"
	"github.com/user/cascade-proxy/internal/circuitbreaker"
	"github.com/user/cascade-proxy/internal/healthcheck"
	"github.com/user/cascade-proxy/internal/middleware"
	"github.com/user/cascade-proxy/internal/proxy"
	"github.com/user/cascade-proxy/internal/ratelimiter"
)

func main() {
	logger := log.New(os.Stdout, "[cascade-proxy] ", log.LstdFlags)

	targetURLs := strings.Split(envOr("BACKENDS", "http://localhost:9090"), ",")

	bal, err := balancer.New(targetURLs)
	if err != nil {
		logger.Fatalf("balancer: %v", err)
	}

	p, err := proxy.New(proxy.DefaultConfig)
	if err != nil {
		logger.Fatalf("proxy: %v", err)
	}

	cb := circuitbreaker.New(circuitbreaker.DefaultConfig)
	rl := ratelimiter.New(ratelimiter.DefaultConfig)

	hc := healthcheck.New(targetURLs, healthcheck.Config{
		Interval: 15 * time.Second,
		Timeout:  3 * time.Second,
	})
	go hc.Start()

	var h http.Handler = p
	h = middleware.NewBalancerMiddleware(bal, h)
	h = middleware.NewRetryMiddleware(middleware.DefaultRetryConfig, h)
	h = middleware.NewCircuitBreakerMiddleware(cb, h)
	h = middleware.NewRateLimitMiddleware(rl, h)
	h = middleware.NewCacheMiddleware(middleware.DefaultCacheConfig, h)
	h = middleware.NewCORSMiddleware(middleware.DefaultCORSConfig, h)
	h = middleware.NewAuthMiddleware(middleware.DefaultAuthConfig, h)
	h = middleware.NewCompressMiddleware(middleware.DefaultCompressConfig, h)
	h = middleware.NewTimeoutMiddleware(middleware.DefaultTimeoutConfig, h)
	h = middleware.NewRecoveryMiddleware(logger, h)
	h = middleware.NewMetricsMiddleware(h)
	h = middleware.Logger(logger, h)

	mux := http.NewServeMux()
	mux.Handle("/health", hc.Handler())
	mux.Handle("/", h)

	addr := envOr("ADDR", ":8080")
	logger.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Fatalf("server: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
