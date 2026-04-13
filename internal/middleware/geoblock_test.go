package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newTestGeoBlockMiddleware(cfg GeoBlockConfig) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	return NewGeoBlockMiddleware(cfg, next)
}

func TestGeoBlockAllowsRequestWithNoCountryHeader(t *testing.T) {
	cfg := DefaultGeoBlockConfig()
	cfg.BlockedCountries = []string{"CN", "RU"}
	h := newTestGeoBlockMiddleware(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGeoBlockBlocksCountryInDenylist(t *testing.T) {
	cfg := DefaultGeoBlockConfig()
	cfg.BlockedCountries = []string{"CN", "RU"}
	h := newTestGeoBlockMiddleware(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "CN")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGeoBlockAllowsCountryNotInDenylist(t *testing.T) {
	cfg := DefaultGeoBlockConfig()
	cfg.BlockedCountries = []string{"CN"}
	h := newTestGeoBlockMiddleware(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "DE")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGeoBlockAllowlistDeniesUnlistedCountry(t *testing.T) {
	cfg := DefaultGeoBlockConfig()
	cfg.AllowedCountries = []string{"US", "GB"}
	h := newTestGeoBlockMiddleware(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "FR")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGeoBlockAllowlistPermitsListedCountry(t *testing.T) {
	cfg := DefaultGeoBlockConfig()
	cfg.AllowedCountries = []string{"US", "GB"}
	h := newTestGeoBlockMiddleware(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "us") // lowercase should still match
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGeoBlockCustomDenyStatus(t *testing.T) {
	cfg := DefaultGeoBlockConfig()
	cfg.BlockedCountries = []string{"KP"}
	cfg.DenyStatus = http.StatusUnavailableForLegalReasons
	h := newTestGeoBlockMiddleware(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Country-Code", "KP")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnavailableForLegalReasons {
		t.Fatalf("expected 451, got %d", rec.Code)
	}
}
