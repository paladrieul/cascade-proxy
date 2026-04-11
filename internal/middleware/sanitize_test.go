package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestSanitizeMiddleware(cfg SanitizeConfig, next http.Handler) http.Handler {
	return NewSanitizeMiddleware(cfg, next)
}

func TestSanitizeStripsDisallowedHeaders(t *testing.T) {
	cfg := DefaultSanitizeConfig()
	var capturedHeader string
	handler := newTestSanitizeMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Forwarded-For")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedHeader != "" {
		t.Errorf("expected X-Forwarded-For to be stripped, got %q", capturedHeader)
	}
}

func TestSanitizeAllowsPermittedMethod(t *testing.T) {
	cfg := DefaultSanitizeConfig()
	handler := newTestSanitizeMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/data", strings.NewReader("body"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestSanitizeRejects405OnDisallowedMethod(t *testing.T) {
	cfg := DefaultSanitizeConfig()
	cfg.AllowedMethods = []string{http.MethodGet}
	handler := newTestSanitizeMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestSanitizeRejects400OnTooManyQueryParams(t *testing.T) {
	cfg := DefaultSanitizeConfig()
	cfg.MaxQueryParams = 2
	handler := newTestSanitizeMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/?a=1&b=2&c=3", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestSanitizeAllowsRequestWithinQueryParamLimit(t *testing.T) {
	cfg := DefaultSanitizeConfig()
	cfg.MaxQueryParams = 3
	handler := newTestSanitizeMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/?a=1&b=2", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
