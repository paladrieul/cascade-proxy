package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestBasicAuthMiddleware(creds map[string]string) http.Handler {
	cfg := DefaultBasicAuthConfig()
	cfg.Credentials = creds
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return NewBasicAuthMiddleware(cfg)(handler)
}

func basicAuthHeader(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

func TestBasicAuthAllowsValidCredentials(t *testing.T) {
	h := newTestBasicAuthMiddleware(map[string]string{"alice": "secret"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", basicAuthHeader("alice", "secret"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBasicAuthRejects401OnMissingHeader(t *testing.T) {
	h := newTestBasicAuthMiddleware(map[string]string{"alice": "secret"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("expected WWW-Authenticate header to be set")
	}
}

func TestBasicAuthRejects401OnWrongPassword(t *testing.T) {
	h := newTestBasicAuthMiddleware(map[string]string{"alice": "secret"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", basicAuthHeader("alice", "wrong"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBasicAuthRejects401OnUnknownUser(t *testing.T) {
	h := newTestBasicAuthMiddleware(map[string]string{"alice": "secret"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", basicAuthHeader("bob", "secret"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestBasicAuthDefaultConfig(t *testing.T) {
	cfg := DefaultBasicAuthConfig()
	if cfg.Realm != "Restricted" {
		t.Fatalf("expected realm 'Restricted', got %q", cfg.Realm)
	}
	if cfg.Credentials == nil {
		t.Fatal("expected non-nil credentials map")
	}
}

func TestBasicAuthSetsWWWAuthenticateRealm(t *testing.T) {
	cfg := DefaultBasicAuthConfig()
	cfg.Realm = "MyApp"
	cfg.Credentials = map[string]string{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := NewBasicAuthMiddleware(cfg)(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	want := `Basic realm="MyApp"`
	if got := rec.Header().Get("WWW-Authenticate"); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
