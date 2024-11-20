package main

import (
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/stormi-li/omiserd-v1"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// 升级 HTTP 请求到 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()
	message := fmt.Sprint("Hello ", r.URL.Query().Get("name"), ", welcome to use omi, send by websocket server")
	// 向客户端发送 "Hello World" 消息
	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
}

var redisAddr = "118.25.196.166:3934"
var password = "12982397StrongPassw0rd"

func main() {
	register := omiserd.NewClient(&redis.Options{Addr: redisAddr, Password: password}, omiserd.Server).NewRegister("hello_websocket_server", "118.25.196.166:8082")
	register.RegisterAndListen(1, func(port string) {
		http.HandleFunc("/hello", wsHandler) // 将 /request 路径映射到 wsHandler
		if err := http.ListenAndServe(port, nil); err != nil {
			fmt.Println("Error starting server:", err)
		}
	})
}
