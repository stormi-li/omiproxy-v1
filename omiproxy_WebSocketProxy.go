package omiproxy

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocket代理
type WebSocketProxy struct{}

func (wp *WebSocketProxy) Forward(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer clientConn.Close()

	r.URL.Scheme = "ws"
	targetConn, _, err := websocket.DefaultDialer.Dial(r.URL.String(), nil)
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
