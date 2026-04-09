package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// buildRequest constructs an outbound request directed at target.
func buildRequest(r *http.Request, target *url.URL) (*http.Request, error) {
	outURL := *target
	outURL.Path = singleJoiningSlash(target.Path, r.URL.Path)
	outURL.RawQuery = r.URL.RawQuery

	outReq, err := http.NewRequestWithContext(r.Context(), r.Method, outURL.String(), r.Body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	copyHeaders(outReq.Header, r.Header)
	outReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	outReq.Header.Set("X-Forwarded-Host", r.Host)

	return outReq, nil
}

// copyHeaders copies headers from src to dst, skipping hop-by-hop headers.
func copyHeaders(dst, src http.Header) {
	hopByHop := map[string]bool{
		"Connection": true, "Keep-Alive": true, "Proxy-Authenticate": true,
		"Proxy-Authorization": true, "Te": true, "Trailers": true,
		"Transfer-Encoding": true, "Upgrade": true,
	}
	for k, vv := range src {
		if hopByHop[k] {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// copyResponse writes the upstream response to the ResponseWriter.
func copyResponse(w http.ResponseWriter, resp *http.Response) {
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// singleJoiningSlash joins two path segments with exactly one slash.
func singleJoiningSlash(a, b string) string {
	aSlash := len(a) > 0 && a[len(a)-1] == '/'
	bSlash := len(b) > 0 && b[0] == '/'
	switch {
	case aSlash && bSlash:
		return a + b[1:]
	case !aSlash && !bSlash:
		if b == "" {
			return a
		}
		return a + "/" + b
	}
	return a + b
}
