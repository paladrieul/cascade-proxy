package middleware

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cascade-proxy/internal/balancer"
)

func newTestBalancerMiddleware(targets []string) (http.Handler, error) {
	b, err := balancer.New(balancer.Config{Targets: targets})
	if err != nil {
		return nil, err
	}
	logger := log.New(os.Stdout, "[balancer-test] ", 0)
	return NewBalancerMiddleware(b, logger), nil
}

func TestBalancerForwardsToBackend(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()

	h, err := newTestBalancerMiddleware([]string{backend.URL})
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "ok" {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestBalancerRoundRobinAcrossBackends(t *testing.T) {
	hits := make([]int, 2)
	backend0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits[0]++ }))
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits[1]++ }))
	defer backend0.Close()
	defer backend1.Close()

	h, err := newTestBalancerMiddleware([]string{backend0.URL, backend1.URL})
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	for i := 0; i < 4; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)
	}

	if hits[0] != 2 || hits[1] != 2 {
		t.Errorf("expected 2 hits each, got %v", hits)
	}
}

func TestBalancerReturns502WhenBackendDown(t *testing.T) {
	// Use a port that is not listening.
	h, err := newTestBalancerMiddleware([]string{"http://127.0.0.1:19999"})
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}
