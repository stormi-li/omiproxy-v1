package omiproxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/stormi-li/omicafe-v1"
)

func pathRequestResolution(r *http.Request, router *router) error {
	serverName := strings.Split(r.URL.Path, "/")[1]
	if !router.Has(serverName) {
		return fmt.Errorf("未找到微服务: %s", serverName)
	}
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+serverName)
	host := router.getAddress(serverName)
	r.URL.Host = host
	return nil
}

func domainNameResolution(r *http.Request, router *router) error {
	domainName := strings.Split(r.Host, ":")[0]
	if !router.Has(domainName){
		return fmt.Errorf("未找到 Web 服务: %s", domainName)
	}
	host := router.getAddress(domainName)
	r.URL.Host = host
	return nil
}

func isWebSocketRequest(r *http.Request) bool {
	// 判断请求头中是否包含 WebSocket 升级请求特有的字段
	return r.Header.Get("Upgrade") == "websocket" &&
		r.Header.Get("Connection") == "Upgrade" &&
		r.Header.Get("Sec-WebSocket-Key") != ""
}

// 自定义 RoundTripper 用于捕获响应数据
type captureResponseRoundTripper struct {
	Transport http.RoundTripper
	cache     *omicafe.FileCache
	url       *url.URL
}

func (c *captureResponseRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	// 执行请求
	resp, err := c.Transport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	// 捕获响应内容
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 恢复原始响应体以便下游处理
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if c.cache != nil && r.Method == "GET" {
		go c.cache.Set(c.url.String(), bodyBytes)
	}

	// 返回原始响应
	return resp, nil
}

// 代理请求并捕获响应数据
func httpForward(w http.ResponseWriter, r *http.Request, cache *omicafe.FileCache) {
	if isWebSocketRequest(r) {
		return
	}
	if cache != nil && r.Method == "GET" {
		data := cache.Get(r.URL.String())
		if len(data) != 0 {
			w.Write(data)
			return
		}
	}
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   r.URL.Host,
	}

	// 使用自定义的 RoundTripper 来捕获响应数据
	proxy := httputil.NewSingleHostReverseProxy(proxyURL)
	proxy.Transport = &captureResponseRoundTripper{Transport: http.DefaultTransport, cache: cache, url: r.URL}

	// 转发请求并处理响应
	proxy.ServeHTTP(w, r)
}

var upgrader = websocket.Upgrader{}

func websocketForward(w http.ResponseWriter, r *http.Request) {
	if !isWebSocketRequest(r) {
		return
	}
	// 将客户端升级为WebSocket
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer clientConn.Close()

	r.URL.Scheme = "ws"

	// 连接到后端WebSocket服务器
	targetConn, _, err := websocket.DefaultDialer.Dial(r.URL.String(), nil)
	if err != nil {
		log.Printf("无法连接到WebSocket服务器: %v", err)
		return
	}
	defer targetConn.Close()

	// 开始数据转发
	errChan := make(chan error, 2)

	go copyWebSocketData(targetConn, clientConn, errChan)
	go copyWebSocketData(clientConn, targetConn, errChan)

	// 等待传输结束
	<-errChan
}

// WebSocket数据复制
func copyWebSocketData(dst, src *websocket.Conn, errChan chan error) {
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			errChan <- err
			return
		}
		if err := dst.WriteMessage(msgType, msg); err != nil {
			errChan <- err
			return
		}
	}
}
