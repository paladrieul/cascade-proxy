package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/casualjim/cascade-proxy/internal/circuitbreaker"
	"github.com/casualjim/cascade-proxy/internal/middleware"
)

func TestCircuitBreakerIntegrationWithRetry(t *testing.T) {
	calls := 0
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	})

	cbCfg := circuitbreaker.DefaultConfig()
	cbCfg.FailureThreshold = 2
	cbCfg.OpenTimeout = 50 * time.Millisecond
	cb := circuitbreaker.New(cbCfg)

	retryCfg := middleware.DefaultRetryConfig()
	retryCfg.MaxAttempts = 2
	retryCfg.RetryableStatuses = []int{http.StatusInternalServerError}

	handler := middleware.NewRetryMiddleware(retryCfg,
		middleware.NewCircuitBreakerMiddleware(cb, backend),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError && rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 500 or 503, got %d", rec.Code)
	}
	if calls == 0 {
		t.Error("expected backend to be called at least once")
	}
}

func TestCircuitBreakerIntegrationOpensAfterFailures(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeaderadGateway)
	})

	cbCfg := circuitbreaker.DefaultConfig()
	cbCfg.FailureThreshold = 3
	cbCfg.OpenTimeout = 100 * time.Millisecond
	cb := circuitbreaker.New(cbCfg)

	handler := middleware.NewCircuitBreakerMiddleware(cb, backend)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// circuit should now be open
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when circuit open, got %d", rec.Code)
	}
}
