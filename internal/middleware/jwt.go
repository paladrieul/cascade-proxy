package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type jwtContextKey struct{}

// JWTClaims holds the parsed claims from a validated JWT token.
type JWTClaims map[string]any

// DefaultJWTConfig returns a JWTConfig with sensible defaults.
func DefaultJWTConfig(secret string) JWTConfig {
	return JWTConfig{
		Secret:        secret,
		Header:        "Authorization",
		Scheme:        "Bearer",
		ClaimsContext: true,
	}
}

// JWTConfig configures the JWT validation middleware.
type JWTConfig struct {
	Secret        string
	Header        string
	Scheme        string
	ClaimsContext bool
	Logger        *slog.Logger
}

// NewJWTMiddleware validates incoming JWT tokens using HS256 (HMAC-SHA256).
// Requests with missing or invalid tokens are rejected with 401.
func NewJWTMiddleware(cfg JWTConfig, next http.Handler) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get(cfg.Header)
		if raw == "" {
			cfg.Logger.Warn("jwt: missing token", "path", r.URL.Path)
			http.Error(w, "unauthorized: missing token", http.StatusUnauthorized)
			return
		}
		token := raw
		if cfg.Scheme != "" {
			prefix := cfg.Scheme + " "
			if !strings.HasPrefix(raw, prefix) {
				cfg.Logger.Warn("jwt: invalid scheme", "path", r.URL.Path)
				http.Error(w, "unauthorized: invalid scheme", http.StatusUnauthorized)
				return
			}
			token = strings.TrimPrefix(raw, prefix)
		}
		claims, err := parseJWT(token, cfg.Secret)
		if err != nil {
			cfg.Logger.Warn("jwt: invalid token", "path", r.URL.Path, "err", err)
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		if cfg.ClaimsContext {
			r = r.WithContext(context.WithValue(r.Context(), jwtContextKey{}, claims))
		}
		next.ServeHTTP(w, r)
	})
}

// JWTClaimsFromContext retrieves parsed JWT claims stored in the request context.
func JWTClaimsFromContext(ctx context.Context) JWTClaims {
	v, _ := ctx.Value(jwtContextKey{}).(JWTClaims)
	return v
}

// parseJWT performs minimal HS256 JWT validation: header.payload.signature.
// It decodes the payload, verifies the signature, and checks exp/nbf claims.
func parseJWT(token, secret string) (JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errJWT("malformed token")
	}
	sigInput := parts[0] + "." + parts[1]
	expected := hmacSHA256B64(sigInput, secret)
	if !secureEqual(expected, parts[2]) {
		return nil, errJWT("invalid signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errJWT("malformed payload")
	}
	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errJWT("invalid payload JSON")
	}
	now := float64(time.Now().Unix())
	if exp, ok := claims["exp"].(float64); ok && now > exp {
		return nil, errJWT("token expired")
	}
	if nbf, ok := claims["nbf"].(float64); ok && now < nbf {
		return nil, errJWT("token not yet valid")
	}
	return claims, nil
}
