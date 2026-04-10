package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newTestDedupeMiddleware(cfg DedupeConfig, next http.Handler) http.Handler {
	return NewDedupeMiddleware(cfg, next)
}

func TestDedupePassesThroughNonMatchingMethod(t *testing.T) {
	calls := int32(0)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})
	cfg := DefaultDedupeConfig()
	mw := newTestDedupeMiddleware(cfg, handler)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/resource", nil)
		mw.ServeHTTP(rec, req)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 calls for GET, got %d", calls)
	}
}

func TestDedupeReturnsCachedResponseOnDuplicate(t *testing.T) {
	calls := int32(0)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("X-Backend", "hit")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created")) //nolint:errcheck
	})
	cfg := DedupeConfig{TTL: 200 * time.Millisecond, Methods: []string{http.MethodPost}}
	mw := newTestDedupeMiddleware(cfg, handler)

	send := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orders", nil)
		req.Header.Set("Idempotency-Key", "key-abc")
		mw.ServeHTTP(rec, req)
		return rec
	}

	first := send()
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", first.Code)
	}
	second := send()
	if second.Code != http.StatusCreated {
		t.Fatalf("expected 201 from cache, got %d", second.Code)
	}
	if second.Header().Get("X-Dedupe-Hit") != "true" {
		t.Error("expected X-Dedupe-Hit header on cached response")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected backend called once, got %d", calls)
	}
}

func TestDedupeExpiresCacheAfterTTL(t *testing.T) {
	calls := int32(0)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})
	cfg := DedupeConfig{TTL: 50 * time.Millisecond, Methods: []string{http.MethodPost}}
	mw := newTestDedupeMiddleware(cfg, handler)

	send := func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/items", nil)
		req.Header.Set("Idempotency-Key", "key-ttl")
		mw.ServeHTTP(rec, req)
	}
	send()
	time.Sleep(80 * time.Millisecond)
	send()

	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 backend calls after TTL expiry, got %d", calls)
	}
}

func TestDedupeDifferentKeysAreIndependent(t *testing.T) {
	calls := int32(0)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})
	cfg := DedupeConfig{TTL: 200 * time.Millisecond, Methods: []string{http.MethodPost}}
	mw := newTestDedupeMiddleware(cfg, handler)

	for _, key := range []string{"key-1", "key-2", "key-3"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/items", nil)
		req.Header.Set("Idempotency-Key", key)
		mw.ServeHTTP(rec, req)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 independent backend calls, got %d", calls)
	}
}
