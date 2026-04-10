package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestCompressMiddleware(cfg CompressConfig, body string, status int) *httptest.ResponseRecorder {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		io.WriteString(w, body) //nolint:errcheck
	})
	mw := NewCompressMiddleware(cfg, handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	mw.ServeHTTP(rec, req)
	return rec
}

func TestCompressesLargeResponse(t *testing.T) {
	body := strings.Repeat("cascade-proxy-test-", 100) // > 1024 bytes
	cfg := DefaultCompressConfig
	rec := newTestCompressMiddleware(cfg, body, http.StatusOK)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected Content-Encoding: gzip for large response")
	}

	r, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer r.Close()
	got, _ := io.ReadAll(r)
	if string(got) != body {
		t.Errorf("decompressed body mismatch: got %q", string(got))
	}
}

func TestDoesNotCompressSmallResponse(t *testing.T) {
	body := "small"
	cfg := DefaultCompressConfig // minLength = 1024
	rec := newTestCompressMiddleware(cfg, body, http.StatusOK)

	if enc := rec.Header().Get("Content-Encoding"); enc == "gzip" {
		t.Fatal("expected no compression for small response")
	}
	if rec.Body.String() != body {
		t.Errorf("body mismatch: got %q want %q", rec.Body.String(), body)
	}
}

func TestCompressSkipsWithoutAcceptEncoding(t *testing.T) {
	body := strings.Repeat("x", 2000)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, body) //nolint:errcheck
	})
	mw := NewCompressMiddleware(DefaultCompressConfig, handler)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Accept-Encoding header
	mw.ServeHTTP(rec, req)

	if enc := rec.Header().Get("Content-Encoding"); enc == "gzip" {
		t.Fatal("should not compress without Accept-Encoding: gzip")
	}
}

func TestCompressPreservesStatusCode(t *testing.T) {
	body := strings.Repeat("error-body-", 200)
	cfg := DefaultCompressConfig
	rec := newTestCompressMiddleware(cfg, body, http.StatusInternalServerError)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestCompressDefaultConfig(t *testing.T) {
	if DefaultCompressConfig.Level != gzip.DefaultCompression {
		t.Errorf("expected default level %d, got %d", gzip.DefaultCompression, DefaultCompressConfig.Level)
	}
	if DefaultCompressConfig.MinLength != 1024 {
		t.Errorf("expected MinLength 1024, got %d", DefaultCompressConfig.MinLength)
	}
}
