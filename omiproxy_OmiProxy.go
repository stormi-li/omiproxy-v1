package omiproxy

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omicafe-v1"
	"github.com/stormi-li/omicert-v1"
	"github.com/stormi-li/omiserd-v1"
)

const Command_open_cache = "open_cache"
const Command_update_cache_size = "update_cache_size"
const Default_cache_size = 100 * 1024 * 1024

type OmiProxy struct {
	Options            *redis.Options
	proxy              *Proxy
	Register           *omiserd.Register
	serverName         string
	address            string
	cache              *omicafe.FileCache
	InsecureSkipVerify bool
	BeforeCallback     []func(w http.ResponseWriter, r *http.Request, handFuncResponse HandleFuncResponse) string
	AfterCallback      []func(w http.ResponseWriter, r *http.Request, handFuncResponse HandleFuncResponse) string
	FailCallback       []func(w http.ResponseWriter, r *http.Request, handFuncResponse HandleFuncResponse) string
}

func newOmiProxy(opts *redis.Options, serverName, address string) *OmiProxy {
	var cache *omicafe.FileCache
	omiProxy := &OmiProxy{
		Options:        opts,
		Register:       omiserd.NewClient(opts, omiserd.Web).NewRegister(serverName, address),
		serverName:     serverName,
		address:        address,
		cache:          cache,
		BeforeCallback: []func(w http.ResponseWriter, r *http.Request, handFuncResponse HandleFuncResponse) string{},
		AfterCallback:  []func(w http.ResponseWriter, r *http.Request, handFuncResponse HandleFuncResponse) string{},
		FailCallback:   []func(w http.ResponseWriter, r *http.Request, handFuncResponse HandleFuncResponse) string{},
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

// 设置缓存
func (omiProxy *OmiProxy) SetCache(cacheDir string, maxSize int) {
	omiProxy.cache = omicafe.NewFileCache(cacheDir, maxSize)
}

type HandleFuncResponse struct {
	Continue string
	Break    string
}

func (omiProxy *OmiProxy) AddBeforeCallback(callback func(w http.ResponseWriter, r *http.Request, handFuncResponse HandleFuncResponse) string) {
	omiProxy.BeforeCallback = append(omiProxy.BeforeCallback, callback)
}

func (omiProxy *OmiProxy) AddAfterCallback(callback func(w http.ResponseWriter, r *http.Request, handFuncResponse HandleFuncResponse) string) {
	omiProxy.AfterCallback = append(omiProxy.AfterCallback, callback)
}

func (omiProxy *OmiProxy) AddForwardFailCallback(callback func(w http.ResponseWriter, r *http.Request, handFuncResponse HandleFuncResponse) string) {
	omiProxy.AfterCallback = append(omiProxy.AfterCallback, callback)
}

func (omiProxy *OmiProxy) SetCORS() {
	omiProxy.AddBeforeCallback(func(w http.ResponseWriter, r *http.Request, handFuncResponse HandleFuncResponse) string {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusOK)
			return handFuncResponse.Break
		}
		return handFuncResponse.Continue
	})
}

// 启动代理服务
func (omiProxy *OmiProxy) Start(weight int, crediential *omicert.Credential) {
	omiProxy.proxy = NewProxy(omiProxy.Options, omiserd.Web, omiProxy.cache, omiProxy.InsecureSkipVerify)
	handleFuncResponse := HandleFuncResponse{
		Continue: "Continue",
		Break:    "Break",
	}
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for _, callback := range omiProxy.BeforeCallback {
			if callback(w, r, handleFuncResponse) == handleFuncResponse.Break {
				return
			}
		}
		err := omiProxy.proxy.Forward(w, r)
		// 如果解析失败，调用失败回调
		if err != nil {
			for _, callback := range omiProxy.FailCallback {
				if callback(w, r, handleFuncResponse) == handleFuncResponse.Break {
					return
				}
			}
		}
		for _, callback := range omiProxy.AfterCallback {
			if callback(w, r, handleFuncResponse) == handleFuncResponse.Break {
				return
			}
		}
	})

	protocal := "http"
	if crediential != nil {
		protocal = "https"
	}

	omiProxy.Register.AddRegisterHandleFunc("protocal", func() string {
		return protocal
	})

	// 开始监听
	omiProxy.Register.RegisterAndServe(weight, func(port string) {
		address := protocal + "://" + omiProxy.address
		log.Printf("omi web server: %s is running on %s", omiProxy.serverName, address)
		if protocal == "http" {
			err := http.ListenAndServe(port, nil)
			if err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}
		} else if protocal == "https" {
			omicert.ListenAndServeTLS(port, crediential)
		}
	})
}
