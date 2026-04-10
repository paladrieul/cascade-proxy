package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cascade-proxy/internal/middleware"
)

func newTestRequestIDMiddleware(next http.Handler) http.Handler {
	return middleware.NewRequestIDMiddleware(next)
}

func TestRequestIDGeneratesIDWhenAbsent(t *testing.T) {
	var capturedID string
	handler := newTestRequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = r.Header.Get(middleware.RequestIDHeader)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if capturedID == "" {
		t.Fatal("expected a generated request ID, got empty string")
	}
	if got := rr.Header().Get(middleware.RequestIDHeader); got != capturedID {
		t.Fatalf("response header mismatch: want %q got %q", capturedID, got)
	}
}

func TestRequestIDPreservesIncomingID(t *testing.T) {
	const existingID = "my-trace-123"
	var capturedID string
	handler := newTestRequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = r.Header.Get(middleware.RequestIDHeader)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.RequestIDHeader, existingID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if capturedID != existingID {
		t.Fatalf("want %q got %q", existingID, capturedID)
	}
	if got := rr.Header().Get(middleware.RequestIDHeader); got != existingID {
		t.Fatalf("response header: want %q got %q", existingID, got)
	}
}

func TestRequestIDStoresIDInContext(t *testing.T) {
	var ctxID string
	handler := newTestRequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = middleware.RequestIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if ctxID == "" {
		t.Fatal("expected request ID in context, got empty string")
	}
}

func TestRequestIDFromContextReturnsEmptyWhenMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if id := middleware.RequestIDFromContext(req.Context()); id != "" {
		t.Fatalf("expected empty string, got %q", id)
	}
}

func TestRequestIDIsUniquePerRequest(t *testing.T) {
	ids := make([]string, 3)
	i := 0
	handler := newTestRequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids[i] = r.Header.Get(middleware.RequestIDHeader)
		i++
	}))

	for range ids {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("duplicate request ID detected: %q", id)
		}
		seen[id] = true
	}
}
