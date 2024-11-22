package omiproxy

import (
	"log"
	"net/http"

	"github.com/stormi-li/omicafe-v1"
	"github.com/stormi-li/omiserd-v1"
)

type DomainNameProxy struct {
	router       *router
	webRegister  *omiserd.Register
	serverName   string
	address      string
	cache        *omicafe.FileCache
	failCallback func(w http.ResponseWriter, r *http.Request)
}

func (proxyServer *DomainNameProxy) dnsProxyHandleFunc(w http.ResponseWriter, r *http.Request) {
	if domainNameResolution(r, proxyServer.router) != nil {
		proxyServer.failCallback(w, r)
		return
	}
	httpForward(w, r, proxyServer.cache)
	websocketForward(w, r)
}

func (proxyServer *DomainNameProxy) SetCache(cacheDir string, maxSize int) {
	proxyServer.cache = omicafe.NewFileCache(cacheDir, maxSize)
}

func (proxyServer *DomainNameProxy) SetFailCallback(failCallback func(w http.ResponseWriter, r *http.Request)) {
	proxyServer.failCallback = failCallback
}

func (proxyServer *DomainNameProxy) Start(proxyProtocal ProxyProtocal, cert, key string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxyServer.dnsProxyHandleFunc(w, r)
	})
	registerHandler(proxyServer.webRegister, proxyServer.cache)
	proxyServer.webRegister.Data["proxy_type"] = "domain_name_proxy"
	proxyServer.webRegister.RegisterAndListen(1, func(port string) {
		if proxyProtocal == Http {
			log.Println("omi web server: " + proxyServer.serverName + " is running on http://" + proxyServer.address)
			err := http.ListenAndServe(port, nil)
			if err != nil {
				log.Fatalf("Failed to start server: %v", err)
			}
		}
		if proxyProtocal == Https {
			proxyServer.webRegister.RegisterAndListen(1, func(port string) {
				log.Println("omi web server: " + proxyServer.serverName + " is running on https://" + proxyServer.address)
				err := http.ListenAndServeTLS(port, cert, key, nil)
				if err != nil {
					log.Fatalf("Failed to start server: %v", err)
				}
			})
		}
	})
}