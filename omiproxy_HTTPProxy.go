package omiproxy

import (
	"bytes"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/stormi-li/omicafe-v1"
	"github.com/stormi-li/omiresolver-v1"
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
	Transport http.RoundTripper
	Resolver  *omiresolver.Resolver
}

func NewHTTPProxy(resolver *omiresolver.Resolver, cache *omicafe.FileCache, insecureSkipVerify bool) *HTTPProxy {
	return &HTTPProxy{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerify, // 禁用证书验证
			},
		},
		Resolver: resolver,
		cache:    cache,
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
	r.URL.Host = r.Host
	targetURL, err := p.Resolver.Resolve(*r.URL)
	if err != nil {
		log.Println(err)
		return
	}

	if targetURL.Scheme == "" {
		targetURL.Scheme = "http"
	}

	proxyURL := &url.URL{
		Scheme: targetURL.Scheme,
		Host:   targetURL.Host,
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	r.URL.Path = targetURL.Path
	proxy.Transport = &CaptureResponseRoundTripper{Transport: p.Transport, cache: p.cache, url: r.URL}
	proxy.ServeHTTP(w, r)
}
