package omiproxy

import (
	"math/rand/v2"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omiserd-v1"
	omiconst "github.com/stormi-li/omiserd-v1/omiserd_const"
	discover "github.com/stormi-li/omiserd-v1/omiserd_discover"
)

const router_refresh_interval = 2 * time.Second

type router struct {
	discover   *discover.Discover
	addressMap map[string][]string
	mutex      sync.Mutex
}

func newRouter(opts *redis.Options, nodeType omiconst.NodeType) *router {
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
	nodeMap := router.discover.GetAll()
	addrMap := map[string][]string{}
	for name, addrs := range nodeMap {
		for _, addr := range addrs {
			data := router.discover.GetData(name, addr)
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
