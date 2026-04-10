// Package middleware provides composable HTTP middleware for cascade-proxy.
//
// # Auth Middleware
//
// NewAuthMiddleware enforces bearer-token / API-key authentication on every
// inbound request.  It inspects the configured header (default:
// "Authorization") and compares the extracted value against a set of
// pre-shared keys supplied at construction time.
//
// Usage:
//
//	cfg := middleware.DefaultAuthConfig(
//	    []string{"my-secret-key"},
//	    log.Default(),
//	)
//	handler := middleware.NewAuthMiddleware(cfg, nextHandler)
//
// Both plain API keys and "Bearer <token>" values are accepted so that the
// middleware integrates with standard Authorization header conventions as well
// as custom key headers.
//
// Requests that fail authentication receive an HTTP 401 Unauthorized response
// and are never forwarded to the upstream backend.
package middleware
