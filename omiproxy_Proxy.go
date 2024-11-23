package omiproxy

import (
	"net/http"
	"strings"

	"github.com/stormi-li/omicafe-v1"
)

// 主代理器
type Proxy struct {
	resolver       Resolver
	httpProxy      *HTTPProxy
	webSocketProxy *WebSocketProxy
}

func NewProxy(resolver Resolver, cache *omicafe.FileCache) *Proxy {
	return &Proxy{
		resolver:       resolver,
		httpProxy:      NewHTTPProxy(),
		webSocketProxy: &WebSocketProxy{},
	}
}
func (p *Proxy) SetCache(cache *omicafe.FileCache) {
	p.httpProxy.SetCache(cache)
}
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	host, err := p.resolver.Resolve(r)
	if err != nil {
		return err
	}
	r.URL.Host = host

	if isWebSocketRequest(r) {
		p.webSocketProxy.Forward(w, r)
	} else {
		p.httpProxy.Forward(w, r)
	}
	return nil
}

// 判断是否为 WebSocket 请求
func isWebSocketRequest(r *http.Request) bool {
	return r.Header.Get("Upgrade") == "websocket" &&
		strings.ToLower(r.Header.Get("Connection")) == "upgrade"
}
