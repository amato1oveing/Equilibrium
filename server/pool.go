package server

import (
	"LB/backend"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	Robin  = "robin"
	Weight = "weight"
	IpHash = "iphash"
	Random = "random"
)

type ServerPool interface {
	AddBackend(serverUrl *url.URL, weight int, proxy *httputil.ReverseProxy)
	GetNextPeer(r *http.Request) *backend.Backend
	MarkBackendStatus(backendUrl *url.URL, alive bool)
	GetBackends() []*backend.Backend
	GetPort() int
}

func NewServerPool(roundType string, port int) *ServerPool {
	var pool ServerPool
	switch roundType {
	case Robin:
		pool = NewRoundRobinPool(port)
	case Weight:
		pool = NewWeightRoundRobinPool(port)
	case IpHash:
		pool = NewIpHashPool(port)
	case Random:
		pool = NewRandomPool(port)
	default:
		panic("unsupported round type")
	}
	return &pool
}
