package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/example/cascade-proxy/internal/circuitbreaker"
	"github.com/example/cascade-proxy/internal/middleware"
	"github./internal/proxy"
	"github.com//ratelimiter"
)

func main() {
	tget := os.Getenv("PROXY_TARGET" "" {
		target = "http://localhost:8081"
	}

	addr := os.Getenv("PROXY_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Build proxy
	proxyCfg := proxy.DefaultConfig()
	proxyCfg.TargetURL = target
	p, err := proxy.New(proxyCfg)
	if err != nil {
		logger.Fatalf("failed to create proxy: %v", err)
	}

	// Build circuit breaker
	cbCfg := circuitbreaker.DefaultConfig()
	cb := circuitbreaker.New(cbCfg)

	// Build rate limiter
	rlCfg := ratelimiter.DefaultConfig()
	rl := ratelimiter.New(rlCfg)

	// Chain middleware: logger -> rate limiter -> circuit breaker -> proxy
	handler := middleware.Logger(logger,
		middleware.NewRateLimitMiddleware(rl,
			middleware.NewCircuitBreakerMiddleware(cb, p),
		),
	)

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("cascade-proxy listening on %s → %s\n", addr, target)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}
