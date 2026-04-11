package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func newTestUpstreamMiddleware(threshold time.Duration, buf *bytes.Buffer, next http.Handler) http.Handler {
	logger := slog.New(slog.NewTextHandler(buf, nil))
	cfg := UpstreamConfig{
		StatusHeader:  "X-Upstream-Ms",
		SlowThreshold: threshold,
		Logger:        logger,
	}
	return NewUpstreamMiddleware(cfg, next)
}

func TestUpstreamWritesLatencyHeader(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := newTestUpstreamMiddleware(500*time.Millisecond, buf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	handler.ServeHTTP(rec, req)

	val := rec.Header().Get("X-Upstream-Ms")
	if val == "" {
		t.Fatal("expected X-Upstream-Ms header to be set")
	}
	ms, err := strconv.ParseInt(val, 10, 64)
	if err != nil || ms < 0 {
		t.Fatalf("unexpected latency value: %s", val)
	}
}

func TestUpstreamLogsSlowRequest(t *testing.T) {
	buf := &bytes.Buffer{}
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	handler := newTestUpstreamMiddleware(5*time.Millisecond, buf, slow)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	handler.ServeHTTP(rec, req)

	if !bytes.Contains(buf.Bytes(), []byte("slow upstream response")) {
		t.Errorf("expected slow upstream log, got: %s", buf.String())
	}
}

func TestUpstreamDoesNotLogFastRequest(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := newTestUpstreamMiddleware(500*time.Millisecond, buf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	handler.ServeHTTP(rec, req)

	if bytes.Contains(buf.Bytes(), []byte("slow upstream")) {
		t.Errorf("did not expect slow upstream log for fast request")
	}
}

func TestUpstreamDefaultConfig(t *testing.T) {
	cfg := DefaultUpstreamConfig()
	if cfg.StatusHeader != "X-Upstream-Ms" {
		t.Errorf("unexpected default header: %s", cfg.StatusHeader)
	}
	if cfg.SlowThreshold != 500*time.Millisecond {
		t.Errorf("unexpected default threshold: %v", cfg.SlowThreshold)
	}
	if cfg.Logger == nil {
		t.Error("expected non-nil default logger")
	}
}
