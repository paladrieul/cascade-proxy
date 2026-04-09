package middleware

import (
	"log"
	"net/http"
	"time"
)

// ResponseRecorder wraps http.ResponseWriter to capture the status code.
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Written    int64
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w, StatusCode: http.StatusOK}
}

func (r *ResponseRecorder) WriteHeader(code int) {
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.Written += int64(n)
	return n, err
}

// Logger returns an HTTP middleware that logs each request with method, path,
// status code, duration, and bytes written.
func Logger(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := NewResponseRecorder(w)

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			logger.Printf("%s %s %d %d bytes %s",
				r.Method,
				r.URL.Path,
				rec.StatusCode,
				rec.Written,
				duration,
			)
		})
	}
}
