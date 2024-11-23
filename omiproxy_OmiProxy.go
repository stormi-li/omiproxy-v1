package omiproxy

import (
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omicafe-v1"
	"github.com/stormi-li/omiserd-v1"
	omiconst "github.com/stormi-li/omiserd-v1/omiserd_const"
	register "github.com/stormi-li/omiserd-v1/omiserd_register"
)

type ProxyMode int

const (
	DomainMode ProxyMode = iota
	PathMode
)

type ProxyProtocal int

var Http ProxyProtocal = 1
var Https ProxyProtocal = 2

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
	Register     *register.Register
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
	return &OmiProxy{
		proxy:        proxy,
		Register:     omiserd.NewClient(opts, omiconst.Web).NewRegister(serverName, address),
		serverName:   serverName,
		address:      address,
		mode:         mode,
		cache:        cache,
		failCallback: func(w http.ResponseWriter, r *http.Request) {},
	}
}

// 处理代理请求
func (proxy *OmiProxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	err := proxy.proxy.ServeHTTP(w, r)
	// 如果解析失败，调用失败回调
	if err != nil {
		proxy.failCallback(w, r)
		return
	}
}

// 设置缓存
func (proxy *OmiProxy) SetCache(cacheDir string, maxSize int) {
	proxy.cache = omicafe.NewFileCache(cacheDir, maxSize)
	proxy.proxy.SetCache(proxy.cache)
}

// 设置失败回调
func (proxy *OmiProxy) SetFailCallback(failCallback func(w http.ResponseWriter, r *http.Request)) {
	proxy.failCallback = failCallback
}

// 启动代理服务
func (proxy *OmiProxy) Start(proxyProtocal ProxyProtocal, weight int, cert, key string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.handleRequest(w, r)
	})

	// 注册代理信息
	cacheInfoHandler(proxy.Register, proxy.cache)
	if proxy.mode == DomainMode {
		proxy.Register.Data["proxy_type"] = "domain_name_proxy"
	} else {
		proxy.Register.Data["proxy_type"] = "reverse_proxy"
	}

	// 开始监听
	proxy.Register.RegisterAndServe(weight, func(port string) {
		address := proxyProtocal.String() + "://" + proxy.address
		log.Printf("omi web server: %s is running on %s", proxy.serverName, address)

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
