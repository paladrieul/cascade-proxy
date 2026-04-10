package middleware

import (
	"net/http"
	"strings"
)

// HeadersConfig defines headers to add, set, or remove from requests and responses.
type HeadersConfig struct {
	// RequestHeaders are added/overwritten on the upstream request.
	RequestHeaders map[string]string
	// ResponseHeaders are added/overwritten on the downstream response.
	ResponseHeaders map[string]string
	// RemoveRequestHeaders lists headers stripped from the upstream request.
	RemoveRequestHeaders []string
	// RemoveResponseHeaders lists headers stripped from the downstream response.
	RemoveResponseHeaders []string
}

// DefaultHeadersConfig returns an empty HeadersConfig.
func DefaultHeadersConfig() HeadersConfig {
	return HeadersConfig{
		RequestHeaders:        make(map[string]string),
		ResponseHeaders:       make(map[string]string),
		RemoveRequestHeaders:  []string{},
		RemoveResponseHeaders: []string{},
	}
}

// NewHeadersMiddleware returns middleware that manipulates request and response
// headers according to cfg before passing control to next.
func NewHeadersMiddleware(cfg HeadersConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mutate a shallow copy so we don't modify the original request.
		outReq := r.Clone(r.Context())

		for _, key := range cfg.RemoveRequestHeaders {
			outReq.Header.Del(key)
		}
		for key, val := range cfg.RequestHeaders {
			outReq.Header.Set(key, val)
		}

		rw := &headerResponseWriter{
			ResponseWriter:        w,
			addHeaders:            cfg.ResponseHeaders,
			removeHeaders:         cfg.RemoveResponseHeaders,
			headersWritten:        false,
		}

		next.ServeHTTP(rw, outReq)
	})
}

// headerResponseWriter intercepts WriteHeader to inject/remove response headers.
type headerResponseWriter struct {
	http.ResponseWriter
	addHeaders     map[string]string
	removeHeaders  []string
	headersWritten bool
}

func (rw *headerResponseWriter) WriteHeader(code int) {
	if !rw.headersWritten {
		rw.applyHeaders()
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *headerResponseWriter) Write(b []byte) (int, error) {
	if !rw.headersWritten {
		rw.applyHeaders()
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *headerResponseWriter) applyHeaders() {
	rw.headersWritten = true
	h := rw.ResponseWriter.Header()
	for _, key := range rw.removeHeaders {
		h.Del(key)
	}
	for key, val := range rw.addHeaders {
		h.Set(key, val)
	}
}

// canonicalKey normalises a header name for comparison.
func canonicalKey(k string) string {
	return strings.ToLower(strings.TrimSpace(k))
}
