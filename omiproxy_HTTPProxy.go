package omiproxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/stormi-li/omicafe-v1"
)

// 捕获响应内容的 RoundTripper
type CaptureResponseRoundTripper struct {
	Transport http.RoundTripper
	cache     *omicafe.FileCache
	url       *url.URL
}

func (c *CaptureResponseRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := c.Transport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	if c.cache != nil && r.Method == "GET" {
		go c.cache.Set(c.url.String(), bodyBytes)
	}
	return resp, nil
}

// HTTP代理
type HTTPProxy struct {
	cache     *omicafe.FileCache
	transport http.RoundTripper
}

func NewHTTPProxy() *HTTPProxy {
	return &HTTPProxy{
		transport: http.DefaultTransport,
	}
}
func (p *HTTPProxy) SetCache(cache *omicafe.FileCache) {
	p.cache = cache
}
func (p *HTTPProxy) Forward(w http.ResponseWriter, r *http.Request) {
	if p.cache != nil && r.Method == "GET" {
		if data, _ := p.cache.Get(r.URL.String()); len(data) != 0 {
			w.Write(data)
			return
		}
	}
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   r.URL.Host,
	}
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	proxy.Transport = &CaptureResponseRoundTripper{Transport: p.transport, cache: p.cache, url: r.URL}
	proxy.ServeHTTP(w, r)
}
