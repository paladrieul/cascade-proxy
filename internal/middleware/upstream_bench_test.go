package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkUpstreamMiddlewareFastBackend(b *testing.B) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))
	cfg := UpstreamConfig{
		StatusHeader:  "X-Upstream-Ms",
		SlowThreshold: 500 * time.Millisecond,
		Logger:        logger,
	}
	handler := NewUpstreamMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/bench", nil)
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkUpstreamMiddlewareSlowBackend(b *testing.B) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))
	cfg := UpstreamConfig{
		StatusHeader:  "X-Upstream-Ms",
		SlowThreshold: 1 * time.Millisecond,
		Logger:        logger,
	}
	handler := NewUpstreamMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/bench-slow", nil)
		handler.ServeHTTP(rec, req)
	}
}
