package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// TestTransformIntegrationWithLogger verifies that the transform middleware
// composes correctly with the logger middleware.
func TestTransformIntegrationWithLogger(t *testing.T) {
	var logBuf strings.Builder

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world"))
	})

	cfg := ResponseTransformConfig{
		Rules: []TransformRule{
			{Pattern: regexp.MustCompile(`world`), Replacement: "cascade"},
		},
	}

	chain := Logger(&logBuf, NewResponseTransformMiddleware(cfg, backend))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/transform", nil)
	chain.ServeHTTP(rec, req)

	body, _ := io.ReadAll(rec.Body)
	if string(body) != "hello cascade" {
		t.Fatalf("expected 'hello cascade', got %q", string(body))
	}
	if !strings.Contains(logBuf.String(), "GET") {
		t.Fatal("expected logger to record the request method")
	}
}

// TestTransformIntegrationWithRequestID verifies that trace IDs survive the
// transform middleware without corruption.
func TestTransformIntegrationWithRequestID(t *testing.T) {
	const fixedID = "test-request-id-42"

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"msg":"ok"}`))
	})

	cfg := ResponseTransformConfig{
		Rules: []TransformRule{
			{Pattern: regexp.MustCompile(`ok`), Replacement: "done"},
		},
		ContentTypes: []string{"application/json"},
	}

	chain := NewRequestIDMiddleware(NewResponseTransformMiddleware(cfg, backend))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", fixedID)
	chain.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-Id") != fixedID {
		t.Fatalf("expected request ID %q to be preserved", fixedID)
	}

	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "done") {
		t.Fatalf("expected transformed body, got %q", string(body))
	}
}
