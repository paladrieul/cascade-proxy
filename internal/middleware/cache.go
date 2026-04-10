package middleware

import (
	"net/http"
	"sync"
	"time"
)

// CacheConfig holds configuration for the response cache middleware.
type CacheConfig struct {
	// TTL is how long a cached response is considered fresh.
	TTL time.Duration
	// MaxEntries is the maximum number of entries to hold in the cache.
	MaxEntries int
}

// DefaultCacheConfig returns a CacheConfig with sensible defaults.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:        30 * time.Second,
		MaxEntries: 256,
	}
}

type cacheEntry struct {
	header http.Header
	body   []byte
	status int
	exp    time.Time
}

type responseCache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
	cfg     CacheConfig
}

func newResponseCache(cfg CacheConfig) *responseCache {
	return &responseCache{
		entries: make(map[string]cacheEntry, cfg.MaxEntries),
		cfg:     cfg,
	}
}

func (c *responseCache) get(key string) (cacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.exp) {
		delete(c.entries, key)
		return cacheEntry{}, false
	}
	return e, true
}

func (c *responseCache) set(key string, e cacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.cfg.MaxEntries {
		// evict one arbitrary entry to stay within limit
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}
	e.exp = time.Now().Add(c.cfg.TTL)
	c.entries[key] = e
}

// NewCacheMiddleware returns an HTTP middleware that caches GET responses.
// Only responses with status 200 are cached. All other methods bypass the cache.
func NewCacheMiddleware(cfg CacheConfig, next http.Handler) http.Handler {
	cache := newResponseCache(cfg)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		key := r.URL.RequestURI()
		if entry, ok := cache.get(key); ok {
			for k, vals := range entry.header {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(entry.status)
			w.Write(entry.body) //nolint:errcheck
			return
		}
		rec := NewResponseRecorder(w)
		next.ServeHTTP(rec, r)
		if rec.Status() == http.StatusOK {
			headerCopy := rec.Header().Clone()
			cache.set(key, cacheEntry{
				header: headerCopy,
				body:   rec.Body(),
				status: rec.Status(),
			})
		}
	})
}
