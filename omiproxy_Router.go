package omiproxy

import (
	"math/rand/v2"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omiserd-v1"
)

type router struct {
	discover   *omiserd.Discover
	addressMap map[string][]string
	mutex      sync.Mutex
}

func newRouter(opts *redis.Options, nodeType omiserd.NodeType) *router {
	router := router{
		discover:   omiserd.NewClient(opts, nodeType).NewDiscover(),
		addressMap: map[string][]string{},
		mutex:      sync.Mutex{},
	}
	go func() {
		for {
			router.refresh()
			time.Sleep(router_refresh_interval)
		}
	}()
	return &router
}

func (router *router) refresh() {
	nodeMap := router.discover.DiscoverAllServers()
	addrMap := map[string][]string{}
	for name, addrs := range nodeMap {
		for addr, data := range addrs {
			weight, _ := strconv.Atoi(data["weight"])
			for i := 0; i < weight; i++ {
				addrMap[name] = append(addrMap[name], addr)
			}
		}
	}
	router.mutex.Lock()
	router.addressMap = addrMap
	router.mutex.Unlock()
}

func (router *router) getAddress(serverName string) string {
	router.mutex.Lock()
	defer router.mutex.Unlock()
	if len(router.addressMap[serverName]) == 0 {
		return ""
	}
	return router.addressMap[serverName][rand.IntN(len(router.addressMap[serverName]))]
}

func (router *router) Has(serverName string) bool {
	return len(router.addressMap[serverName]) != 0
}
