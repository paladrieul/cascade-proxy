// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Response Transform Middleware
//
// NewResponseTransformMiddleware rewrites response bodies using a list of
// regular-expression find-and-replace rules before the response is sent to
// the client.
//
// Rules are applied in declaration order, so later rules operate on the
// output of earlier ones.  An optional ContentTypes filter restricts
// transformation to responses whose Content-Type header contains one of
// the specified substrings; leave the slice empty to transform all types.
//
// Example usage:
//
//	cfg := middleware.ResponseTransformConfig{
//	    Rules: []middleware.TransformRule{
//	        {
//	            Pattern:     regexp.MustCompile(`https://internal\.svc`),
//	            Replacement: "https://public.example.com",
//	        },
//	    },
//	    ContentTypes: []string{"application/json", "text/html"},
//	}
//	handler = middleware.NewResponseTransformMiddleware(cfg, handler)
package middleware
