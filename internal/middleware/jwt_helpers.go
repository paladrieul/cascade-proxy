package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// hmacSHA256B64 computes the HMAC-SHA256 of data using key and returns the
// result as a base64url-encoded string (no padding), matching the JWT spec.
func hmacSHA256B64(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// secureEqual compares two strings in constant time to prevent timing attacks.
func secureEqual(a, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}

// errJWT wraps a descriptive message as a sentinel JWT error.
func errJWT(msg string) error {
	return errors.New(msg)
}

// buildJWT constructs a signed HS256 JWT from the given payload map.
// Intended for use in tests only — not for production token issuance.
func buildJWT(payload map[string]any, secret string) (string, error) {
	headerJSON := `{"alg":"HS256","typ":"JWT"}`
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))

	import_json, err := jsonMarshal(payload)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(import_json)

	sigInput := headerB64 + "." + payloadB64
	sig := hmacSHA256B64(sigInput, secret)
	return sigInput + "." + sig, nil
}
