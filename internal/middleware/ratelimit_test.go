package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cascade-proxy/internal/ratelimiter"
)

func newTestRateLimitMiddleware(rps float64, burst int) (*RateLimitMiddleware, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "", 0)
	rl := ratelimiter.New(ratelimiter.Config{
		RequestsPerSecond: rps,
		Burst:             burst,
	})
	return NewRateLimitMiddleware(rl, logger), buf
}

func TestRateLimitAllowsRequestWithinLimit(t *testing.T) {
	mw, _ := newTestRateLimitMiddleware(10, 5)
	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.RemoteAddr = "127.0.0.1:9000"
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRateLimitBlocks429(t *testing.T) {
	mw, _ := newTestRateLimitMiddleware(1, 1)
	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	send := func() int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.2:8080"
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	send() // consume burst
	if code := send(); code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", code)
	}
}

func TestRateLimitLogsRejection(t *testing.T) {
	mw, buf := newTestRateLimitMiddleware(1, 1)
	handler := mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	send := func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.3:1111"
		handler.ServeHTTP(rec, req)
	}

	send()
	send() // this should be rejected and logged

	if !strings.Contains(buf.String(), "rate limit exceeded") {
		t.Fatalf("expected log to contain 'rate limit exceeded', got: %s", buf.String())
	}
}
