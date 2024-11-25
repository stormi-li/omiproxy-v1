package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omiproxy-v1"
)

var redisAddr = "118.25.196.166:3934"
var password = "12982397StrongPassw0rd"

func main() {
	c := omiproxy.NewClient(&redis.Options{Addr: redisAddr, Password: password})
	proxy := c.NewProxy("http7000", "118.25.196.166:7000", omiproxy.DomainMode)
	proxy.Start(omiproxy.Http, 1, "", "")
}
