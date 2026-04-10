package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// DefaultDedupeConfig returns a sensible default configuration.
func DefaultDedupeConfig() DedupeConfig {
	return DedupeConfig{
		TTL:    500 * time.Millisecond,
		Methods: []string{http.MethodPost, http.MethodPut, http.MethodPatch},
	}
}

// DedupeConfig controls deduplication behaviour.
type DedupeConfig struct {
	// TTL is how long a request fingerprint is remembered.
	TTL time.Duration
	// Methods lists the HTTP methods that are subject to deduplication.
	Methods []string
}

type dedupeEntry struct {
	expiry time.Time
	status int
	body   []byte
	headers http.Header
}

type dedupeMiddleware struct {
	cfg     DedupeConfig
	mu      sync.Mutex
	seen    map[string]*dedupeEntry
	methods map[string]struct{}
}

// NewDedupeMiddleware returns middleware that short-circuits duplicate
// in-flight requests within the configured TTL window.
func NewDedupeMiddleware(cfg DedupeConfig, next http.Handler) http.Handler {
	methods := make(map[string]struct{}, len(cfg.Methods))
	for _, m := range cfg.Methods {
		methods[m] = struct{}{}
	}
	d := &dedupeMiddleware{
		cfg:     cfg,
		seen:    make(map[string]*dedupeEntry),
		methods: methods,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := d.methods[r.Method]; !ok {
			next.ServeHTTP(w, r)
			return
		}
		key := d.fingerprint(r)
		d.mu.Lock()
		if entry, exists := d.seen[key]; exists && time.Now().Before(entry.expiry) {
			d.mu.Unlock()
			for k, vals := range entry.headers {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.Header().Set("X-Dedupe-Hit", "true")
			w.WriteHeader(entry.status)
			w.Write(entry.body) //nolint:errcheck
			return
		}
		d.mu.Unlock()

		rec := newBufferedRecorder(w)
		next.ServeHTTP(rec, r)

		d.mu.Lock()
		d.seen[key] = &dedupeEntry{
			expiry:  time.Now().Add(d.cfg.TTL),
			status:  rec.status,
			body:    rec.body.Bytes(),
			headers: rec.Header().Clone(),
		}
		d.mu.Unlock()
	})
}

func (d *dedupeMiddleware) fingerprint(r *http.Request) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s:%s", r.Method, r.URL.String())
	if id := r.Header.Get("Idempotency-Key"); id != "" {
		fmt.Fprintf(h, ":%s", id)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
