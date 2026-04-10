package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func newTestRewriteMiddleware(cfg RewriteConfig) (http.Handler, *string) {
	captured := new(string)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	return NewRewriteMiddleware(cfg, inner), captured
}

func TestRewriteNoRulesPassesThrough(t *testing.T) {
	h, captured := newTestRewriteMiddleware(DefaultRewriteConfig())
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if *captured != "/api/users" {
		t.Fatalf("expected /api/users, got %s", *captured)
	}
}

func TestRewriteAppliesFirstMatchingRule(t *testing.T) {
	cfg := RewriteConfig{
		Rules: []RewriteRule{
			{Pattern: regexp.MustCompile(`^/v1/(.*)`), Replacement: "/api/$1"},
			{Pattern: regexp.MustCompile(`^/v2/(.*)`), Replacement: "/api/v2/$1"},
		},
	}
	h, captured := newTestRewriteMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if *captured != "/api/users" {
		t.Fatalf("expected /api/users, got %s", *captured)
	}
}

func TestRewriteSkipsNonMatchingRules(t *testing.T) {
	cfg := RewriteConfig{
		Rules: []RewriteRule{
			{Pattern: regexp.MustCompile(`^/v1/(.*)`), Replacement: "/api/$1"},
		},
	}
	h, captured := newTestRewriteMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if *captured != "/health" {
		t.Fatalf("expected /health, got %s", *captured)
	}
}

func TestRewriteStripPrefix(t *testing.T) {
	cfg := RewriteConfig{StripPrefix: "/api"}
	h, captured := newTestRewriteMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if *captured != "/users" {
		t.Fatalf("expected /users, got %s", *captured)
	}
}

func TestRewriteStripPrefixRootFallback(t *testing.T) {
	cfg := RewriteConfig{StripPrefix: "/api"}
	h, captured := newTestRewriteMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if *captured != "/" {
		t.Fatalf("expected /, got %s", *captured)
	}
}

func TestRewriteRuleThenStripPrefix(t *testing.T) {
	cfg := RewriteConfig{
		Rules:       []RewriteRule{{Pattern: regexp.MustCompile(`^/legacy(.*)`), Replacement: "/api$1"}},
		StripPrefix: "/api",
	}
	h, captured := newTestRewriteMiddleware(cfg)
	req := httptest.NewRequest(http.MethodGet, "/legacy/orders", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
	if *captured != "/orders" {
		t.Fatalf("expected /orders, got %s", *captured)
	}
}
