package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newTestCacheMiddleware(cfg CacheConfig, handler http.Handler) http.Handler {
	return NewCacheMiddleware(cfg, handler)
}

func TestCacheReturnsCachedResponseOnSecondGet(t *testing.T) {
	var calls atomic.Int32
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello")) //nolint:errcheck
	})
	h := newTestCacheMiddleware(DefaultCacheConfig(), backend)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/foo", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("expected backend called once, got %d", calls.Load())
	}
}

func TestCacheSetsHitHeader(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := newTestCacheMiddleware(DefaultCacheConfig(), backend)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/bar", nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/bar", nil))
	if rec.Header().Get("X-Cache") != "HIT" {
		t.Fatal("expected X-Cache: HIT on second request")
	}
}

func TestCacheBypassesNonGetMethods(t *testing.T) {
	var calls atomic.Int32
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	h := newTestCacheMiddleware(DefaultCacheConfig(), backend)

	for i := 0; i < 3; i++ {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/baz", nil))
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 backend calls for POST, got %d", calls.Load())
	}
}

func TestCacheDoesNotCacheNon200(t *testing.T) {
	var calls atomic.Int32
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	})
	h := newTestCacheMiddleware(DefaultCacheConfig(), backend)

	for i := 0; i < 2; i++ {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/err", nil))
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 backend calls for non-200, got %d", calls.Load())
	}
}

func TestCacheExpiresTTL(t *testing.T) {
	var calls atomic.Int32
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	cfg := CacheConfig{TTL: 20 * time.Millisecond, MaxEntries: 16}
	h := newTestCacheMiddleware(cfg, backend)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ttl", nil))
	time.Sleep(40 * time.Millisecond)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ttl", nil))
	if calls.Load() != 2 {
		t.Fatalf("expected 2 backend calls after TTL expiry, got %d", calls.Load())
	}
}
