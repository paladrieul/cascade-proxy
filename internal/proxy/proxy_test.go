package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cascade-proxy/internal/circuitbreaker"
	"github.com/cascade-proxy/internal/proxy"
)

func newTestProxy(t *testing.T, target string) *proxy.Proxy {
	t.Helper()
	cfg := proxy.Config{
		TargetURL:      target,
		MaxRetries:     1,
		RetryDelay:     10 * time.Millisecond,
		RequestTimeout: 2 * time.Second,
		CBConfig:       circuitbreaker.DefaultConfig(),
	}
	p, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("proxy.New: %v", err)
	}
	return p
}

func TestProxyForwardsRequest(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))
	defer backend.Close()

	p := newTestProxy(t, backend.URL)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "pong" {
		t.Errorf("expected body 'pong', got %q", body)
	}
}

func TestProxyRetriesOnServerError(t *testing.T) {
	attempts := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	p := newTestProxy(t, backend.URL)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 after retry, got %d", rec.Code)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestProxyReturns502WhenBackendDown(t *testing.T) {
	p := newTestProxy(t, "http://127.0.0.1:19999") // nothing listening

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}

func TestProxyInvalidTargetURL(t *testing.T) {
	_, err := proxy.New(proxy.Config{TargetURL: "://bad url"})
	if err == nil {
		t.Error("expected error for invalid target URL")
	}
}
