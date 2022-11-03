/*
	Random
	随机访问
*/

package server

import (
	"LB/backend"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type RandomPool struct {
	backends []*backend.Backend
	port     int
}

func NewRandomPool(port int) *RandomPool {
	pool := &RandomPool{port: port}
	rand.Seed(time.Now().UnixNano())
	return pool
}

func (r *RandomPool) AddBackend(serverUrl *url.URL, weight int, proxy *httputil.ReverseProxy) {
	r.backends = append(r.backends, backend.NewBackend(serverUrl, proxy))
}

func (r *RandomPool) GetNextPeer(_ *http.Request) *backend.Backend {
	next := rand.Intn(len(r.backends))
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
