package server

import (
	"LB/backend"
	"net/http/httputil"
	"net/url"
)

const (
	Robin  = "robin"
	Weight = "weight"
)

type ServerPool interface {
	AddBackend(serverUrl *url.URL, weight int, proxy *httputil.ReverseProxy)
	GetNextPeer() *backend.Backend
	MarkBackendStatus(backendUrl *url.URL, alive bool)
	GetBackends() []*backend.Backend
	GetPort() int
}

func NewServerPool(roundType string, port int) *ServerPool {
	var pool ServerPool
	switch roundType {
	case Robin:
		pool = &RoundRobinPool{port: port}
	case Weight:
		pool = &WeightRoundRobinPool{port: port}
	default:
		panic("unsupported round type")
	}
	return &pool
}
