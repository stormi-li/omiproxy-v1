package omiproxy

import (
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omicafe-v1"
	register "github.com/stormi-li/omiserd-v1/omiserd_register"
)

type Client struct {
	opts *redis.Options
}

func (c *Client) NewProxy(serverName, address string, mode ProxyMode) *OmiProxy {
	proxy := newOmiProxy(c.opts, serverName, address, mode)
	proxy.Register.AddRegisterHandleFunc(func(register *register.Register) {
		cacheInfoHandler(register, proxy.cache)
	})
	return proxy
}

// func (c *Client) NewDomainNameProxy(serverName, address string) *DomainNameProxy {
// 	return &DomainNameProxy{
// 		router:       newRouter(c.opts, omiserd.Web),
// 		webRegister:  omiserd.NewClient(c.opts, omiserd.Web).NewRegister(serverName, address),
// 		serverName:   serverName,
// 		address:      address,
// 		failCallback: func(w http.ResponseWriter, r *http.Request) {},
// 	}
// }

// func (c *Client) NewReverseProxy(serverName, address string) *ReverseProxy {
// 	return &ReverseProxy{
// 		router:       newRouter(c.opts, omiserd.Server),
// 		webRegister:  omiserd.NewClient(c.opts, omiserd.Web).NewRegister(serverName, address),
// 		serverName:   serverName,
// 		address:      address,
// 		failCallback: func(w http.ResponseWriter, r *http.Request) {},
// 	}
// }

func cacheInfoHandler(register *register.Register, cache *omicafe.FileCache) {
	if cache == nil {
		register.Data["cache_state"] = "closed"
	} else {
		register.Data["cache_state"] = "open"
		register.Data["cache_max_size"] = strconv.Itoa(cache.MaxSize)
		register.Data["cache_current_size"] = strconv.Itoa(cache.CurrentSize())
		register.Data["cache_num"] = strconv.Itoa(int(cache.GetCacheNum()))
		register.Data["cache_hit_count"] = strconv.Itoa(int(cache.GetCacheHitCount()))
		register.Data["cache_miss_count"] = strconv.Itoa(int(cache.GetCacheMissCount()))
		register.Data["cache_clear_count"] = strconv.Itoa(int(cache.GetCacheClearCount()))
	}
}
