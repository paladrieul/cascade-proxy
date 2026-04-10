package middleware

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/cascade-proxy/internal/balancer"
)

// NewBalancerMiddleware returns an http.Handler that forwards each request to
// the next backend selected by the provided Balancer using a reverse proxy.
// On balancer errors a 502 Bad Gateway is returned.
func NewBalancerMiddleware(b *balancer.Balancer, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, err := b.Next()
		if err != nil {
			logger.Printf("balancer: no target available: %v", err)
			http.Error(w, "no backend available", http.StatusBadGateway)
			return
		}

		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.Host = target.Host
				if _, ok := req.Header["User-Agent"]; !ok {
					req.Header.Set("User-Agent", "cascade-proxy")
				}
			},
			ErrorHandler: func(rw http.ResponseWriter, req *http.Request, e error) {
				logger.Printf("balancer: upstream error for %s: %v", target.Host, e)
				http.Error(rw, "bad gateway", http.StatusBadGateway)
			},
		}

		logger.Printf("balancer: routing request to %s", target.Host)
		proxy.ServeHTTP(w, r)
	})
}
