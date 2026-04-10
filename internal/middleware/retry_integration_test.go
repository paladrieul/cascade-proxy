package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/casualjim/cascade-proxy/internal/middleware"
)

func TestRetryIntegrationWithTimeout(t *testing.T) {
	attempts := 0
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadGateway)
	})

	toCfg := middleware.DefaultTimeoutConfig()
	toCfg.Timeout = 200 * time.Millisecond

	retryCfg := middleware.DefaultRetryConfig()
	retryCfg.MaxAttempts = 3
	retryCfg.RetryDelay = 10 * time.Millisecond
	retryCfg.RetryableStatuses = []int{http.StatusBadGateway}

	handler := middleware.NewRetryMiddleware(retryCfg,
		middleware.NewTimeoutMiddleware(toCfg, backend),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502 after all retries, got %d", rec.Code)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryIntegrationSucceedsOnThirdAttempt(t *testing.T) {
	attempts := 0
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	retryCfg := middleware.DefaultRetryConfig()
	retryCfg.MaxAttempts = 3
	retryCfg.RetryDelay = 5 * time.Millisecond
	retryCfg.RetryableStatuses = []int{http.StatusServiceUnavailable}

	handler := middleware.NewRetryMiddleware(retryCfg, backend)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 on third attempt, got %d", rec.Code)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
