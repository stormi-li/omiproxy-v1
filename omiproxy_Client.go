package omiproxy

import (
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omiserd-v1"
)

type Client struct {
	opts *redis.Options
}

func (c *Client) NewDNSProxy(serverName, address string) *DNSProxy {
	return &DNSProxy{
		router:       newRouter(c.opts, omiserd.Web),
		webRegister:  omiserd.NewClient(c.opts, omiserd.Web).NewRegister(serverName, address),
		serverName:   serverName,
		address:      address,
		failCallback: func(w http.ResponseWriter, r *http.Request) { return },
	}
}

func (c *Client) NewReverseProxy(serverName, address string) *ReverseProxy {
	return &ReverseProxy{
		router:       newRouter(c.opts, omiserd.Server),
		webRegister:  omiserd.NewClient(c.opts, omiserd.Web).NewRegister(serverName, address),
		serverName:   serverName,
		address:      address,
		failCallback: func(w http.ResponseWriter, r *http.Request) { return },
	}
}
