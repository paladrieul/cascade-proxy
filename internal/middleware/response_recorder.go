package middleware

import (
	"bytes"
	"net/http"
)

// ResponseRecorder wraps http.ResponseWriter to capture status code,
// response headers, and the written body for inspection or caching.
type ResponseRecorder struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
	wrote  bool
}

// NewResponseRecorder wraps w in a ResponseRecorder.
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w}
}

// WriteHeader captures the status code and forwards it to the underlying writer.
func (r *ResponseRecorder) WriteHeader(code int) {
	if !r.wrote {
		r.status = code
		r.wrote = true
		r.ResponseWriter.WriteHeader(code)
	}
}

// Write captures the body bytes and forwards them to the underlying writer.
func (r *ResponseRecorder) Write(b []byte) (int, error) {
	if !r.wrote {
		r.WriteHeader(http.StatusOK)
	}
	r.buf.Write(b) //nolint:errcheck
	return r.ResponseWriter.Write(b)
}

// Status returns the captured HTTP status code.
// Returns 200 if WriteHeader was never called.
func (r *ResponseRecorder) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

// Body returns the captured response body.
func (r *ResponseRecorder) Body() []byte {
	return r.buf.Bytes()
}

// BytesWritten returns the number of body bytes written.
func (r *ResponseRecorder) BytesWritten() int {
	return r.buf.Len()
}
