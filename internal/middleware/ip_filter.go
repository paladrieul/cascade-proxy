package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"
)

// IPFilterConfig holds configuration for IP-based access control.
type IPFilterConfig struct {
	// AllowedCIDRs is a list of CIDR blocks that are permitted.
	// If non-empty, requests from IPs outside these ranges are denied.
	AllowedCIDRs []string
	// BlockedCIDRs is a list of CIDR blocks that are explicitly denied.
	BlockedCIDRs []string
	Logger       *log.Logger
}

// DefaultIPFilterConfig returns a permissive default configuration.
func DefaultIPFilterConfig() IPFilterConfig {
	return IPFilterConfig{
		AllowedCIDRs: []string{},
		BlockedCIDRs: []string{},
		Logger:       log.Default(),
	}
}

// NewIPFilterMiddleware returns an HTTP middleware that enforces IP-based
// access control. Blocked CIDRs are evaluated before allowed CIDRs.
func NewIPFilterMiddleware(cfg IPFilterConfig, next http.Handler) http.Handler {
	allowed := parseCIDRs(cfg.AllowedCIDRs)
	blocked := parseCIDRs(cfg.BlockedCIDRs)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)

		if matchesCIDR(ip, blocked) {
			cfg.Logger.Printf("ip_filter: blocked request from %s", ip)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if len(allowed) > 0 && !matchesCIDR(ip, allowed) {
			cfg.Logger.Printf("ip_filter: denied request from %s (not in allowlist)", ip)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseCIDRs(cidrs []string) []*net.IPNet {
	var nets []*net.IPNet
	for _, c := range cidrs {
		_, ipNet, err := net.ParseCIDR(c)
		if err == nil {
			nets = append(nets, ipNet)
		}
	}
	return nets
}

func matchesCIDR(ip net.IP, nets []*net.IPNet) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func extractIP(r *http.Request) net.IP {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.SplitN(fwd, ",", 2)
		if ip := net.ParseIP(strings.TrimSpace(parts[0])); ip != nil {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return net.ParseIP(r.RemoteAddr)
	}
	return net.ParseIP(host)
}
