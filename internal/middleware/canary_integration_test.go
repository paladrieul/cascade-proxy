package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCanaryIntegrationWithRequestID verifies that the canary middleware
// correctly forwards the X-Request-ID header injected by RequestIDMiddleware.
func TestCanaryIntegrationWithRequestID(t *testing.T) {
	var receivedID string

	canary := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedID = r.Header.Get("X-Request-Id")
		w.WriteHeader(http.StatusOK)
	})

	canarySrv := httptest.NewServer(canary)
	t.Cleanup(canarySrv.Close)

	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	cfg := CanaryConfig{Weight: 100, HeaderOverride: ""}
	cfg.CanaryURL = canarySrv.URL
	canaryHandler, err := NewCanaryMiddleware(cfg, primary)
	if err != nil {
		t.Fatalf("NewCanaryMiddleware: %v", err)
	}

	ridHandler := NewRequestIDMiddleware(canaryHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "integration-test-id")
	rec := httptest.NewRecorder()
	ridHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 from canary, got %d", rec.Code)
	}
	if receivedID != "integration-test-id" {
		t.Errorf("expected request ID to be forwarded, got %q", receivedID)
	}
}

// TestCanaryIntegrationWithLogger verifies that canary + logger middleware
// stack does not panic and returns the expected status code.
func TestCanaryIntegrationWithLogger(t *testing.T) {
	canary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	canarySrv := httptest.NewServer(canary)
	t.Cleanup(canarySrv.Close)

	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	cfg := CanaryConfig{Weight: 100, HeaderOverride: "X-Force-Canary"}
	cfg.CanaryURL = canarySrv.URL
	canaryHandler, err := NewCanaryMiddleware(cfg, primary)
	if err != nil {
		t.Fatalf("NewCanaryMiddleware: %v", err)
	}

	stack := Logger(canaryHandler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Force-Canary", "1")
	rec := httptest.NewRecorder()
	stack.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202 from canary via logger stack, got %d", rec.Code)
	}
}
