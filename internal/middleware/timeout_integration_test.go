package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/casualjim/cascade-proxy/internal/middleware"
)

func TestTimeoutIntegrationWithRecovery(t *testing.T) {
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	cfg := middleware.DefaultTimeoutConfig()
	cfg.Timeout = 50 * time.Millisecond

	handler := middleware.NewRecoveryMiddleware(
		middleware.NewTimeoutMiddleware(cfg, slow),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", rec.Code)
	}
}

func TestTimeoutIntegrationWithLogger(t *testing.T) {
	fast := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := middleware.DefaultTimeoutConfig()
	cfg.Timeout = 100 * time.Millisecond

	var logged bool
	logger := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logged = true
			next.ServeHTTP(w, r)
		})
	}

	handler := logger(middleware.NewTimeoutMiddleware(cfg, fast))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !logged {
		t.Error("expected logger to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
