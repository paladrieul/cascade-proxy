package middleware

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
)

// TransformRule defines a single find-and-replace rule applied to response bodies.
type TransformRule struct {
	Pattern     *regexp.Regexp
	Replacement string
}

// ResponseTransformConfig holds configuration for the response transform middleware.
type ResponseTransformConfig struct {
	// Rules is the ordered list of transformations to apply.
	Rules []TransformRule
	// ContentTypes limits transformation to responses whose Content-Type contains
	// one of the listed substrings. Empty means all content types.
	ContentTypes []string
}

// DefaultResponseTransformConfig returns a config with no rules.
func DefaultResponseTransformConfig() ResponseTransformConfig {
	return ResponseTransformConfig{}
}

// NewResponseTransformMiddleware returns middleware that rewrites response bodies
// according to the configured transformation rules.
func NewResponseTransformMiddleware(cfg ResponseTransformConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(cfg.Rules) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		buf := &transformRecorder{header: w.Header(), code: http.StatusOK}
		next.ServeHTTP(buf, r)

		if !matchesContentType(buf.header.Get("Content-Type"), cfg.ContentTypes) {
			w.WriteHeader(buf.code)
			_, _ = w.Write(buf.body.Bytes())
			return
		}

		body := buf.body.Bytes()
		for _, rule := range cfg.Rules {
			body = rule.Pattern.ReplaceAll(body, []byte(rule.Replacement))
		}

		w.Header().Set("Content-Length", "")
		w.WriteHeader(buf.code)
		_, _ = w.Write(body)
	})
}

func matchesContentType(ct string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if len(ct) >= len(a) && containsSubstring(ct, a) {
			return true
		}
	}
	return false
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

type transformRecorder struct {
	header http.Header
	code   int
	body   bytes.Buffer
}

func (t *transformRecorder) Header() http.Header        { return t.header }
func (t *transformRecorder) WriteHeader(code int)       { t.code = code }
func (t *transformRecorder) Write(b []byte) (int, error) { return t.body.Write(b) }
func (t *transformRecorder) Result() *http.Response {
	return &http.Response{
		StatusCode: t.code,
		Header:     t.header,
		Body:       io.NopCloser(&t.body),
	}
}
