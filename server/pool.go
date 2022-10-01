package server

import (
	"LB/backend"
	"net/http/httputil"
	"net/url"
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
	case "robin":
		pool = &RoundRobinPool{Port: port}
	case "weight":
		pool = &WeightRoundRobinPool{Port: port, backends: make(map[*backend.Backend]int)}
	default:
		panic("unsupported round type")
	}
	return &pool
}
