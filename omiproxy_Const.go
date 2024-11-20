package omiproxy

import (
	"time"
)

const router_refresh_interval = 2 * time.Second

type ProxyType int
type ProxyProtocal int

var Http ProxyProtocal = 1
var Https ProxyProtocal = 2
