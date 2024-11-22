package omiproxy

import (
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omicafe-v1"
	"github.com/stormi-li/omiserd-v1"
)

type Client struct {
	opts *redis.Options
}

func (c *Client) NewDomainNameProxy(serverName, address string) *DomainNameProxy {
	return &DomainNameProxy{
		router:       newRouter(c.opts, omiserd.Web),
		webRegister:  omiserd.NewClient(c.opts, omiserd.Web).NewRegister(serverName, address),
		serverName:   serverName,
		address:      address,
		failCallback: func(w http.ResponseWriter, r *http.Request) {},
	}
}

func (c *Client) NewReverseProxy(serverName, address string) *ReverseProxy {
	return &ReverseProxy{
		router:       newRouter(c.opts, omiserd.Server),
		webRegister:  omiserd.NewClient(c.opts, omiserd.Web).NewRegister(serverName, address),
		serverName:   serverName,
		address:      address,
		failCallback: func(w http.ResponseWriter, r *http.Request) {},
	}
}

func registerHandler(register *omiserd.Register, cache *omicafe.FileCache) {
	register.SetRegisterHandler(func(register *omiserd.Register) {
		if cache == nil {
			register.Data["cache_state"] = "closed"
			register.Data["cache_max_size"] = "0"
			register.Data["cache_current_size"] = "0"
			register.Data["cache_count"] = "0"
			register.Data["cache_hit_num"] = "0"
			register.Data["cache_fail_num"] = "0"
			register.Data["cache_clear_num"] = "0"
		} else {
			register.Data["cache_state"] = "open"
			register.Data["cache_max_size"] = strconv.Itoa(cache.MaxSize())
			register.Data["cache_current_size"] = strconv.Itoa(cache.CurrentSize())
			register.Data["cache_count"] = strconv.Itoa(int(cache.Cache_count))
			register.Data["cache_hit_num"] = strconv.Itoa(int(cache.Cache_hit_num))
			register.Data["cache_fail_num"] = strconv.Itoa(int(cache.Cache_fail_num))
			register.Data["cache_clear_num"] = strconv.Itoa(int(cache.Cache_clear_num))
		}
	})
}
