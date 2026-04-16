package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestProxyErrorMiddleware(handler http.Handler, details bool) (http.Handler, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))
	cfg := ProxyErrorConfig{Logger: logger, IncludeDetails: details}
	return NewProxyErrorMiddleware(cfg, handler), buf
}

func TestProxyErrorPassesThroughOnSuccess(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h, buf := newTestProxyErrorMiddleware(backend, false)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ok", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no log output, got: %s", buf.String())
	}
}

func TestProxyErrorLogs502(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})
	h, buf := newTestProxyErrorMiddleware(backend, false)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
	if !strings.Contains(buf.String(), "proxy upstream error") {
		t.Fatalf("expected log entry, got: %s", buf.String())
	}
}

func TestProxyErrorLogs503And504(t *testing.T) {
	for _, code := range []int{http.StatusServiceUnavailable, http.StatusGatewayTimeout} {
		backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
		})
		h, buf := newTestProxyErrorMiddleware(backend, false)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != code {
			t.Fatalf("expected %d, got %d", code, rec.Code)
		}
		if !strings.Contains(buf.String(), "proxy upstream error") {
			t.Fatalf("expected log for %d, got: %s", code, buf.String())
		}
	}
}

func TestProxyErrorSetsHeaderWhenDetailsEnabled(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})
	h, _ := newTestProxyErrorMiddleware(backend, true)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("X-Proxy-Error") == "" {
		t.Fatal("expected X-Proxy-Error header to be set")
	}
}

func TestProxyErrorDefaultConfig(t *testing.T) {
	cfg := DefaultProxyErrorConfig()
	if cfg.Logger == nil {
		t.Fatal("expected non-nil logger")
	}
	if cfg.IncludeDetails {
		t.Fatal("expected IncludeDetails to be false by default")
	}
}
