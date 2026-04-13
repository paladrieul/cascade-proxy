package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestRedirectMiddleware(cfg RedirectConfig) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return NewRedirectMiddleware(cfg, next)
}

func TestRedirectPassesThroughWhenNoRules(t *testing.T) {
	handler := newTestRedirectMiddleware(DefaultRedirectConfig())
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRedirectExactMatchReturns301(t *testing.T) {
	cfg := RedirectConfig{
		Rules: []RedirectRule{
			{From: "/old", To: "/new", Code: http.StatusMovedPermanently},
		},
	}
	handler := newTestRedirectMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/old", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/new" {
		t.Fatalf("expected Location /new, got %s", loc)
	}
}

func TestRedirectDefaultsTo302(t *testing.T) {
	cfg := RedirectConfig{
		Rules: []RedirectRule{
			{From: "/temp", To: "/elsewhere"},
		},
	}
	handler := newTestRedirectMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/temp", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
}

func TestRedirectWildcardPreservesSubPath(t *testing.T) {
	cfg := RedirectConfig{
		Rules: []RedirectRule{
			{From: "/v1/*", To: "/v2", Code: http.StatusMovedPermanently},
		},
	}
	handler := newTestRedirectMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/v1/users/42", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/v2/users/42" {
		t.Fatalf("expected /v2/users/42, got %s", loc)
	}
}

func TestRedirectPreservesQueryString(t *testing.T) {
	cfg := RedirectConfig{
		Rules: []RedirectRule{
			{From: "/search", To: "/find", Code: http.StatusMovedPermanently},
		},
	}
	handler := newTestRedirectMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/search?q=go", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if loc := rec.Header().Get("Location"); loc != "/find?q=go" {
		t.Fatalf("expected /find?q=go, got %s", loc)
	}
}

func TestRedirectFirstRuleWins(t *testing.T) {
	cfg := RedirectConfig{
		Rules: []RedirectRule{
			{From: "/page", To: "/first", Code: http.StatusFound},
			{From: "/page", To: "/second", Code: http.StatusFound},
		},
	}
	handler := newTestRedirectMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/page", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if loc := rec.Header().Get("Location"); loc != "/first" {
		t.Fatalf("expected /first, got %s", loc)
	}
}
