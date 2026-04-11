package middleware

import (
	"log/slog"
	"net/http"
)

// DefaultRequestSizeConfig returns a RequestSizeConfig with sensible defaults.
func DefaultRequestSizeConfig() RequestSizeConfig {
	return RequestSizeConfig{
		MaxHeaderBytes: 8 * 1024,       // 8 KB
		MaxURLLength:   2048,           // 2048 chars
		MaxQueryParams: 50,
		Logger:         slog.Default(),
	}
}

// RequestSizeConfig configures limits on incoming request metadata.
type RequestSizeConfig struct {
	// MaxHeaderBytes is the maximum total size of request headers in bytes.
	MaxHeaderBytes int
	// MaxURLLength is the maximum length of the raw request URL.
	MaxURLLength int
	// MaxQueryParams is the maximum number of query parameters allowed.
	MaxQueryParams int
	Logger         *slog.Logger
}

// NewRequestSizeMiddleware rejects requests whose headers, URL, or query
// parameter count exceed the configured limits.
func NewRequestSizeMiddleware(cfg RequestSizeConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.MaxURLLength > 0 && len(r.RequestURI) > cfg.MaxURLLength {
			cfg.Logger.Warn("request URI too long",
				"uri_length", len(r.RequestURI),
				"max", cfg.MaxURLLength,
			)
			http.Error(w, "414 URI Too Long", http.StatusRequestURITooLong)
			return
		}

		if cfg.MaxQueryParams > 0 {
			if err := r.ParseForm(); err == nil {
				if len(r.Form) > cfg.MaxQueryParams {
					cfg.Logger.Warn("too many query parameters",
						"count", len(r.Form),
						"max", cfg.MaxQueryParams,
					)
					http.Error(w, "400 Bad Request", http.StatusBadRequest)
					return
				}
			}
		}

		if cfg.MaxHeaderBytes > 0 {
			total := 0
			for name, vals := range r.Header {
				total += len(name)
				for _, v := range vals {
					total += len(v)
				}
			}
			if total > cfg.MaxHeaderBytes {
				cfg.Logger.Warn("request headers too large",
					"header_bytes", total,
					"max", cfg.MaxHeaderBytes,
				)
				http.Error(w, "431 Request Header Fields Too Large", http.StatusRequestHeaderFieldsTooLarge)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
