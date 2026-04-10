package middleware

import (
	"bytes"
	"net/http"
)

// ResponseRecorder buffers the status code, headers, and body so that the
// response can be inspected or discarded before being forwarded to the real
// http.ResponseWriter.
type ResponseRecorder struct {
	Status  int
	headers http.Header
	body    bytes.Buffer
	wroteHeader bool
}

// NewResponseRecorder returns a new ResponseRecorder. The underlying
// http.ResponseWriter is NOT written to until Flush is called.
func NewResponseRecorder(_ http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		Status:  http.StatusOK,
		headers: make(http.Header),
	}
}

func (r *ResponseRecorder) Header() http.Header {
	return r.headers
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	if !r.wroteHeader {
		r.Status = statusCode
		r.wroteHeader = true
	}
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.wroteHeader = true
	}
	return r.body.Write(b)
}

// Flush copies the buffered response to the real http.ResponseWriter.
func (r *ResponseRecorder) Flush(w http.ResponseWriter) {
	for k, vals := range r.headers {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(r.Status)
	_, _ = w.Write(r.body.Bytes())
}
