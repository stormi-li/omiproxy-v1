package omiproxy

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omicafe-v1"
	"github.com/stormi-li/omiserd-v1"
)

const Command_open_cache = "open_cache"
const Command_update_cache_size = "update_cache_size"
const Default_cache_size = 100 * 1024 * 1024

type ProxyMode int

const DomainMode ProxyMode = 1
const PathMode ProxyMode = 2

func (Mode ProxyMode) String() string {
	switch Mode {
	case DomainMode:
		return "DomainMode"
	case PathMode:
		return "PathMode"
	default:
		return "unknown"
	}
}

type ProxyProtocal int

const Http ProxyProtocal = 1
const Https ProxyProtocal = 2

// ProxyProtocal 枚举类型的字符串表示
func (proto ProxyProtocal) String() string {
	switch proto {
	case Http:
		return "http"
	case Https:
		return "https"
	default:
		return "unknown"
	}
}

type OmiProxy struct {
	proxy        *Proxy
	Register     *omiserd.Register
	serverName   string
	address      string
	cache        *omicafe.FileCache
	failCallback func(w http.ResponseWriter, r *http.Request)
	mode         ProxyMode
}

func newOmiProxy(opts *redis.Options, serverName, address string, mode ProxyMode) *OmiProxy {
	var cache *omicafe.FileCache
	var proxy *Proxy
	if mode == DomainMode {
		proxy = NewProxy(NewDomainResolver(opts), cache)
	} else {
		proxy = NewProxy(NewPathResolver(opts), cache)
	}
	omiProxy := &OmiProxy{
		proxy:        proxy,
		Register:     omiserd.NewClient(opts, omiserd.Web).NewRegister(serverName, address),
		serverName:   serverName,
		address:      address,
		mode:         mode,
		cache:        cache,
		failCallback: func(w http.ResponseWriter, r *http.Request) {},
	}
	omiProxy.Register.AddRegisterHandleFunc("cache_state", func() string {
		if omiProxy.cache == nil {
			return "closed"
		} else {
			return "open"
		}
	})
	omiProxy.Register.AddRegisterHandleFunc("cache_max_size", func() string {
		if omiProxy.cache == nil {
			return "0"
		} else {
			return strconv.Itoa(omiProxy.cache.MaxSize)
		}
	})
	omiProxy.Register.AddRegisterHandleFunc("cache_current_size", func() string {
		if omiProxy.cache == nil {
			return "0"
		} else {
			return strconv.Itoa(omiProxy.cache.CurrentSize())
		}
	})
	omiProxy.Register.AddRegisterHandleFunc("cache_hit_count", func() string {
		if omiProxy.cache == nil {
			return "0"
		} else {
			return strconv.Itoa(omiProxy.cache.CacheHitCount)
		}
	})
	omiProxy.Register.AddRegisterHandleFunc("cache_miss_count", func() string {
		if omiProxy.cache == nil {
			return "0"
		} else {
			return strconv.Itoa(omiProxy.cache.CacheMissCount)
		}
	})
	omiProxy.Register.AddRegisterHandleFunc("cache_clear_count", func() string {
		if omiProxy.cache == nil {
			return "0"
		} else {
			return strconv.Itoa(omiProxy.cache.CacheClearNum)
		}
	})
	omiProxy.Register.AddRegisterHandleFunc("cache_num", func() string {
		if omiProxy.cache == nil {
			return "0"
		} else {
			return strconv.Itoa(omiProxy.cache.GetCacheNum())
		}
	})
	omiProxy.Register.AddRegisterHandleFunc("proxy_mode", func() string {
		return mode.String()
	})
	omiProxy.Register.AddMessageHandleFunc(Command_open_cache, func(message string) {
		if size, err := strconv.Atoi(message); err == nil {
			omiProxy.cache = omicafe.NewFileCache("cache", size)
		} else {
			omiProxy.cache = omicafe.NewFileCache("cache", Default_cache_size)
		}
	})
	omiProxy.Register.AddMessageHandleFunc(Command_update_cache_size, func(message string) {
		if size, err := strconv.Atoi(message); err == nil && omiProxy.cache != nil {
			omiProxy.cache.MaxSize = size
		}
	})
	return omiProxy
}

// 处理代理请求
func (omiProxy *OmiProxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	err := omiProxy.proxy.ServeHTTP(w, r)
	// 如果解析失败，调用失败回调
	if err != nil {
		omiProxy.failCallback(w, r)
		return
	}
}

// 设置缓存
func (omiProxy *OmiProxy) SetCache(cacheDir string, maxSize int) {
	omiProxy.cache = omicafe.NewFileCache(cacheDir, maxSize)
	omiProxy.proxy.SetCache(omiProxy.cache)
}

// 设置失败回调
func (omiProxy *OmiProxy) SetFailCallback(failCallback func(w http.ResponseWriter, r *http.Request)) {
	omiProxy.failCallback = failCallback
}

// 启动代理服务
func (omiProxy *OmiProxy) Start(proxyProtocal ProxyProtocal, weight int, cert, key string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		omiProxy.handleRequest(w, r)
	})

	omiProxy.Register.AddRegisterHandleFunc("proxy_protocal", func() string {
		return proxyProtocal.String()
	})

	// 开始监听
	omiProxy.Register.RegisterAndServe(weight, func(port string) {
		address := proxyProtocal.String() + "://" + omiProxy.address
		log.Printf("omi web server: %s is running on %s", omiProxy.serverName, address)

		var err error
		if proxyProtocal == Http {
			err = http.ListenAndServe(port, nil)
		} else if proxyProtocal == Https {
			err = http.ListenAndServeTLS(port, cert, key, nil)
		}

		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	})
}
