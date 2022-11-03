/*
	Ip Hash
	IpHash is a simple load balancer that uses the IP address of the client to
	select a backend server. It is useful when you have a small number of
	clients and a large number of backend servers.
*/
package server

import (
	"LB/backend"
	"crypto/md5"
	"io"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type IpHashPool struct {
	backends []*backend.Backend
	ipHashes map[string]int
	port     int
}

func NewIpHashPool(port int) *IpHashPool {
	return &IpHashPool{port: port}
}

func (i *IpHashPool) AddBackend(serverUrl *url.URL, weight int, proxy *httputil.ReverseProxy) {
	i.backends = append(i.backends, backend.NewBackend(serverUrl, proxy))
}

func (i *IpHashPool) GetNextPeer(r *http.Request) *backend.Backend {
	return i.getNextPeer(r.RemoteAddr)
}

func (i *IpHashPool) getNextPeer(ip string) *backend.Backend {
	hash := md5.New()
	io.WriteString(hash, ip)
	hashed := hash.Sum(nil)
	if v, ok := i.ipHashes[string(hashed)]; ok {
		if (*i.backends[v]).IsAlive() {
			return i.backends[v]
		}
		i.getNextPeer(string(hashed))
	} else {
		intn := rand.Intn(len(i.backends))
		l := len(i.backends) + intn
		for j := intn; j < l; j++ {
			idx := j % len(i.backends)
			if (*i.backends[idx]).IsAlive() {
				i.ipHashes[string(hashed)] = idx
				return i.backends[idx]
			}
		}
	}
	return nil
}

func (i *IpHashPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range i.backends {
		if (*b).GetURL().String() == backendUrl.String() {
			(*b).SetAlive(alive)
			break
		}
	}
}

func (i *IpHashPool) GetBackends() []*backend.Backend {
	return i.backends
}

func (i *IpHashPool) GetPort() int {
	return i.port
}
