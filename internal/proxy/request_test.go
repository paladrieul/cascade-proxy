package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestBuildRequestSetsForwardedHeaders(t *testing.T) {
	target, _ := url.Parse("http://backend:8080")
	r := httptest.NewRequest(http.MethodGet, "/api/v1", nil)
	r.RemoteAddr = "1.2.3.4:5678"
	r.Host = "frontend.example.com"

	out, err := buildRequest(r, target)
	if err != nil {
		t.Fatalf("buildRequest: %v", err)
	}

	if got := out.Header.Get("X-Forwarded-For"); got != r.RemoteAddr {
		t.Errorf("X-Forwarded-For = %q, want %q", got, r.RemoteAddr)
	}
	if got := out.Header.Get("X-Forwarded-Host"); got != r.Host {
		t.Errorf("X-Forwarded-Host = %q, want %q", got, r.Host)
	}
}

func TestBuildRequestComposesURL(t *testing.T) {
	target, _ := url.Parse("http://backend:8080/prefix")
	r := httptest.NewRequest(http.MethodGet, "/resource?foo=bar", nil)

	out, err := buildRequest(r, target)
	if err != nil {
		t.Fatalf("buildRequest: %v", err)
	}

	if !strings.Contains(out.URL.Path, "resource") {
		t.Errorf("expected path to contain 'resource', got %q", out.URL.Path)
	}
	if out.URL.RawQuery != "foo=bar" {
		t.Errorf("expected query 'foo=bar', got %q", out.URL.RawQuery)
	}
}

func TestSingleJoiningSlash(t *testing.T) {
	cases := []struct{ a, b, want string }{
		{"/prefix/", "/path", "/prefix/path"},
		{"/prefix", "path", "/prefix/path"},
		{"/prefix/", "path", "/prefix/path"},
		{"/prefix", "", "/prefix"},
	}
	for _, c := range cases {
		if got := singleJoiningSlash(c.a, c.b); got != c.want {
			t.Errorf("singleJoiningSlash(%q, %q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

func TestCopyHeadersSkipsHopByHop(t *testing.T) {
	src := http.Header{}
	src.Set("Content-Type", "application/json")
	src.Set("Connection", "keep-alive")
	src.Set("Transfer-Encoding", "chunked")

	dst := http.Header{}
	copyHeaders(dst, src)

	if dst.Get("Content-Type") == "" {
		t.Error("expected Content-Type to be copied")
	}
	if dst.Get("Connection") != "" {
		t.Error("expected Connection to be stripped")
	}
	if dst.Get("Transfer-Encoding") != "" {
		t.Error("expected Transfer-Encoding to be stripped")
	}
}
