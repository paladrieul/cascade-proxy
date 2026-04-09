package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestMetricsMiddleware() *MetricsMiddleware {
	return NewMetricsMiddleware()
}

func TestMetricsCountsTotalRequests(t *testing.T) {
	mm := newTestMetricsMiddleware()
	handler := mm.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	if mm.Metrics.TotalRequests != 5 {
		t.Fatalf("expected 5 total requests, got %d", mm.Metrics.TotalRequests)
	}
}

func TestMetricsCountsErrors(t *testing.T) {
	mm := newTestMetricsMiddleware()
	handler := mm.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	if mm.Metrics.TotalErrors != 3 {
		t.Fatalf("expected 3 errors, got %d", mm.Metrics.TotalErrors)
	}
}

func TestMetricsTracksStatusCodes(t *testing.T) {
	mm := newTestMetricsMiddleware()
	statuses := []int{200, 200, 404, 500}
	idx := 0
	handler := mm.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statuses[idx])
		idx++
	}))

	for range statuses {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	if mm.Metrics.StatusCounts[200] != 2 {
		t.Errorf("expected 2 x 200, got %d", mm.Metrics.StatusCounts[200])
	}
	if mm.Metrics.StatusCounts[404] != 1 {
		t.Errorf("expected 1 x 404, got %d", mm.Metrics.StatusCounts[404])
	}
}

func TestMetricsRecordsLatency(t *testing.T) {
	mm := newTestMetricsMiddleware()
	handler := mm.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if mm.Metrics.TotalDuration <= 0 {
		t.Error("expected positive total duration")
	}
}

func TestMetricsHandlerReturnsJSON(t *testing.T) {
	mm := newTestMetricsMiddleware()
	handler := mm.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	rec := httptest.NewRecorder()
	mm.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := out["total_requests"]; !ok {
		t.Error("missing total_requests field")
	}
}
