/*
	Weight Round Robin
	权重轮询
*/
package server

import (
	"LB/backend"
	"net/http/httputil"
	"net/url"
)

type WeightRoundRobinPool struct {
	backends map[*backend.Backend]int // key: 后端服务, value: 权重
	current  uint64                   // 当前后端服务索引
	Port     int
}

func (w *WeightRoundRobinPool) AddBackend(serverUrl *url.URL, weight int, proxy *httputil.ReverseProxy) {
	w.backends[backend.NewBackend(serverUrl, proxy)] = weight
}

func (w *WeightRoundRobinPool) GetNextPeer() *backend.Backend {
	//TODO implement me
	panic("implement me")
}

func (w *WeightRoundRobinPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for k, _ := range w.backends {
		if (*k).GetURL().String() == backendUrl.String() {
			(*k).SetAlive(alive)
			break
		}
	}
}

func (w *WeightRoundRobinPool) GetBackends() []*backend.Backend {
	var backends []*backend.Backend
	for k, _ := range w.backends {
		backends = append(backends, k)
	}
	return backends
}

func (w *WeightRoundRobinPool) GetPort() int {
	return w.Port
}
