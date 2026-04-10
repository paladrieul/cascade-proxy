package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestDedupeIntegrationWithRequestID verifies that deduplication works
// correctly when chained after the request-ID middleware.
func TestDedupeIntegrationWithRequestID(t *testing.T) {
	calls := int32(0)
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusCreated)
	})

	dedupeCfg := DedupeConfig{TTL: 300 * time.Millisecond, Methods: []string{http.MethodPost}}
	chain := NewRequestIDMiddleware(NewDedupeMiddleware(dedupeCfg, backend))

	send := func(idempotencyKey string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/checkout", nil)
		req.Header.Set("Idempotency-Key", idempotencyKey)
		chain.ServeHTTP(rec, req)
		return rec
	}

	first := send("order-99")
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", first.Code)
	}
	second := send("order-99")
	if second.Header().Get("X-Dedupe-Hit") != "true" {
		t.Error("expected X-Dedupe-Hit on second request")
	}
	if second.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID to be set by request-ID middleware")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 backend call, got %d", calls)
	}
}

// TestDedupeIntegrationWithLogger verifies the middleware chain does not
// interfere with logging when a deduplicated response is returned.
func TestDedupeIntegrationWithLogger(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	dedupeCfg := DedupeConfig{TTL: 300 * time.Millisecond, Methods: []string{http.MethodPut}}

	var logged bool
	logger := Logger(func(status, bytes int, r *http.Request, d time.Duration) {
		logged = true
		if status != http.StatusAccepted {
			t.Errorf("logger saw status %d, want 202", status)
		}
	})

	chain := logger(NewDedupeMiddleware(dedupeCfg, backend))

	send := func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/resource/1", nil)
		req.Header.Set("Idempotency-Key", "put-key-1")
		chain.ServeHTTP(rec, req)
	}
	send()
	send() // should hit dedupe cache

	if !logged {
		t.Error("expected logger to be called")
	}
}
