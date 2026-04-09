package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds aggregated request statistics.
type Metrics struct {
	TotalRequests  uint64
	TotalErrors    uint64
	TotalDuration  int64 // nanoseconds
	StatusCounts   map[int]uint64
	mu             sync.RWMutex
}

func newMetrics() *Metrics {
	return &Metrics{StatusCounts: make(map[int]uint64)}
}

func (m *Metrics) record(status int, duration time.Duration) {
	atomic.AddUint64(&m.TotalRequests, 1)
	atomic.AddInt64(&m.TotalDuration, int64(duration))
	if status >= 500 {
		atomic.AddUint64(&m.TotalErrors, 1)
	}
	m.mu.Lock()
	m.StatusCounts[status]++
	m.mu.Unlock()
}

// MetricsMiddleware tracks per-request latency and status codes.
type MetricsMiddleware struct {
	Metrics *Metrics
}

// NewMetricsMiddleware returns a middleware that collects request metrics.
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{Metrics: newMetrics()}
}

// Handler wraps next and records metrics for every request.
func (m *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := NewResponseRecorder(w)
		start := time.Now()
		next.ServeHTTP(rec, r)
		m.Metrics.record(rec.Status, time.Since(start))
	})
}

// ServeHTTP exposes a simple JSON metrics snapshot.
func (m *MetricsMiddleware) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	m.Metrics.mu.RLock()
	defer m.Metrics.mu.RUnlock()

	total := atomic.LoadUint64(&m.Metrics.TotalRequests)
	errors := atomic.LoadUint64(&m.Metrics.TotalErrors)
	durNs := atomic.LoadInt64(&m.Metrics.TotalDuration)

	var avgMs float64
	if total > 0 {
		avgMs = float64(durNs) / float64(total) / 1e6
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	body := `{"total_requests":` + strconv.FormatUint(total, 10) +
		`,"total_errors":` + strconv.FormatUint(errors, 10) +
		`,"avg_latency_ms":` + strconv.FormatFloat(avgMs, 'f', 3, 64) + `}`
	_, _ = w.Write([]byte(body))
}
