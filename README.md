# Omiproxy 反向代理框架
**作者**: stormi-li  
**Email**: 2785782829@qq.com  
## 简介
**Omiproxy** 是一款强大的反向代理框架，能够无缝接入注册中心。它提供了一种更简洁的方式，无需依赖复杂的配置文件来进行代理服务的管理和配置，适用于多种协议，并且支持缓存和域名解析功能。
## 功能
- **支持多种协议**：包括 HTTP、HTTPS、WS、WSS 和 HTTP2 协议。
- **反向代理**：支持将请求转发到目标服务器。
- **域名解析**：自动处理域名解析，无需手动配置。
- **缓存支持**：集成缓存机制，提高性能。
- **接入注册中心**：与注册中心无缝集成，支持动态发现服务和高可用性。
## 教程
### 安装
```shell
go get github.com/stormi-li/omiproxy-v1
```
### 使用
```go
package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/stormi-li/omiproxy-v1"
)

func main() {
	// 创建 Omiproxy 客户端，连接 Redis
	c := omiproxy.NewClient(&redis.Options{Addr: "localhost:6379"})
	
	// 创建反向代理，指定代理名称、目标地址和模式（DomainMode 域名代理，PathMode 路径代理）
	proxy := c.NewProxy("http80", "118.25.196.166:80", omiproxy.DomainMode)
	
	// 设置缓存，缓存大小为 100MB
	proxy.SetCache("cache", 100*1024*1024)
	
	// 启动代理，协议类型为 （HTTP，HTTPS），权重 1，"crt文件路径","key文件路径"
	proxy.Start(omiproxy.Http, 1, "", "")
}
```