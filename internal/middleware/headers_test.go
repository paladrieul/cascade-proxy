package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestHeadersMiddleware(cfg HeadersConfig, handler http.Handler) http.Handler {
	if handler == nil {
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	}
	return NewHeadersMiddleware(cfg, handler)
}

func TestHeadersAddsRequestHeader(t *testing.T) {
	var capturedHeader string
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Custom")
		w.WriteHeader(http.StatusOK)
	})

	cfg := DefaultHeadersConfig()
	cfg.RequestHeaders["X-Custom"] = "injected"
	h := newTestHeadersMiddleware(cfg, backend)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if capturedHeader != "injected" {
		t.Errorf("expected X-Custom=injected, got %q", capturedHeader)
	}
}

func TestHeadersRemovesRequestHeader(t *testing.T) {
	var capturedHeader string
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})

	cfg := DefaultHeadersConfig()
	cfg.RemoveRequestHeaders = []string{"Authorization"}
	h := newTestHeadersMiddleware(cfg, backend)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if capturedHeader != "" {
		t.Errorf("expected Authorization to be removed, got %q", capturedHeader)
	}
}

func TestHeadersAddsResponseHeader(t *testing.T) {
	cfg := DefaultHeadersConfig()
	cfg.ResponseHeaders["X-Powered-By"] = "cascade-proxy"
	h := newTestHeadersMiddleware(cfg, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if got := rw.Header().Get("X-Powered-By"); got != "cascade-proxy" {
		t.Errorf("expected X-Powered-By=cascade-proxy, got %q", got)
	}
}

func TestHeadersRemovesResponseHeader(t *testing.T) {
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Internal-Secret", "hidden")
		w.WriteHeader(http.StatusOK)
	})

	cfg := DefaultHeadersConfig()
	cfg.RemoveResponseHeaders = []string{"X-Internal-Secret"}
	h := newTestHeadersMiddleware(cfg, backend)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if got := rw.Header().Get("X-Internal-Secret"); got != "" {
		t.Errorf("expected X-Internal-Secret to be removed, got %q", got)
	}
}

func TestHeadersDoesNotMutateOriginalRequest(t *testing.T) {
	cfg := DefaultHeadersConfig()
	cfg.RequestHeaders["X-Added"] = "yes"
	h := newTestHeadersMiddleware(cfg, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if req.Header.Get("X-Added") != "" {
		t.Error("original request should not be mutated")
	}
}
