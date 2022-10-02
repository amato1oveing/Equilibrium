/*
	Weight Round Robin
	权重轮询
*/
package server

import (
	"LB/backend"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type WeightRoundRobinPool struct {
	backends    []*backend.Backend // 后端服务列表
	weights     []int              // 后端服务权重列表
	weightTotal int                // 后端服务权重总和
	current     uint64             // 当前后端服务索引
	port        int
	mux         sync.RWMutex
}

func (w *WeightRoundRobinPool) AddBackend(serverUrl *url.URL, weight int, proxy *httputil.ReverseProxy) {
	w.backends = append(w.backends, backend.NewBackend(serverUrl, proxy))
	w.weights = append(w.weights, weight)
	w.weightTotal += weight
}

func (w *WeightRoundRobinPool) GetNextPeer() *backend.Backend {
	//根据权重选择下一个后端服务
	next := int(atomic.AddUint64(&w.current, uint64(1)) % uint64(w.weightTotal))
	for i := 0; i < len(w.weights)*2; i++ { // 避免无限循环
		idx := i % len(w.weights) // 避免数组越界
		next -= w.weights[idx]    // 减去当前权重，如果小于0，则表示当前后端服务被选中
		if next < 0 {
			// 判断当前后端服务是否可用
			if (*w.backends[idx]).IsAlive() {
				// 如果后端服务可用，则返回
				return w.backends[idx]
			}
			// 如果后端服务不可用，则将当前权重设置为0，避免被选中
			w.setFailedBackendWeight0(idx)
		}
	}
	return nil
}

func (w *WeightRoundRobinPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for i, k := range w.backends {
		if (*k).GetURL().String() == backendUrl.String() {
			(*k).SetAlive(alive)
			if !alive {
				w.setFailedBackendWeight0(i)
			}
			break
		}
	}
}

// 将失败的后端服务权重设置为0
func (w *WeightRoundRobinPool) setFailedBackendWeight0(index int) {
	if w.weights[index] == 0 {
		return
	}

	w.mux.Lock()
	defer w.mux.Unlock()
	w.weightTotal -= w.weights[index]
	w.weights[index] = 0
}

func (w *WeightRoundRobinPool) GetBackends() []*backend.Backend {
	return w.backends
}

func (w *WeightRoundRobinPool) GetPort() int {
	return w.port
}
