package healthcheck_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cascade-proxy/internal/healthcheck"
)

func newHealthyBackend() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func newUnhealthyBackend() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
}

func TestProbeReportsHealthyBackend(t *testing.T) {
	backend := newHealthyBackend()
	defer backend.Close()

	checker := healthcheck.New([]string{backend.URL}, 2*time.Second)
	checker.Probe(context.Background())

	statuses := checker.Statuses()
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if !statuses[0].Healthy {
		t.Errorf("expected backend to be healthy")
	}
}

func TestProbeReportsUnhealthyBackend(t *testing.T) {
	backend := newUnhealthyBackend()
	defer backend.Close()

	checker := healthcheck.New([]string{backend.URL}, 2*time.Second)
	checker.Probe(context.Background())

	statuses := checker.Statuses()
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].Healthy {
		t.Errorf("expected backend to be unhealthy")
	}
}

func TestHandlerReturns200WhenAllHealthy(t *testing.T) {
	backend := newHealthyBackend()
	defer backend.Close()

	checker := healthcheck.New([]string{backend.URL}, 2*time.Second)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/__health", nil)
	checker.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["healthy"] != true {
		t.Errorf("expected healthy=true in response body")
	}
}

func TestHandlerReturns503WhenBackendUnhealthy(t *testing.T) {
	backend := newUnhealthyBackend()
	defer backend.Close()

	checker := healthcheck.New([]string{backend.URL}, 2*time.Second)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/__health", nil)
	checker.Handler()(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestProbeUnreachableBackendIsUnhealthy(t *testing.T) {
	checker := healthcheck.New([]string{"http://127.0.0.1:1"}, 200*time.Millisecond)
	checker.Probe(context.Background())

	statuses := checker.Statuses()
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].Healthy {
		t.Errorf("expected unreachable backend to be unhealthy")
	}
}
