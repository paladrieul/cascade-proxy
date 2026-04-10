package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newTestAuthMiddleware(keys []string) http.Handler {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	logger := log.New(os.Stderr, "test-auth: ", 0)
	cfg := DefaultAuthConfig(keys, logger)
	return NewAuthMiddleware(cfg, inner)
}

func TestAuthAllowsValidBearerToken(t *testing.T) {
	h := newTestAuthMiddleware([]string{"secret-token"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthRejects401OnMissingHeader(t *testing.T) {
	h := newTestAuthMiddleware([]string{"secret-token"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthRejects401OnWrongToken(t *testing.T) {
	h := newTestAuthMiddleware([]string{"secret-token"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthAllowsPlainAPIKey(t *testing.T) {
	h := newTestAuthMiddleware([]string{"plain-key"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "plain-key") // no "Bearer " prefix
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthMultipleValidKeys(t *testing.T) {
	h := newTestAuthMiddleware([]string{"key-a", "key-b"})

	for _, key := range []string{"key-a", "key-b"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+key)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("key %q: expected 200, got %d", key, rec.Code)
		}
	}
}
