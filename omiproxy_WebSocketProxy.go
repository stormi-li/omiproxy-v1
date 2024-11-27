package omiproxy

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/stormi-li/omiresolver-v1"
)

// WebSocket代理
type WebSocketProxy struct {
	Resolver *omiresolver.Resolver
	Dialer   websocket.Dialer
}

func NewWebSocketProxy(resolver *omiresolver.Resolver, insecureSkipVerify bool) *WebSocketProxy {
	return &WebSocketProxy{
		Resolver: resolver,
		Dialer: websocket.Dialer{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerify, // 跳过证书验证
			},
		},
	}
}

var upgrader = websocket.Upgrader{}

func (wp *WebSocketProxy) Forward(w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer clientConn.Close()

	proxyURL, err := wp.Resolver.Resolve(*r.URL)
	if err != nil {
		log.Println(err)
		return
	}
	if proxyURL.Scheme == "https" {
		proxyURL.Scheme = "wss"
	} else {
		proxyURL.Scheme = "ws"
	}
	targetConn, _, err := wp.Dialer.Dial(proxyURL.String(), r.Header)
	if err != nil {
		log.Printf("无法连接到WebSocket服务器: %v", err)
		return
	}
	defer targetConn.Close()

	errChan := make(chan error, 2)
	go wp.copyData(clientConn, targetConn, errChan)
	go wp.copyData(targetConn, clientConn, errChan)
	<-errChan
}

func (wp *WebSocketProxy) copyData(src, dst *websocket.Conn, errChan chan error) {
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
