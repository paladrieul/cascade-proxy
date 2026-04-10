package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cascade-proxy/internal/middleware"
)

// TestRequestIDIntegrationWithLogger verifies that the request ID written into
// the context by NewRequestIDMiddleware is accessible to a handler sitting
// deeper in the chain — simulating how a logger or tracer would consume it.
func TestRequestIDIntegrationWithLogger(t *testing.T) {
	var loggedID string

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedID = middleware.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	chain := middleware.NewRequestIDMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	rr := httptest.NewRecorder()
	chain.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", rr.Code)
	}
	if loggedID == "" {
		t.Fatal("inner handler did not receive a request ID via context")
	}
	if got := rr.Header().Get(middleware.RequestIDHeader); got != loggedID {
		t.Fatalf("context ID %q does not match response header %q", loggedID, got)
	}
}

// TestRequestIDIntegrationPreservesIDThroughMultipleMiddleware checks that a
// client-supplied ID survives a realistic middleware stack.
func TestRequestIDIntegrationPreservesIDThroughMultipleMiddleware(t *testing.T) {
	const clientID = "client-supplied-id-abc"

	var receivedID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedID = middleware.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusAccepted)
	})

	// Wrap with a second, unrelated middleware layer to ensure context
	// propagation is not broken by intermediate handlers.
	middle := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate work without touching the request ID.
		inner.ServeHTTP(w, r)
	})

	chain := middleware.NewRequestIDMiddleware(middle)

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.Header.Set(middleware.RequestIDHeader, clientID)
	rr := httptest.NewRecorder()
	chain.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("want 202 got %d", rr.Code)
	}
	if receivedID != clientID {
		t.Fatalf("want %q got %q", clientID, receivedID)
	}
}
