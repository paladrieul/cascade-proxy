package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func newTestTransformMiddleware(cfg ResponseTransformConfig, body string, status int, ct string) *httptest.ResponseRecorder {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
	handler := NewResponseTransformMiddleware(cfg, backend)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)
	return rec
}

func TestTransformAppliesRuleToBody(t *testing.T) {
	cfg := ResponseTransformConfig{
		Rules: []TransformRule{
			{Pattern: regexp.MustCompile(`foo`), Replacement: "bar"},
		},
	}
	rec := newTestTransformMiddleware(cfg, "hello foo world", http.StatusOK, "text/plain")
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "hello bar world" {
		t.Fatalf("expected 'hello bar world', got %q", string(body))
	}
}

func TestTransformPassesThroughWhenNoRules(t *testing.T) {
	cfg := DefaultResponseTransformConfig()
	rec := newTestTransformMiddleware(cfg, "unchanged", http.StatusOK, "text/plain")
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "unchanged" {
		t.Fatalf("expected 'unchanged', got %q", string(body))
	}
}

func TestTransformRespectsContentTypeFilter(t *testing.T) {
	cfg := ResponseTransformConfig{
		Rules:        []TransformRule{{Pattern: regexp.MustCompile(`foo`), Replacement: "bar"}},
		ContentTypes: []string{"application/json"},
	}
	// Content-Type is text/plain — should NOT be transformed.
	rec := newTestTransformMiddleware(cfg, "foo", http.StatusOK, "text/plain")
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "foo" {
		t.Fatalf("expected body untouched, got %q", string(body))
	}
}

func TestTransformAppliesMultipleRulesInOrder(t *testing.T) {
	cfg := ResponseTransformConfig{
		Rules: []TransformRule{
			{Pattern: regexp.MustCompile(`foo`), Replacement: "baz"},
			{Pattern: regexp.MustCompile(`baz`), Replacement: "qux"},
		},
	}
	rec := newTestTransformMiddleware(cfg, "foo", http.StatusOK, "")
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "qux" {
		t.Fatalf("expected 'qux', got %q", string(body))
	}
}

func TestTransformPreservesStatusCode(t *testing.T) {
	cfg := ResponseTransformConfig{
		Rules: []TransformRule{{Pattern: regexp.MustCompile(`x`), Replacement: "y"}},
	}
	rec := newTestTransformMiddleware(cfg, "x", http.StatusCreated, "text/plain")
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}
