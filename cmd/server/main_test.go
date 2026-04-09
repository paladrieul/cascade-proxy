package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/cascade-proxy/internal/circuitbreaker"
	"github.com/example/cascade-proxy/internal/middleware"
	"github.com/example/cascade-proxy/internal/proxy"
	"github.com/example/cascade-proxy/internal/ratelimiter"
	"log"
	"os"
)

func buildHandler(targetURL string) (http.Handler, error) {
	logger := log.New(os.Stdout, "", 0)

	proxyCfg := proxy.DefaultConfig()
	proxyCfg.TargetURL = targetURL
	p, err := proxy.New(proxyCfg)
	if err != nil {
		return nil, err
	}

	cb := circuitbreaker.New(circuitbreaker.DefaultConfig())
	rl := ratelimiter.New(ratelimiter.DefaultConfig())

	return middleware.Logger(logger,
		middleware.NewRateLimitMiddleware(rl,
			middleware.NewCircuitBreakerMiddleware(cb, p),
		),
	), nil
}

func TestFullStackForwardsRequest(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer backend.Close()

	handler, err := buildHandler(backend.URL)
	if err != nil {
		t.Fatalf("buildHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFullStackReturns502WhenBackendDown(t *testing.T) {
	handler, err := buildHandler("http://127.0.0.1:19999")
	if err != nil {
		t.Fatalf("buildHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}
