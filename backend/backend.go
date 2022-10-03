package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL          *url.URL
	alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func NewBackend(serverUrl *url.URL, proxy *httputil.ReverseProxy) *Backend {
	return &Backend{URL: serverUrl, alive: true, ReverseProxy: proxy}
}

func (b *Backend) GetURL() *url.URL {
	return b.URL
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.alive
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.alive = alive
	b.mux.Unlock()
}

func (b *Backend) GetProxy() *httputil.ReverseProxy {
	return b.ReverseProxy
}
