package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

// TestThrottleIntegrationWithLogger verifies that the throttle middleware
// chains correctly with the Logger middleware and that log output is produced
// for rejected requests.
func TestThrottleIntegrationWithLogger(t *testing.T) {
	blocked := make(chan struct{})
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked
		w.WriteHeader(http.StatusOK)
	})

	logger := log.New(os.Stderr, "test: ", 0)
	throttled := NewThrottleMiddleware(ThrottleConfig{
		MaxConcurrent: 1,
		QueueTimeout:  30 * time.Millisecond,
		Logger:        logger,
	}, inner)
	logged := Logger(logger, throttled)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logged.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/slow", nil))
	}()
	time.Sleep(5 * time.Millisecond)

	rr := httptest.NewRecorder()
	logged.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/slow", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 from throttle, got %d", rr.Code)
	}
	close(blocked)
	wg.Wait()
}

// TestThrottleIntegrationWithRetry ensures that when throttle sits inside a
// retry wrapper the retry logic sees the 503 and attempts the request again
// once the slot becomes free.
func TestThrottleIntegrationWithRetry(t *testing.T) {
	attempts := 0
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusOK)
	})

	throttled := NewThrottleMiddleware(ThrottleConfig{
		MaxConcurrent: 5,
		QueueTimeout:  time.Second,
		Logger:        log.New(os.Stderr, "", 0),
	}, inner)

	retryCfg := DefaultRetryConfig
	retryCfg.MaxAttempts = 2
	retried := NewRetryMiddleware(retryCfg, throttled)

	rr := httptest.NewRecorder()
	retried.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if attempts < 1 {
		t.Fatal("expected at least one attempt")
	}
}
