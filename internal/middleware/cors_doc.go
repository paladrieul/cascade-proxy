// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # CORS Middleware
//
// The CORS middleware adds Cross-Origin Resource Sharing headers to HTTP
// responses, enabling browsers to make cross-origin requests to the proxy.
//
// Usage:
//
//	cfg := middleware.DefaultCORSConfig()
//	// Restrict to specific origins in production:
//	cfg.AllowedOrigins = []string{"https://app.example.com"}
//
//	handler := middleware.NewCORSMiddleware(cfg)(nextHandler)
//
// Preflight requests (OPTIONS) are handled automatically and return 204
// No Content with the appropriate Allow-Methods and Allow-Headers headers.
//
// When AllowedOrigins contains "*", the wildcard is sent directly.
// When specific origins are listed, the Vary: Origin header is added so
// that caches can store separate responses per origin.
package middleware
