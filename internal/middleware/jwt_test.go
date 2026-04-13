package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testJWTSecret = "super-secret-key-for-testing"

func makeToken(t *testing.T, claims map[string]any, secret string) string {
	t.Helper()
	headerJSON := `{"alg":"HS256","typ":"JWT"}`
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	sigInput := headerB64 + "." + payloadB64
	sig := hmacSHA256B64(sigInput, secret)
	return sigInput + "." + sig
}

func newTestJWTMiddleware(secret string) http.Handler {
	cfg := DefaultJWTConfig(secret)
	cfg.Logger = slog.Default()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return NewJWTMiddleware(cfg, next)
}

func TestJWTAllowsValidToken(t *testing.T) {
	claims := map[string]any{"sub": "user-1", "exp": float64(time.Now().Add(time.Hour).Unix())}
	token := makeToken(t, claims, testJWTSecret)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	newTestJWTMiddleware(testJWTSecret).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestJWTRejects401OnMissingHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	newTestJWTMiddleware(testJWTSecret).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestJWTRejects401OnWrongSecret(t *testing.T) {
	claims := map[string]any{"sub": "user-1", "exp": float64(time.Now().Add(time.Hour).Unix())}
	token := makeToken(t, claims, "wrong-secret")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	newTestJWTMiddleware(testJWTSecret).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestJWTRejects401OnExpiredToken(t *testing.T) {
	claims := map[string]any{"sub": "user-1", "exp": float64(time.Now().Add(-time.Hour).Unix())}
	token := makeToken(t, claims, testJWTSecret)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	newTestJWTMiddleware(testJWTSecret).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestJWTStoresClaimsInContext(t *testing.T) {
	claims := map[string]any{"sub": "user-42", "exp": float64(time.Now().Add(time.Hour).Unix())}
	token := makeToken(t, claims, testJWTSecret)

	var gotClaims JWTClaims
	cfg := DefaultJWTConfig(testJWTSecret)
	cfg.Logger = slog.Default()
	handler := NewJWTMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims = JWTClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(rr, req)

	if gotClaims["sub"] != "user-42" {
		t.Fatalf("expected sub=user-42, got %v", gotClaims["sub"])
	}
}

func TestJWTClaimsFromContextReturnsNilWhenMissing(t *testing.T) {
	claims := JWTClaimsFromContext(context.Background())
	if claims != nil {
		t.Fatalf("expected nil claims, got %v", claims)
	}
}
