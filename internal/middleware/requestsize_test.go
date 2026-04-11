package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joeydtaylor/cascade-proxy/internal/middleware"
)

func newTestRequestSizeMiddleware(cfg middleware.RequestSizeConfig) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return middleware.NewRequestSizeMiddleware(cfg, next)
}

func TestRequestSizeAllowsNormalRequest(t *testing.T) {
	cfg := middleware.DefaultRequestSizeConfig()
	cfg.Logger = slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := newTestRequestSizeMiddleware(cfg)

	req := httptest.NewRequest(http.MethodGet, "/hello?foo=bar", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestSizeRejects414OnLongURI(t *testing.T) {
	cfg := middleware.DefaultRequestSizeConfig()
	cfg.MaxURLLength = 20
	cfg.Logger = slog.Default()
	handler := newTestRequestSizeMiddleware(cfg)

	long := "/" + strings.Repeat("a", 50)
	req := httptest.NewRequest(http.MethodGet, long, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestURITooLong {
		t.Fatalf("expected 414, got %d", rec.Code)
	}
}

func TestRequestSizeRejects400OnTooManyQueryParams(t *testing.T) {
	cfg := middleware.DefaultRequestSizeConfig()
	cfg.MaxQueryParams = 2
	cfg.Logger = slog.Default()
	handler := newTestRequestSizeMiddleware(cfg)

	req := httptest.NewRequest(http.MethodGet, "/?a=1&b=2&c=3", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRequestSizeRejects431OnLargeHeaders(t *testing.T) {
	cfg := middleware.DefaultRequestSizeConfig()
	cfg.MaxHeaderBytes = 50
	cfg.Logger = slog.Default()
	handler := newTestRequestSizeMiddleware(cfg)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Big-Header", strings.Repeat("x", 100))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestHeaderFieldsTooLarge {
		t.Fatalf("expected 431, got %d", rec.Code)
	}
}

func TestRequestSizeDefaultConfig(t *testing.T) {
	cfg := middleware.DefaultRequestSizeConfig()
	if cfg.MaxHeaderBytes <= 0 {
		t.Error("expected positive MaxHeaderBytes")
	}
	if cfg.MaxURLLength <= 0 {
		t.Error("expected positive MaxURLLength")
	}
	if cfg.MaxQueryParams <= 0 {
		t.Error("expected positive MaxQueryParams")
	}
}
