// Package middleware provides a collection of composable HTTP middleware
// components for the cascade-proxy server.
//
// # Redirect Middleware
//
// NewRedirectMiddleware evaluates an ordered list of RedirectRule entries
// against each incoming request path. The first matching rule issues an HTTP
// redirect to the configured target URL and short-circuits the handler chain.
//
// Rules support two matching modes:
//
//   - Exact match — the request path must equal Rule.From exactly.
//   - Wildcard suffix — when Rule.From ends with "/*" any request whose path
//     starts with the prefix matches. The sub-path after the prefix is
//     appended to Rule.To so that /v1/users/42 → /v2/users/42.
//
// The redirect status code defaults to 302 (Found) when Rule.Code is zero.
// Query strings are always forwarded to the redirect target.
//
// Example:
//
//	cfg := middleware.RedirectConfig{
//	    Rules: []middleware.RedirectRule{
//	        {From: "/v1/*", To: "/v2", Code: http.StatusMovedPermanently},
//	        {From: "/legacy", To: "/home", Code: http.StatusFound},
//	    },
//	}
//	handler := middleware.NewRedirectMiddleware(cfg, next)
package middleware
