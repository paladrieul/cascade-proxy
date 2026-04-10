package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type requestIDKey struct{}

const RequestIDHeader = "X-Request-ID"

// NewRequestIDMiddleware injects a unique request ID into every incoming
// request. If the client already supplies an X-Request-ID header that value
// is reused; otherwise a new UUID v4 is generated. The ID is stored in the
// request context and forwarded to the upstream backend via the same header.
func NewRequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		// Propagate to upstream.
		r.Header.Set(RequestIDHeader, id)

		// Expose to the client.
		w.Header().Set(RequestIDHeader, id)

		// Store in context so downstream handlers can read it.
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext returns the request ID stored in ctx, or an empty
// string if none is present.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey{}).(string)
	return v
}
