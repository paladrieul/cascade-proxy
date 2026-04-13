package middleware

import (
	"net/http"
	"strings"
)

// RedirectRule defines a single redirect mapping.
type RedirectRule struct {
	From string
	To   string
	Code int // e.g. 301, 302
}

// RedirectConfig holds configuration for the redirect middleware.
type RedirectConfig struct {
	Rules []RedirectRule
}

// DefaultRedirectConfig returns a RedirectConfig with no rules.
func DefaultRedirectConfig() RedirectConfig {
	return RedirectConfig{
		Rules: []RedirectRule{},
	}
}

// NewRedirectMiddleware returns an HTTP middleware that redirects requests
// matching any configured rule. Rules are evaluated in order; the first match
// wins. If the rule Code is 0 it defaults to 302.
func NewRedirectMiddleware(cfg RedirectConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		for _, rule := range cfg.Rules {
			if matchRedirect(path, rule.From) {
				code := rule.Code
				if code == 0 {
					code = http.StatusFound
				}
				target := buildRedirectTarget(r, rule)
				http.Redirect(w, r, target, code)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// matchRedirect returns true when path equals from, or when from ends with
// "/*" and path starts with the prefix.
func matchRedirect(path, from string) bool {
	if strings.HasSuffix(from, "/*") {
		prefix := strings.TrimSuffix(from, "/*")
		return strings.HasPrefix(path, prefix)
	}
	return path == from
}

// buildRedirectTarget constructs the redirect URL, preserving the query string.
func buildRedirectTarget(r *http.Request, rule RedirectRule) string {
	target := rule.To
	if strings.HasSuffix(rule.From, "/*") {
		prefix := strings.TrimSuffix(rule.From, "/*")
		rest := strings.TrimPrefix(r.URL.Path, prefix)
		if rest != "" && rest != "/" {
			target = strings.TrimSuffix(rule.To, "/") + rest
		}
	}
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	return target
}
