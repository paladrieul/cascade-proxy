package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/cascade-proxy/internal/circuitbreaker"
)

func newTestCBMiddleware(t *testing.T) (*CircuitBreakerMiddleware, *circuitbreaker.CircuitBreaker) {
	t.Helper()
	cfg := circuitbreaker.DefaultConfig()
	cfg.FailureThreshold = 2
	cfg.SuccessThreshold = 1
	cfg.Timeout = 50 * time.Millisecond
	cb := circuitbreaker.New(cfg)
	logger := log.New(os.Stdout, "[test-cb] ", 0)
	return NewCircuitBreakerMiddleware(cb, logger), cb
}

func TestCircuitBreakerAllowsRequestWhenClosed(t *testing.T) {
	mw, _ := newTestCBMiddleware(t)

	handler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCircuitBreakerRejects503WhenOpen(t *testing.T) {
	mw, _ := newTestCBMiddleware(t)

	// trigger failures to open the circuit
	failHandler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		failHandler.ServeHTTP(rec, req)
	}

	// now circuit should be open
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	failHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestCircuitBreakerRecoveryAfterTimeout(t *testing.T) {
	mw, _ := newTestCBMiddleware(t)

	failHandler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		failHandler.ServeHTTP(rec, req)
	}

	time.Sleep(60 * time.Millisecond)

	successHandler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	successHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 after recovery, got %d", rec.Code)
	}
}
