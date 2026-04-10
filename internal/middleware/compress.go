package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// DefaultCompressConfig holds sensible defaults for compression middleware.
var DefaultCompressConfig = CompressConfig{
	Level:     gzip.DefaultCompression,
	MinLength: 1024,
}

// CompressConfig controls gzip compression behaviour.
type CompressConfig struct {
	// Level is the gzip compression level (1–9, or -1 for default).
	Level int
	// MinLength is the minimum response body size in bytes before compression
	// is applied. Responses smaller than this are passed through unchanged.
	MinLength int
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer    *gzip.Writer
	status    int
	minLength int
	buf       []byte
	committed bool
}

func (g *gzipResponseWriter) WriteHeader(status int) {
	g.status = status
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if !g.committed {
		g.buf = append(g.buf, b...)
		if len(g.buf) >= g.minLength {
			g.commit(true)
			return g.writer.Write(g.buf)
		}
		return len(b), nil
	}
	if g.writer != nil {
		return g.writer.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipResponseWriter) commit(compress bool) {
	g.committed = true
	if compress {
		g.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		g.ResponseWriter.Header().Del("Content-Length")
	}
	if g.status != 0 {
		g.ResponseWriter.WriteHeader(g.status)
	}
}

func (g *gzipResponseWriter) flush() {
	if !g.committed {
		// buffer never reached minLength — write plain
		g.commit(false)
		g.ResponseWriter.Write(g.buf) //nolint:errcheck
	} else if g.writer != nil {
		g.writer.Close() //nolint:errcheck
	}
}

var gzipPool = sync.Pool{
	New: func() any { w, _ := gzip.NewWriter(io.Discard, gzip.DefaultCompression); return w },
}

// NewCompressMiddleware returns middleware that gzip-compresses responses when
// the client signals support via Accept-Encoding.
func NewCompressMiddleware(cfg CompressConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer func() {
			gzipPool.Put(gz)
		}()

		grw := &gzipResponseWriter{
			ResponseWriter: w,
			writer:         gz,
			minLength:      cfg.MinLength,
		}
		defer grw.flush()

		next.ServeHTTP(grw, r)
	})
}
