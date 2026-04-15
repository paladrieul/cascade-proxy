package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cascade-proxy/internal/middleware"
	"github.com/cascade-proxy/internal/ratelimiter"
)

func newTestPathRateLimitMiddleware(rules []middleware.PathRateLimitRule, fallback *ratelimiter.Config) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	cfg := middleware.DefaultPathRateLimitConfig()
	cfg.Rules = rules
	cfg.Fallback = fallback
	return middleware.NewPathRateLimitMiddleware(cfg, handler)
}

func TestPathRateLimitAllowsMatchingPathWithinBurst(t *testing.T) {
	rules := []middleware.PathRateLimitRule{
		{Prefix: "/api", Rate: 10, Burst: 3, TTL: time.Minute},
	}
	h := newTestPathRateLimitMiddleware(rules, nil)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestPathRateLimitBlocks429WhenBurstExceeded(t *testing.T) {
	rules := []middleware.PathRateLimitRule{
		{Prefix: "/api", Rate: 0.001, Burst: 1, TTL: time.Minute},
	}
	h := newTestPathRateLimitMiddleware(rules, nil)

	send := func() int {
		req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	send() // consume burst
	if got := send(); got != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", got)
	}
}

func TestPathRateLimitUsesFirstMatchingRule(t *testing.T) {
	rules := []middleware.PathRateLimitRule{
		{Prefix: "/api/admin", Rate: 0.001, Burst: 1, TTL: time.Minute},
		{Prefix: "/api", Rate: 100, Burst: 100, TTL: time.Minute},
	}
	h := newTestPathRateLimitMiddleware(rules, nil)

	// /api/admin should hit the strict rule
	req := httptest.NewRequest(http.MethodGet, "/api/admin/action", nil)
	req.RemoteAddr = "5.5.5.5:1"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first request should pass, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request to strict path should be 429, got %d", rec2.Code)
	}
}

func TestPathRateLimitFallbackAppliedForUnmatchedPath(t *testing.T) {
	fallback := &ratelimiter.Config{Rate: 0.001, Burst: 1, TTL: time.Minute}
	h := newTestPathRateLimitMiddleware(nil, fallback)

	send := func() int {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.RemoteAddr = "9.9.9.9:80"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	send()
	if got := send(); got != http.StatusTooManyRequests {
		t.Fatalf("expected 429 from fallback, got %d", got)
	}
}

func TestPathRateLimitNoRulesNoFallbackAlwaysAllows(t *testing.T) {
	h := newTestPathRateLimitMiddleware(nil, nil)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/anything", nil)
		req.RemoteAddr = "3.3.3.3:1"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}
