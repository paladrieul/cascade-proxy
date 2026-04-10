package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newTestIPFilter(cfg IPFilterConfig) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	cfg.Logger = log.New(os.Stdout, "", 0)
	return NewIPFilterMiddleware(cfg, next)
}

func TestIPFilterAllowsRequestWithNoRules(t *testing.T) {
	h := newTestIPFilter(DefaultIPFilterConfig())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.10:1234"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestIPFilterBlocksExplicitlyBlockedCIDR(t *testing.T) {
	cfg := DefaultIPFilterConfig()
	cfg.BlockedCIDRs = []string{"10.0.0.0/8"}
	h := newTestIPFilter(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestIPFilterAllowsRequestInAllowedCIDR(t *testing.T) {
	cfg := DefaultIPFilterConfig()
	cfg.AllowedCIDRs = []string{"192.168.0.0/16"}
	h := newTestIPFilter(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.5.5:9000"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestIPFilterDeniesRequestOutsideAllowedCIDR(t *testing.T) {
	cfg := DefaultIPFilterConfig()
	cfg.AllowedCIDRs = []string{"192.168.0.0/16"}
	h := newTestIPFilter(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestIPFilterBlockedTakesPrecedenceOverAllowed(t *testing.T) {
	cfg := DefaultIPFilterConfig()
	cfg.AllowedCIDRs = []string{"172.16.0.0/12"}
	cfg.BlockedCIDRs = []string{"172.16.1.0/24"}
	h := newTestIPFilter(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.16.1.50:8080"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 (blocked takes precedence), got %d", rec.Code)
	}
}

func TestIPFilterUsesXForwardedFor(t *testing.T) {
	cfg := DefaultIPFilterConfig()
	cfg.BlockedCIDRs = []string{"203.0.113.0/24"}
	h := newTestIPFilter(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.42, 10.0.0.1")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 based on X-Forwarded-For, got %d", rec.Code)
	}
}
