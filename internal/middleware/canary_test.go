package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestCanaryMiddleware(t *testing.T, cfg CanaryConfig, primary, canary http.Handler) http.Handler {
	t.Helper()
	var canaryURL string
	if canary != nil {
		srv := httptest.NewServer(canary)
		t.Cleanup(srv.Close)
		canaryURL = srv.URL
	}
	cfg.CanaryURL = canaryURL
	h, err := NewCanaryMiddleware(cfg, primary)
	if err != nil {
		t.Fatalf("NewCanaryMiddleware: %v", err)
	}
	return h
}

func TestCanaryHeaderOverrideRoutesToCanary(t *testing.T) {
	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Backend", "primary")
	})
	canary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Backend", "canary")
	})

	cfg := CanaryConfig{Weight: 0, HeaderOverride: "X-Canary"}
	h := newTestCanaryMiddleware(t, cfg, primary, canary)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Canary", "true")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Canary-Routed"); got != "true" {
		t.Errorf("expected X-Canary-Routed=true, got %q", got)
	}
}

func TestCanaryZeroWeightRoutesPrimary(t *testing.T) {
	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Backend", "primary")
	})
	canary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Backend", "canary")
	})

	cfg := CanaryConfig{Weight: 0, HeaderOverride: ""}
	h := newTestCanaryMiddleware(t, cfg, primary, canary)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Canary-Routed"); got != "" {
		t.Errorf("expected no canary routing, got header %q", got)
	}
}

func TestCanaryFullWeightAlwaysCanary(t *testing.T) {
	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	canary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := CanaryConfig{Weight: 100, HeaderOverride: ""}
	h := newTestCanaryMiddleware(t, cfg, primary, canary)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 from canary, got %d", rec.Code)
	}
}

func TestCanaryNoURLPassesThrough(t *testing.T) {
	called := false
	primary := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
	})

	h, err := NewCanaryMiddleware(CanaryConfig{Weight: 100}, primary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !called {
		t.Error("expected primary handler to be called when no canary URL is set")
	}
}

func TestCanaryDefaultConfig(t *testing.T) {
	cfg := DefaultCanaryConfig()
	if cfg.Weight != 10 {
		t.Errorf("expected default weight 10, got %d", cfg.Weight)
	}
	if cfg.HeaderOverride != "X-Canary" {
		t.Errorf("expected default header X-Canary, got %q", cfg.HeaderOverride)
	}
}
