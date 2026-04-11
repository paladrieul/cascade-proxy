package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casualjim/cascade-proxy/internal/middleware"
)

func newTestBodySizeMiddleware(maxBytes int64) http.Handler {
	cfg := middleware.BodySizeConfig{MaxBytes: maxBytes}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})
	return middleware.NewBodySizeMiddleware(cfg)(handler)
}

func TestBodySizeAllowsRequestWithinLimit(t *testing.T) {
	h := newTestBodySizeMiddleware(100)
	body := strings.Repeat("a", 50)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != body {
		t.Fatalf("expected body %q, got %q", body, rec.Body.String())
	}
}

func TestBodySizeRejects413WhenContentLengthExceedsLimit(t *testing.T) {
	h := newTestBodySizeMiddleware(10)
	body := strings.Repeat("a", 50)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.ContentLength = 50
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestBodySizeRejects413WhenBodyExceedsLimitDuringRead(t *testing.T) {
	h := newTestBodySizeMiddleware(10)
	body := bytes.Repeat([]byte("b"), 50)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	// ContentLength not set so the early check is skipped; limit is enforced by MaxBytesReader
	req.ContentLength = -1
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError && rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 or 500, got %d", rec.Code)
	}
}

func TestBodySizeDefaultConfig(t *testing.T) {
	cfg := middleware.DefaultBodySizeConfig()
	if cfg.MaxBytes != 1<<20 {
		t.Fatalf("expected default MaxBytes 1048576, got %d", cfg.MaxBytes)
	}
}

func TestBodySizeAllowsNilBody(t *testing.T) {
	h := newTestBodySizeMiddleware(100)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Body = nil
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
