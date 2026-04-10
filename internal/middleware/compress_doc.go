// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Compress Middleware
//
// NewCompressMiddleware wraps a handler and transparently gzip-compresses
// response bodies when the downstream client advertises support through the
// "Accept-Encoding: gzip" request header.
//
// Compression is skipped for responses whose buffered body is smaller than
// CompressConfig.MinLength (default 1 KiB) to avoid the overhead of
// compressing tiny payloads that would not benefit from it.
//
// The middleware:
//   - Pools gzip.Writer instances to reduce allocations under load.
//   - Removes the Content-Length header when compression is applied so that
//     the client does not receive a length that no longer matches the
//     compressed stream.
//   - Preserves the original HTTP status code.
//
// Example:
//
//	cfg := middleware.DefaultCompressConfig
//	h   := middleware.NewCompressMiddleware(cfg, proxyHandler)
//	http.ListenAndServe(":8080", h)
package middleware
