package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// BreachLogConfig holds configuration for the breach logger middleware.
type BreachLogConfig struct {
	// Logger is the structured logger to write breach events to.
	Logger *slog.Logger
	// WindowSize is how many recent status codes to keep per key.
	WindowSize int
	// ErrorThreshold is the fraction of errors (0.0–1.0) that triggers a breach log.
	ErrorThreshold float64
	// KeyFunc extracts a grouping key from the request (e.g. route, client IP).
	KeyFunc func(r *http.Request) string
}

// DefaultBreachLogConfig returns sensible defaults.
func DefaultBreachLogConfig(logger *slog.Logger) BreachLogConfig {
	return BreachLogConfig{
		Logger:         logger,
		WindowSize:     20,
		ErrorThreshold: 0.5,
		KeyFunc:        func(r *http.Request) string { return r.URL.Path },
	}
}

type breachWindow struct {
	mu      sync.Mutex
	codes   []int
	pos     int
	full    bool
	logged  time.Time
}

func (w *breachWindow) record(code, size int) (ratio float64, breach bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.codes[w.pos] = code
	w.pos = (w.pos + 1) % size
	if w.pos == 0 {
		w.full = true
	}
	length := size
	if !w.full {
		length = w.pos
	}
	if length == 0 {
		return 0, false
	}
	errors := 0
	for i := 0; i < length; i++ {
		if w.codes[i] >= 500 {
			errors++
		}
	}
	ratio = float64(errors) / float64(length)
	breach = w.full && time.Since(w.logged) > 5*time.Second
	return ratio, breach
}

// NewBreachLogMiddleware returns middleware that logs when error rates breach a threshold.
func NewBreachLogMiddleware(cfg BreachLogConfig) func(http.Handler) http.Handler {
	windows := &sync.Map{}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)
			key := cfg.KeyFunc(r)
			val, _ := windows.LoadOrStore(key, &breachWindow{
				codes: make([]int, cfg.WindowSize),
			})
			win := val.(*breachWindow)
			ratio, breach := win.record(rec.Status(), cfg.WindowSize)
			if breach && ratio >= cfg.ErrorThreshold {
				win.mu.Lock()
				win.logged = time.Now()
				win.mu.Unlock()
				cfg.Logger.Warn("error rate breach detected",
					"key", key,
					"error_ratio", ratio,
					"threshold", cfg.ErrorThreshold,
					"window", cfg.WindowSize,
				)
			}
		})
	}
}
