package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func BenchmarkResponseTransformSmallBody(b *testing.B) {
	cfg := ResponseTransformConfig{
		Rules: []TransformRule{
			{Pattern: regexp.MustCompile(`foo`), Replacement: "bar"},
		},
	}
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("foo baz foo"))
	})
	handler := NewResponseTransformMiddleware(cfg, backend)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkResponseTransformLargeBody(b *testing.B) {
	cfg := ResponseTransformConfig{
		Rules: []TransformRule{
			{Pattern: regexp.MustCompile(`internal\.svc`), Replacement: "public.example.com"},
		},
		ContentTypes: []string{"application/json"},
	}
	payload := strings.Repeat(`{"url":"https://internal.svc/path"}`, 500)
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(payload))
	})
	handler := NewResponseTransformMiddleware(cfg, backend)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rec, req)
	}
}
