package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const TraceIDKey contextKey = "trace_id"

// TraceIDHeader is the HTTP header used to propagate trace IDs.
const TraceIDHeader = "X-Trace-ID"

// NewTracingMiddleware injects a trace ID into every request context and
// response header. If the incoming request already carries an X-Trace-ID
// header that value is reused, otherwise a new UUID is generated.
func NewTracingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		r = r.WithContext(ctx)

		w.Header().Set(TraceIDHeader, traceID)

		start := time.Now()
		rec := NewBufferedResponseRecorder(w)
		next.ServeHTTP(rec, r)

		logger.Info("request traced",
			"trace_id", traceID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// TraceIDFromContext retrieves the trace ID stored in ctx, returning an empty
// string when none is present.
func TraceIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(TraceIDKey).(string)
	return v
}

// NewBufferedResponseRecorder wraps w so that the status code written by the
// downstream handler can be inspected after the fact.
func NewBufferedResponseRecorder(w http.ResponseWriter) *bufferedRecorder {
	return &bufferedRecorder{ResponseWriter: w, status: http.StatusOK}
}

type bufferedRecorder struct {
	http.ResponseWriter
	status int
}

func (b *bufferedRecorder) WriteHeader(code int) {
	b.status = code
	b.ResponseWriter.WriteHeader(code)
}

func (b *bufferedRecorder) Status() int { return b.status }
