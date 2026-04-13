package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkCanaryMiddlewareHeaderOverride(b *testing.B) {
	canary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	canarySrv := httptest.NewServer(canary)
	b.Cleanup(canarySrv.Close)

	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := CanaryConfig{Weight: 0, HeaderOverride: "X-Canary", CanaryURL: canarySrv.URL}
	h, err := NewCanaryMiddleware(cfg, primary)
	if err != nil {
		b.Fatalf("NewCanaryMiddleware: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Canary", "true")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}
}

func BenchmarkCanaryMiddlewareWeightedSplit(b *testing.B) {
	canary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	canarySrv := httptest.NewServer(canary)
	b.Cleanup(canarySrv.Close)

	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := CanaryConfig{Weight: 50, HeaderOverride: "", CanaryURL: canarySrv.URL}
	h, err := NewCanaryMiddleware(cfg, primary)
	if err != nil {
		b.Fatalf("NewCanaryMiddleware: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}
}
