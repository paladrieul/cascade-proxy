package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cascade-proxy/internal/middleware"
)

// TestMaintenanceIntegrationWithLogger verifies that the maintenance middleware
// composes correctly with the Logger middleware: the logger still records the
// 503 status emitted during maintenance mode.
func TestMaintenanceIntegrationWithLogger(t *testing.T) {
	cfg := middleware.DefaultMaintenanceConfig()
	cfg.Enabled.Store(true)

	var logged int
	logger := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := middleware.NewResponseRecorder(w)
			next.ServeHTTP(rec, r)
			logged = rec.Status()
		})
	}

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := logger(middleware.NewMaintenanceMiddleware(cfg)(base))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))

	if logged != http.StatusServiceUnavailable {
		t.Fatalf("logger expected to capture 503, got %d", logged)
	}
}

// TestMaintenanceIntegrationWithRequestID verifies that the request-ID
// middleware upstream of maintenance mode still injects its header even when
// the maintenance middleware short-circuits the chain.
func TestMaintenanceIntegrationWithRequestID(t *testing.T) {
	cfg := middleware.DefaultMaintenanceConfig()
	cfg.Enabled.Store(true)

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := middleware.NewRequestIDMiddleware(middleware.NewMaintenanceMiddleware(cfg)(base))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header to be set by RequestID middleware")
	}
}
