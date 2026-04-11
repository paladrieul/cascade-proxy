package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestFailoverMiddleware(backends []string, statuses []int, handler http.Handler) http.Handler {
	cfg := DefaultFailoverConfig()
	cfg.Backends = backends
	if statuses != nil {
		cfg.RetryableStatuses = statuses
	}
	return NewFailoverMiddleware(cfg, handler)
}

func TestFailoverUsePrimaryOnSuccess(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("primary"))
	}))
	defer primary.Close()

	h := newTestFailoverMiddleware([]string{primary.URL}, nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(r.URL.String())
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, primary.URL+"/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestFailoverCascadesToSecondBackend(t *testing.T) {
	attempt := 0
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv1.Close()
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv2.Close()

	_ = inner
	cfg := DefaultFailoverConfig()
	cfg.Backends = []string{srv1.URL, srv2.URL}

	h := NewFailoverMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(r.URL.String())
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 after failover, got %d", rec.Code)
	}
}

func TestFailoverReturns502WhenAllBackendsFail(t *testing.T) {
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer down.Close()

	cfg := DefaultFailoverConfig()
	cfg.Backends = []string{down.URL, down.URL}

	h := NewFailoverMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(r.URL.String())
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}

func TestFailoverNonRetryableStatusIsNotCascaded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	cfg := DefaultFailoverConfig()
	cfg.Backends = []string{srv.URL, srv.URL}

	h := NewFailoverMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(r.URL.String())
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	// 403 is not retryable, so it should be returned immediately.
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
