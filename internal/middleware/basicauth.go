package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

// BasicAuthConfig holds configuration for HTTP Basic Authentication middleware.
type BasicAuthConfig struct {
	// Credentials is a map of username to password.
	Credentials map[string]string
	// Realm is the authentication realm returned in WWW-Authenticate header.
	Realm string
}

// DefaultBasicAuthConfig returns a BasicAuthConfig with sensible defaults.
func DefaultBasicAuthConfig() BasicAuthConfig {
	return BasicAuthConfig{
		Credentials: map[string]string{},
		Realm:       "Restricted",
	}
}

// NewBasicAuthMiddleware returns an HTTP middleware that enforces HTTP Basic
// Authentication. Requests without valid credentials receive a 401 response
// with a WWW-Authenticate challenge header.
func NewBasicAuthMiddleware(cfg BasicAuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := parseBasicAuth(r)
			if !ok || !validateCredentials(cfg.Credentials, user, pass) {
				w.Header().Set("WWW-Authenticate", `Basic realm="`+cfg.Realm+`"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// parseBasicAuth extracts username and password from the Authorization header.
func parseBasicAuth(r *http.Request) (string, string, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Basic ") {
		return "", "", false
	}
	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// validateCredentials checks the provided username/password against the
// configured credentials map using constant-time comparison.
func validateCredentials(creds map[string]string, user, pass string) bool {
	expected, ok := creds[user]
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(pass), []byte(expected)) == 1
}
