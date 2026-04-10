// Package middleware provides a collection of composable HTTP middleware
// components for cascade-proxy.
//
// # Rewrite Middleware
//
// NewRewriteMiddleware rewrites incoming request paths before they are
// forwarded to the upstream backend. It supports two complementary
// mechanisms:
//
//  1. Regexp Rules – an ordered list of [RewriteRule] values. Each rule
//     pairs a compiled *regexp.Regexp with a replacement string that
//     follows the same syntax as [regexp.Regexp.ReplaceAllString]. Rules
//     are evaluated in declaration order and only the first matching rule
//     is applied.
//
//  2. StripPrefix – a plain string prefix that is removed from the path
//     after rule matching. If stripping leaves an empty path it is
//     normalised to "/".
//
// Both mechanisms can be combined: rules run first, then prefix stripping.
//
// Example:
//
//	cfg := middleware.RewriteConfig{
//		Rules: []middleware.RewriteRule{
//			{Pattern: regexp.MustCompile(`^/v1/(.*)`), Replacement: "/api/$1"},
//		},
//		StripPrefix: "/api",
//	}
//	handler = middleware.NewRewriteMiddleware(cfg, next)
package middleware
