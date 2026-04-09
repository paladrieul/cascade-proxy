package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func defaultConfig() Config {
	return Config{
		RequestsPerSecond: 10,
		Burst:             3,
	}
}

func TestAllowsUpToBurst(t *testing.T) {
	rl := New(defaultConfig())
	for i := 0; i < 3; i++ {
		if !rl.Allow("client1") {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}
}

func TestBlocksAfterBurstExceeded(t *testing.T) {
	rl := New(defaultConfig())
	for i := 0; i < 3; i++ {
		rl.Allow("client2")
	}
	if rl.Allow("client2") {
		t.Fatal("expected request to be blocked after burst exceeded")
	}
}

func TestRefillsTokensOverTime(t *testing.T) {
	rl := New(Config{RequestsPerSecond: 100, Burst: 1})
	rl.Allow("client3") // consume the single token

	time.Sleep(20 * time.Millisecond)

	if !rl.Allow("client3") {
		t.Fatal("expected token to be refilled after sleep")
	}
}

func TestIndependentKeysAreIsolated(t *testing.T) {
	rl := New(defaultConfig())
	for i := 0; i < 3; i++ {
		rl.Allow("clientA")
	}
	if !rl.Allow("clientB") {
		t.Fatal("clientB should not be affected by clientA's usage")
	}
}

func TestMiddlewareReturns429WhenLimitExceeded(t *testing.T) {
	rl := New(Config{RequestsPerSecond: 1, Burst: 1})

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	makeReq := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		handler.ServeHTTP(rec, req)
		return rec
	}

	first := makeReq()
	if first.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", first.Code)
	}

	second := makeReq()
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", second.Code)
	}
}

func TestMiddlewareForwardsXForwardedFor(t *testing.T) {
	rl := New(Config{RequestsPerSecond: 1, Burst: 1})
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
