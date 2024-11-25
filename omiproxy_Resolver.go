package omiproxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omiserd-v1"
)

// 定义接口用于服务地址解析
type Resolver interface {
	Resolve(r *http.Request) (string, error)
}

// PathResolver 基于路径的请求解析器
type PathResolver struct {
	router *router
}

func (pr *PathResolver) Resolve(r *http.Request) (string, error) {
	serverName := strings.Split(r.URL.Path, "/")[1]
	if !pr.router.Has(serverName) {
		return "", fmt.Errorf("未找到微服务: %s", serverName)
	}
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+serverName)
	return pr.router.getAddress(serverName), nil
}

// DomainResolver 基于域名的请求解析器
type DomainResolver struct {
	router *router
}

func (dr *DomainResolver) Resolve(r *http.Request) (string, error) {
	domainName := strings.Split(r.Host, ":")[0]
	if !dr.router.Has(domainName) {
		return "", fmt.Errorf("未找到 Web 服务: %s", domainName)
	}
	return dr.router.getAddress(domainName), nil
}

func NewPathResolver(opts *redis.Options) *PathResolver {
	return &PathResolver{
		router: newRouter(opts, omiserd.Server),
	}
}

func NewDomainResolver(opts *redis.Options) *DomainResolver {
	return &DomainResolver{
		router: newRouter(opts, omiserd.Web),
	}
}
