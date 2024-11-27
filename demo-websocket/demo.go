package main

import (
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/stormi-li/omiserd-v1"
)

var redisAddr = "118.25.196.166:3934"
var password = "12982397StrongPassw0rd"

func main() {
	Client()
}

func Client() {
	c, _, err := websocket.DefaultDialer.Dial("ws://118.25.196.166:8000/hello_websocket_server/hello", nil)
	fmt.Println(err)
	c.WriteMessage(1, []byte("lili"))
	_, d, _ := c.ReadMessage()
	fmt.Println(string(d))
}

func Server() {
	register := omiserd.NewClient(&redis.Options{Addr: redisAddr, Password: password}, omiserd.Web).NewRegister("hello_websocket_server", "118.25.196.166:8082")
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		clientConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		msgtype, msg, _ := clientConn.ReadMessage()
		clientConn.WriteMessage(msgtype, []byte("hello "+string(msg)))
		fmt.Println(string(msg))
	})
	register.RegisterAndServe(1, func(port string) {
		err := http.ListenAndServe(port, nil)
		fmt.Println(err)
	})
}
