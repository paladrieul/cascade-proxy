package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func newTestMaintenanceMiddleware(cfg MaintenanceConfig) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	return NewMaintenanceMiddleware(cfg)(handler)
}

func TestMaintenancePassesThroughWhenDisabled(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	// Enabled defaults to false
	h := newTestMaintenanceMiddleware(cfg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/data", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMaintenanceReturns503WhenEnabled(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	cfg.Enabled.Store(true)
	h := newTestMaintenanceMiddleware(cfg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/data", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestMaintenanceAllowedPathBypassesMaintenance(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	cfg.Enabled.Store(true)
	h := newTestMaintenanceMiddleware(cfg)

	for _, path := range []string{"/healthz", "/readyz"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("path %s: expected 200, got %d", path, rec.Code)
		}
	}
}

func TestMaintenanceSetsRetryAfterHeader(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	cfg.Enabled.Store(true)
	cfg.RetryAfter = 60
	h := newTestMaintenanceMiddleware(cfg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))

	if got := rec.Header().Get("Retry-After"); got != "60" {
		t.Fatalf("expected Retry-After: 60, got %q", got)
	}
}

func TestMaintenanceResponseBodyIsJSON(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	cfg.Enabled.Store(true)
	h := newTestMaintenanceMiddleware(cfg)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("expected valid JSON body: %v", err)
	}
	if body["error"] != "maintenance" {
		t.Fatalf("expected error=maintenance, got %v", body["error"])
	}
}

func TestMaintenanceTogglesAtRuntime(t *testing.T) {
	enabled := &atomic.Bool{}
	cfg := MaintenanceConfig{
		Enabled:    enabled,
		StatusCode: http.StatusServiceUnavailable,
		Message:    "down",
	}
	h := newTestMaintenanceMiddleware(cfg)

	// First request: maintenance off
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Toggle on
	enabled.Store(true)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	// Toggle off again
	enabled.Store(false)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 after toggle off, got %d", rec.Code)
	}
}
