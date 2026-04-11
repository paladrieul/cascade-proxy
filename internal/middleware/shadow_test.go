package middleware_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/casualjim/cascade-proxy/internal/middleware"
)

func newTestShadowMiddleware(shadowURL string) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := middleware.DefaultShadowConfig(shadowURL, logger)
	cfg.Timeout = 500 * time.Millisecond
	primary := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("primary"))
	})
	return middleware.NewShadowMiddleware(cfg, primary)
}

func TestShadowPrimaryResponseUnaffected(t *testing.T) {
	shadow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer shadow.Close()

	h := newTestShadowMiddleware(shadow.URL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from primary, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "primary" {
		t.Fatalf("expected body 'primary', got %q", body)
	}
}

func TestShadowRequestReachesShadowBackend(t *testing.T) {
	var shadowHit atomic.Int32
	shadow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shadowHit.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer shadow.Close()

	h := newTestShadowMiddleware(shadow.URL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	h.ServeHTTP(rec, req)

	// Give goroutine time to complete.
	time.Sleep(100 * time.Millisecond)

	if shadowHit.Load() != 1 {
		t.Fatalf("expected shadow backend to be hit once, got %d", shadowHit.Load())
	}
}

func TestShadowSetsXShadowHeader(t *testing.T) {
	var gotHeader string
	shadow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Shadow-Request")
		w.WriteHeader(http.StatusOK)
	}))
	defer shadow.Close()

	h := newTestShadowMiddleware(shadow.URL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/data", strings.NewReader(`{"key":"val"}`))
	h.ServeHTTP(rec, req)

	time.Sleep(100 * time.Millisecond)

	if gotHeader != "true" {
		t.Fatalf("expected X-Shadow-Request: true, got %q", gotHeader)
	}
}

func TestShadowDoesNotBlockOnUnreachableBackend(t *testing.T) {
	h := newTestShadowMiddleware("http://127.0.0.1:1")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fast", nil)

	start := time.Now()
	h.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Fatalf("primary handler blocked for %s; expected near-instant", elapsed)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
