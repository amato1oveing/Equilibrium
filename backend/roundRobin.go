package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type RobinBackend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *RobinBackend) GetURL() *url.URL {
	return b.URL
}

func (b *RobinBackend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

func (b *RobinBackend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *RobinBackend) GetProxy() *httputil.ReverseProxy {
	return b.ReverseProxy
}
