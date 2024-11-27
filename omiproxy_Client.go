package omiproxy

import (
	"github.com/go-redis/redis/v8"
)

type Client struct {
	opts *redis.Options
}

func (c *Client) NewProxy(serverName, address string) *OmiProxy {
	proxy := newOmiProxy(c.opts, serverName, address)
	return proxy
}
