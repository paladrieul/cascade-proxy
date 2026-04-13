package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
)

// DefaultGeoBlockConfig returns a permissive default configuration.
func DefaultGeoBlockConfig() GeoBlockConfig {
	return GeoBlockConfig{
		BlockedCountries: []string{},
		AllowedCountries: []string{},
		CountryHeader:    "X-Country-Code",
		DenyStatus:       http.StatusForbidden,
	}
}

// GeoBlockConfig configures the geo-blocking middleware.
// Either BlockedCountries (denylist) or AllowedCountries (allowlist) may be
// set; if both are set, AllowedCountries takes precedence.
type GeoBlockConfig struct {
	// BlockedCountries is a list of ISO 3166-1 alpha-2 country codes to deny.
	BlockedCountries []string
	// AllowedCountries is an allowlist; any country not listed is denied.
	// When non-empty, BlockedCountries is ignored.
	AllowedCountries []string
	// CountryHeader is the request header that carries the country code,
	// typically injected by an upstream CDN or load balancer.
	CountryHeader string
	// DenyStatus is the HTTP status code returned for blocked requests.
	DenyStatus int
	Logger     *slog.Logger
}

// NewGeoBlockMiddleware returns an http.Handler that enforces country-based
// access control using the country code found in a request header.
func NewGeoBlockMiddleware(cfg GeoBlockConfig, next http.Handler) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	allowed := toSet(cfg.AllowedCountries)
	blocked := toSet(cfg.BlockedCountries)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		country := strings.ToUpper(strings.TrimSpace(r.Header.Get(cfg.CountryHeader)))

		if country == "" {
			// No country header — derive a best-effort value from RemoteAddr.
			// In production this would call a GeoIP lookup; here we pass through.
			next.ServeHTTP(w, r)
			return
		}

		if len(allowed) > 0 {
			if !allowed[country] {
				cfg.Logger.Warn("geo-block: country not in allowlist",
					"country", country,
					"remote", remoteIP(r),
				)
				http.Error(w, http.StatusText(cfg.DenyStatus), cfg.DenyStatus)
				return
			}
		} else if blocked[country] {
			cfg.Logger.Warn("geo-block: country in blocklist",
				"country", country,
				"remote", remoteIP(r),
			)
			http.Error(w, http.StatusText(cfg.DenyStatus), cfg.DenyStatus)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func toSet(codes []string) map[string]bool {
	s := make(map[string]bool, len(codes))
	for _, c := range codes {
		s[strings.ToUpper(c)] = true
	}
	return s
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
