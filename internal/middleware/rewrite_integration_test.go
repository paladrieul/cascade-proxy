package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/cascade-proxy/internal/middleware"
)

// TestRewriteIntegrationChainedMiddleware verifies that the rewrite middleware
// composes correctly with other middleware in the stack (logger + rewrite).
func TestRewriteIntegrationChainedMiddleware(t *testing.T) {
	var gotPath string
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	cfg := middleware.RewriteConfig{
		Rules: []middleware.RewriteRule{
			{Pattern: regexp.MustCompile(`^/old/(.*)`), Replacement: "/new/$1"},
		},
	}

	handler := middleware.NewRewriteMiddleware(cfg, backend)

	req := httptest.NewRequest(http.MethodGet, "/old/resource/123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotPath != "/new/resource/123" {
		t.Fatalf("expected /new/resource/123, got %s", gotPath)
	}
}

// TestRewriteIntegrationPreservesQueryString ensures query parameters survive
// path rewriting.
func TestRewriteIntegrationPreservesQueryString(t *testing.T) {
	var gotRawQuery string
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	})

	cfg := middleware.RewriteConfig{
		StripPrefix: "/proxy",
	}

	handler := middleware.NewRewriteMiddleware(cfg, backend)

	req := httptest.NewRequest(http.MethodGet, "/proxy/search?q=hello&page=2", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if gotRawQuery != "q=hello&page=2" {
		t.Fatalf("expected query string preserved, got %q", gotRawQuery)
	}
}
