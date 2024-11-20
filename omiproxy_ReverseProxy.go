package omiproxy

import (
	"log"
	"net/http"

	"github.com/stormi-li/omicafe-v1"
	"github.com/stormi-li/omiserd-v1"
)

type ReverseProxy struct {
	router       *router
	webRegister  *omiserd.Register
	serverName   string
	address      string
	cache        *omicafe.FileCache
	failCallback func(w http.ResponseWriter, r *http.Request)
}

func (proxyServer *ReverseProxy) reverseProxyHandleFunc(w http.ResponseWriter, r *http.Request) {
	if pathRequestResolution(r, proxyServer.router) != nil {
		proxyServer.failCallback(w, r)
		return
	}
	httpForward(w, r, proxyServer.cache)
	websocketForward(w, r)
}

func (proxyServer *ReverseProxy) SetCache(cacheDir string, maxSize int) {
	proxyServer.cache = omicafe.NewFileCache(cacheDir, maxSize)
}

func (proxyServer *ReverseProxy) SetFailCallback(failCallback func(w http.ResponseWriter, r *http.Request)) {
	proxyServer.failCallback = failCallback
}

func (proxyServer *ReverseProxy) Start(weight int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxyServer.reverseProxyHandleFunc(w, r)
	})
	proxyServer.webRegister.RegisterAndListen(1, func(port string) {
		log.Println("omi web server: " + proxyServer.serverName + " is running on http://" + proxyServer.address)
		err := http.ListenAndServe(port, nil)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	})
}
