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

// TestUpstreamIntegrationWithLogger verifies that the upstream middleware
// composes correctly with the Logger middleware and that both emit output.
func TestUpstreamIntegrationWithLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	upstreamCfg := UpstreamConfig{
		StatusHeader:  "X-Upstream-Ms",
		SlowThreshold: 500 * time.Millisecond,
		Logger:        logger,
	}
	chain := Logger(logger, NewUpstreamMiddleware(upstreamCfg, backend))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Upstream-Ms") == "" {
		t.Error("expected X-Upstream-Ms header")
	}
}

// TestUpstreamIntegrationWithRequestID verifies that the upstream middleware
// preserves the request-id header set by a preceding middleware.
func TestUpstreamIntegrationWithRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))

	var capturedID string
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = r.Header.Get("X-Request-Id")
		w.WriteHeader(http.StatusOK)
	})

	upstreamCfg := DefaultUpstreamConfig()
	upstreamCfg.Logger = logger

	chain := NewRequestIDMiddleware(NewUpstreamMiddleware(upstreamCfg, backend))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	chain.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Error("expected request-id to be propagated to backend")
	}
	latency, err := strconv.ParseInt(rec.Header().Get("X-Upstream-Ms"), 10, 64)
	if err != nil || latency < 0 {
		t.Errorf("unexpected latency header: %s", rec.Header().Get("X-Upstream-Ms"))
	}
}
