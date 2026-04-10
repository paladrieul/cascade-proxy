package middleware

import (
	"net/http"
	"regexp"
	"strings"
)

// RewriteRule defines a single path rewrite rule using a regexp pattern.
type RewriteRule struct {
	Pattern     *regexp.Regexp
	Replacement string
}

// RewriteConfig holds configuration for the rewrite middleware.
type RewriteConfig struct {
	// Rules is an ordered list of rewrite rules applied to the request path.
	Rules []RewriteRule
	// StripPrefix removes the given prefix from the path before forwarding.
	StripPrefix string
}

// DefaultRewriteConfig returns a no-op RewriteConfig.
func DefaultRewriteConfig() RewriteConfig {
	return RewriteConfig{}
}

// NewRewriteMiddleware returns an HTTP middleware that rewrites request paths
// according to the provided RewriteConfig before passing the request downstream.
//
// Rules are applied in order; the first matching rule wins. If StripPrefix is
// set it is applied after rule matching.
func NewRewriteMiddleware(cfg RewriteConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		for _, rule := range cfg.Rules {
			if rule.Pattern.MatchString(path) {
				path = rule.Pattern.ReplaceAllString(path, rule.Replacement)
				break
			}
		}

		if cfg.StripPrefix != "" {
			path = strings.TrimPrefix(path, cfg.StripPrefix)
			if path == "" {
				path = "/"
			}
		}

		// Clone the URL so we don't mutate the original.
		newURL := *r.URL
		newURL.Path = path
		newReq := r.Clone(r.Context())
		newReq.URL = &newURL
		newReq.RequestURI = newURL.RequestURI()

		next.ServeHTTP(w, newReq)
	})
}
