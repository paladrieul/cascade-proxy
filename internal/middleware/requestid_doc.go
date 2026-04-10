// Package middleware provides a collection of HTTP middleware components for
// cascade-proxy.
//
// # Request ID Middleware
//
// NewRequestIDMiddleware ensures every request carries a unique identifier
// throughout its lifecycle.
//
// Behaviour:
//
//   - If the incoming request already contains an X-Request-ID header its
//     value is reused, enabling end-to-end correlation across services.
//   - Otherwise a new UUID v4 is generated for the request.
//   - The ID is forwarded to the upstream backend via the X-Request-ID request
//     header so it appears in upstream access logs.
//   - The ID is echoed back to the caller in the X-Request-ID response header.
//   - The ID is stored in the request context and can be retrieved by
//     downstream handlers using RequestIDFromContext.
//
// Usage:
//
//	handler = middleware.NewRequestIDMiddleware(handler)
package middleware
