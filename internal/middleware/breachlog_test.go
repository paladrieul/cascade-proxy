package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestBreachLogMiddleware(threshold float64, windowSize int, buf *bytes.Buffer) func(http.Handler) http.Handler {
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cfg := BreachLogConfig{
		Logger:         logger,
		WindowSize:     windowSize,
		ErrorThreshold: threshold,
		KeyFunc:        func(r *http.Request) string { return r.URL.Path },
	}
	return NewBreachLogMiddleware(cfg)
}

func TestBreachLogNoBreachBelowThreshold(t *testing.T) {
	var buf bytes.Buffer
	mw := newTestBreachLogMiddleware(0.8, 4, &buf)
	// 2 errors out of 4 = 0.5 ratio, below 0.8 threshold
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fail") == "1" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	handler := mw(backend)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api?fail=1", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
	if strings.Contains(buf.String(), "breach") {
		t.Errorf("expected no breach log, got: %s", buf.String())
	}
}

func TestBreachLogLogsWhenThresholdExceeded(t *testing.T) {
	var buf bytes.Buffer
	mw := newTestBreachLogMiddleware(0.5, 4, &buf)
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	handler := mw(backend)
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/svc", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
	if !strings.Contains(buf.String(), "breach") {
		t.Errorf("expected breach log, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "/svc") {
		t.Errorf("expected key /svc in log, got: %s", buf.String())
	}
}

func TestBreachLogIsolatesKeysByPath(t *testing.T) {
	var buf bytes.Buffer
	mw := newTestBreachLogMiddleware(0.5, 4, &buf)
	goodBackend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/bad") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	handler := mw(goodBackend)
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/bad", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/good", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
	if strings.Contains(buf.String(), "/good") {
		t.Errorf("expected /good path not to breach, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "/bad") {
		t.Errorf("expected /bad path to breach, got: %s", buf.String())
	}
}

func TestBreachLogDefaultConfig(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	cfg := DefaultBreachLogConfig(logger)
	if cfg.WindowSize != 20 {
		t.Errorf("expected window size 20, got %d", cfg.WindowSize)
	}
	if cfg.ErrorThreshold != 0.5 {
		t.Errorf("expected threshold 0.5, got %f", cfg.ErrorThreshold)
	}
	if cfg.KeyFunc == nil {
		t.Error("expected KeyFunc to be set")
	}
}
