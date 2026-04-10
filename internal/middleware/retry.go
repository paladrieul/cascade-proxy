package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

// RetryConfig holds configuration for the retry middleware.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts (1 = no retry).
	MaxAttempts int
	// Delay is the wait time between attempts.
	Delay time.Duration
	// RetryableStatus defines HTTP status codes that trigger a retry.
	RetryableStatus []int
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		Delay:           100 * time.Millisecond,
		RetryableStatus: []int{502, 503, 504},
	}
}

// NewRetryMiddleware wraps handler with automatic retry logic based on cfg.
func NewRetryMiddleware(cfg RetryConfig, logger *log.Logger, next http.Handler) http.Handler {
	retryable := make(map[int]bool, len(cfg.RetryableStatus))
	for _, s := range cfg.RetryableStatus {
		retryable[s] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rec *ResponseRecorder

		for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
			rec = NewResponseRecorder(w)
			next.ServeHTTP(rec, r)

			if !retryable[rec.Status] {
				rec.Flush(w)
				return
			}

			if attempt < cfg.MaxAttempts {
				logger.Printf("retry: attempt %d/%d failed with status %d, retrying in %s",
					attempt, cfg.MaxAttempts, rec.Status, cfg.Delay)
				time.Sleep(cfg.Delay)
			}
		}

		// All attempts exhausted — write the last recorded response.
		rec.Flush(w)
	})
}

// NewResponseRecorder creates a buffered ResponseRecorder that defers writes.
func newBufferedRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}
