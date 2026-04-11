package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTestThrottleMiddleware(maxConcurrent int, queueTimeout time.Duration, next http.Handler) http.Handler {
	cfg := ThrottleConfig{
		MaxConcurrent: maxConcurrent,
		QueueTimeout:  queueTimeout,
		Logger:        log.New(os.Stderr, "throttle: ", 0),
	}
	return NewThrottleMiddleware(cfg, next)
}

func TestThrottleAllowsRequestsWithinLimit(t *testing.T) {
	handler := newTestThrottleMiddleware(5, time.Second, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestThrottleRejects503WhenQueueTimesOut(t *testing.T) {
	blocked := make(chan struct{})
	handler := newTestThrottleMiddleware(1, 50*time.Millisecond, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	time.Sleep(10 * time.Millisecond)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
	close(blocked)
	wg.Wait()
}

func TestThrottleLimitsConcurrency(t *testing.T) {
	const max = 3
	var active int64
	var exceeded int64

	handler := newTestThrottleMiddleware(max, time.Second, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt64(&active, 1)
		defer atomic.AddInt64(&active, -1)
		if current > max {
			atomic.AddInt64(&exceeded, 1)
		}
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()
	if exceeded > 0 {
		t.Fatalf("concurrency limit exceeded: %d requests ran simultaneously above max", exceeded)
	}
}

func TestThrottleDefaultConfig(t *testing.T) {
	cfg := DefaultThrottleConfig(nil)
	if cfg.MaxConcurrent != 10 {
		t.Fatalf("expected MaxConcurrent=10, got %d", cfg.MaxConcurrent)
	}
	if cfg.QueueTimeout != 5*time.Second {
		t.Fatalf("expected QueueTimeout=5s, got %v", cfg.QueueTimeout)
	}
}
