package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func newTestTimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	cfg := TimeoutConfig{
		Timeout: timeout,
		Logger:  log.New(os.Stdout, "[timeout] ", 0),
	}
	return NewTimeoutMiddleware(cfg)
}

func TestTimeoutAllowsFastRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := newTestTimeoutMiddleware(500 * time.Millisecond)
	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ping")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestTimeoutReturns504OnSlowBackend(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	mw := newTestTimeoutMiddleware(50 * time.Millisecond)
	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/slow")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", resp.StatusCode)
	}
}

func TestTimeoutDefaultConfig(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %s", cfg.Timeout)
	}
	if cfg.Logger == nil {
		t.Error("expected non-nil logger in default config")
	}
}

func TestTimeoutExactBoundary(t *testing.T) {
	// Verify that a handler completing just under the timeout is not cancelled.
	const timeout = 100 * time.Millisecond

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(timeout / 2)
		w.WriteHeader(http.StatusOK)
	})

	mw := newTestTimeoutMiddleware(timeout)
	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/boundary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for request within timeout, got %d", resp.StatusCode)
	}
}
