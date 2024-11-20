package omiproxy

import (
	"log"
	"net/http"

	"github.com/stormi-li/omicafe-v1"
	"github.com/stormi-li/omiserd-v1"
)

type DNSProxy struct {
	router      *router
	webRegister *omiserd.Register
	serverName  string
	address     string
	cache       *omicafe.FileCache
}

func (proxyServer *DNSProxy) dnsProxyHandleFunc(w http.ResponseWriter, r *http.Request) {
	domainNameResolution(r, proxyServer.router)
	httpForward(w, r, proxyServer.cache)
	websocketForward(w, r)
}

func (proxyServer *DNSProxy) SetCache(cacheDir string, maxSize int) {
	proxyServer.cache = omicafe.NewFileCache(cacheDir, maxSize)
}

func (proxyServer *DNSProxy) Start(proxyProtocal ProxyProtocal, cert, key string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxyServer.dnsProxyHandleFunc(w, r)
	})
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
