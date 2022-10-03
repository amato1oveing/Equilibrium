/*
	Round Robin
	普通轮询
*/
package server

import (
	"LB/backend"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type RoundRobinPool struct {
	backends []*backend.Backend // 后端服务列表
	current  uint64             // 当前后端服务索引
	port     int
}

func NewRoundRobinPool(port int) *RoundRobinPool {
	return &RoundRobinPool{port: port}
}

// AddBackend 添加后端服务
func (s *RoundRobinPool) AddBackend(serverUrl *url.URL, weight int, proxy *httputil.ReverseProxy) {
	s.backends = append(s.backends, backend.NewBackend(serverUrl, proxy))
}

// NextIndex 获取下一个后端服务的索引
func (s *RoundRobinPool) nextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

// GetNextPeer 获取下一个可用的后端服务
func (s *RoundRobinPool) GetNextPeer() *backend.Backend {
	next := s.nextIndex()
	l := len(s.backends) + next // 避免无限循环
	for i := next; i < l; i++ {
		idx := i % len(s.backends) // 避免数组越界
		// 如果后端服务可用，则返回
		if (*s.backends[idx]).IsAlive() {
			return s.backends[idx]
		}
		atomic.AddUint64(&s.current, uint64(1)) // 后端服务不可用，索引加1
	}
	return nil
}

// MarkBackendStatus 标记后端服务状态
func (s *RoundRobinPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if (*b).GetURL().String() == backendUrl.String() {
			(*b).SetAlive(alive)
			break
		}
	}
}

func (s *RoundRobinPool) GetBackends() []*backend.Backend {
	return s.backends
}

func (s *RoundRobinPool) GetPort() int {
	return s.port
}
