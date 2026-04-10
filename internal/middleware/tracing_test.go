package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newTestTracingMiddleware(next http.Handler) http.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return NewTracingMiddleware(logger, next)
}

func TestTracingInjectsTraceIDHeader(t *testing.T) {
	handler := newTestTracingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get(TraceIDHeader) == "" {
		t.Error("expected X-Trace-ID header to be set")
	}
}

func TestTracingPreservesIncomingTraceID(t *testing.T) {
	const existingID = "my-trace-123"

	handler := newTestTracingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(TraceIDHeader, existingID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(TraceIDHeader); got != existingID {
		t.Errorf("expected trace ID %q, got %q", existingID, got)
	}
}

func TestTracingStoresIDInContext(t *testing.T) {
	var capturedID string

	handler := newTestTracingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = TraceIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Error("expected trace ID in context, got empty string")
	}
	if capturedID != rec.Header().Get(TraceIDHeader) {
		t.Errorf("context trace ID %q does not match header %q", capturedID, rec.Header().Get(TraceIDHeader))
	}
}

func TestTraceIDFromContextReturnsEmptyWhenMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if id := TraceIDFromContext(req.Context()); id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}

func TestTracingPassesThroughStatusCode(t *testing.T) {
	handler := newTestTracingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("expected status %d, got %d", http.StatusTeapot, rec.Code)
	}
}
