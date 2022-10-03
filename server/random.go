/*
	Random
	随机访问
*/

package server

import (
	"LB/backend"
	"math/rand"
	"net/http/httputil"
	"net/url"
	"time"
)

type RandomPool struct {
	backends []*backend.Backend
	random   *rand.Rand
	port     int
}

func NewRandomPool(port int) *RandomPool {
	pool := &RandomPool{random: &rand.Rand{}, port: port}
	pool.random.Seed(time.Now().UnixNano())
	return pool
}

func (r *RandomPool) AddBackend(serverUrl *url.URL, weight int, proxy *httputil.ReverseProxy) {
	r.backends = append(r.backends, backend.NewBackend(serverUrl, proxy))
}

func (r *RandomPool) GetNextPeer() *backend.Backend {
	next := r.random.Intn(len(r.backends))
	l := len(r.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(r.backends)
		if (*r.backends[idx]).IsAlive() {
			return r.backends[idx]
		}
	}
	return nil
}

func (r *RandomPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range r.backends {
		if (*b).GetURL().String() == backendUrl.String() {
			(*b).SetAlive(alive)
			break
		}
	}
}

func (r *RandomPool) GetBackends() []*backend.Backend {
	return r.backends
}

func (r *RandomPool) GetPort() int {
	return r.port
}
