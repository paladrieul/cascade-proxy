// Package middleware provides HTTP middleware components for cascade-proxy.
//
// # Body Size Limiting
//
// NewBodySizeMiddleware enforces a maximum request body size, protecting
// upstream services from unexpectedly large payloads.
//
// Requests are rejected with HTTP 413 Request Entity Too Large in two cases:
//
//  1. The Content-Length header exceeds the configured limit — the request is
//     rejected immediately before the body is read.
//
//  2. The body is read and its actual size exceeds the limit — http.MaxBytesReader
//     is used so the limit is enforced lazily during reading.
//
// # Usage
//
//	cfg := middleware.DefaultBodySizeConfig() // 1 MB
//	cfg.MaxBytes = 512 * 1024                 // override to 512 KB
//	mux.Use(middleware.NewBodySizeMiddleware(cfg))
package middleware
