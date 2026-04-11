package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casualjim/cascade-proxy/internal/middleware"
)

func newTestReqTransformMiddleware(cfg middleware.RequestTransformConfig) http.Handler {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})
	return middleware.NewRequestTransformMiddleware(cfg, backend)
}

func TestReqTransformPassesThroughWhenNoRules(t *testing.T) {
	cfg := middleware.DefaultRequestTransformConfig()
	h := newTestReqTransformMiddleware(cfg)

	body := `{"key":"value"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != body {
		t.Fatalf("expected %q, got %q", body, got)
	}
}

func TestReqTransformAppliesRule(t *testing.T) {
	cfg := middleware.RequestTransformConfig{
		Rules: []middleware.RequestTransformRule{
			{Find: "foo", Replace: "bar"},
		},
	}
	h := newTestReqTransformMiddleware(cfg)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("hello foo world"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != "hello bar world" {
		t.Fatalf("expected 'hello bar world', got %q", got)
	}
}

func TestReqTransformRespectsContentTypeFilter(t *testing.T) {
	cfg := middleware.RequestTransformConfig{
		Rules: []middleware.RequestTransformRule{
			{Find: "foo", Replace: "bar", ContentType: "application/json"},
		},
	}
	h := newTestReqTransformMiddleware(cfg)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("hello foo"))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// rule should NOT apply because content type doesn't match
	if got := rec.Body.String(); got != "hello foo" {
		t.Fatalf("expected body unchanged, got %q", got)
	}
}

func TestReqTransformUpdatesContentLength(t *testing.T) {
	cfg := middleware.RequestTransformConfig{
		Rules: []middleware.RequestTransformRule{
			{Find: "hi", Replace: "hello"},
		},
	}
	var capturedLen int64
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedLen = r.ContentLength
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.NewRequestTransformMiddleware(cfg, backend)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("hi there"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if capturedLen != int64(len("hello there")) {
		t.Fatalf("expected ContentLength %d, got %d", len("hello there"), capturedLen)
	}
}
