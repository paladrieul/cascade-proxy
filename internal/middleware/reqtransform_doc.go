// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Request Transform Middleware
//
// NewRequestTransformMiddleware rewrites the body of incoming HTTP requests
// before they are forwarded to the upstream backend. It applies a list of
// find-and-replace rules in order, optionally scoped to a specific
// Content-Type.
//
// # Usage
//
//	cfg := middleware.RequestTransformConfig{
//		Rules: []middleware.RequestTransformRule{
//			{
//				Find:        "staging.internal",
//				Replace:     "prod.internal",
//				ContentType: "application/json",
//			},
//		},
//	}
//	handler := middleware.NewRequestTransformMiddleware(cfg, next)
//
// Rules with an empty ContentType field are applied regardless of the
// request's Content-Type header. Rules are applied in the order they are
// defined; later rules operate on the already-modified body.
package middleware
