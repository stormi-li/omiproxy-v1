package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omiproxy-v1"
)

var redisAddr = "118.25.196.166:3934"
var password = "12982397StrongPassw0rd"

func main() {
	c := omiproxy.NewClient(&redis.Options{Addr: redisAddr, Password: password})
	proxy := c.NewDomainNameProxy("http8000", "118.25.196.166:8000")
	proxy.Start(omiproxy.Http, "certs/stormili.crt", "certs/stormili.key")
}
