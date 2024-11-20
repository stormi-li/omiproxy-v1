package main

import (
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omiserd-v1"
)

var redisAddr = "118.25.196.166:3934"
var password = "12982397StrongPassw0rd"

func main() {
	register := omiserd.NewClient(&redis.Options{Addr: redisAddr, Password: password}, omiserd.Server).NewRegister("hello_server", "118.25.196.166:8081")
	register.RegisterAndListen(1, func(port string) {
		http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello", r.URL.Query().Get("name"), ", welcome to use omi, send by http server")
		})
		http.ListenAndServe(port, nil)
	})
}
