package omiproxy

import (
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omicafe-v1"
	"github.com/stormi-li/omiresolver-v1"
	"github.com/stormi-li/omiserd-v1"
)

// 主代理器
type Proxy struct {
	httpProxy      *HTTPProxy
	webSocketProxy *WebSocketProxy
}

func NewProxy(opts *redis.Options, nodeType omiserd.NodeType, cache *omicafe.FileCache, insecureSkipVerify bool) *Proxy {
	resolver := omiresolver.NewResolver(opts, nodeType)
	return &Proxy{
		httpProxy:      NewHTTPProxy(resolver, cache, insecureSkipVerify),
		webSocketProxy: NewWebSocketProxy(resolver, insecureSkipVerify),
	}
}
func (p *Proxy) SetCache(cache *omicafe.FileCache) {
	p.httpProxy.SetCache(cache)
}
func (p *Proxy) Forward(w http.ResponseWriter, r *http.Request) error {
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
