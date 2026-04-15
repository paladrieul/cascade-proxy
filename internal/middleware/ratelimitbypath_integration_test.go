package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cascade-proxy/internal/middleware"
)

func TestPathRateLimitIntegrationWithLogger(t *testing.T) {
	// Compose PathRateLimit → Logger → handler
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	logged := middleware.Logger(inner)

	cfg := middleware.DefaultPathRateLimitConfig()
	cfg.Rules = []middleware.PathRateLimitRule{
		{Prefix: "/limited", Rate: 0.001, Burst: 1, TTL: time.Minute},
	}
	h := middleware.NewPathRateLimitMiddleware(cfg, logged)

	req := func() *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/limited/resource", nil)
		r.RemoteAddr = "7.7.7.7:1"
		return r
	}

	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, req())
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req())
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", rec2.Code)
	}
}

func TestPathRateLimitIntegrationWithRequestID(t *testing.T) {
	// Compose RequestID → PathRateLimit → handler so each request has an ID.
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := middleware.DefaultPathRateLimitConfig()
	cfg.Rules = []middleware.PathRateLimitRule{
		{Prefix: "/secure", Rate: 0.001, Burst: 2, TTL: time.Minute},
	}
	rateLimited := middleware.NewPathRateLimitMiddleware(cfg, inner)
	h := middleware.NewRequestIDMiddleware(rateLimited)

	send := func() int {
		req := httptest.NewRequest(http.MethodGet, "/secure/data", nil)
		req.RemoteAddr = "8.8.8.8:1"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	for i := 0; i < 2; i++ {
		if code := send(); code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, code)
		}
	}

	if code := send(); code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after burst exhausted, got %d", code)
	}
}
