package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestCORSMiddleware(cfg CORSConfig) http.Handler {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return NewCORSMiddleware(cfg)(inner)
}

func TestCORSPassesThroughNonCORSRequest(t *testing.T) {
	h := newTestCORSMiddleware(DefaultCORSConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header for non-CORS request, got %q", got)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCORSSetsWildcardOrigin(t *testing.T) {
	h := newTestCORSMiddleware(DefaultCORSConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected '*', got %q", got)
	}
}

func TestCORSRejectsDisallowedOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"https://allowed.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}
	h := newTestCORSMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header for disallowed origin, got %q", got)
	}
}

func TestCORSHandlesPreflight(t *testing.T) {
	h := newTestCORSMiddleware(DefaultCORSConfig())
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods header on preflight")
	}
	if got := rec.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Errorf("expected max-age 86400, got %q", got)
	}
}

func TestCORSSetsVaryHeaderForSpecificOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"https://allowed.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}
	h := newTestCORSMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://allowed.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.com" {
		t.Errorf("expected specific origin, got %q", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("expected Vary: Origin, got %q", got)
	}
}
