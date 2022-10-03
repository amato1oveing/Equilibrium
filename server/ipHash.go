/*
	Ip Hash
	IpHash is a simple load balancer that uses the IP address of the client to
	select a backend server. It is useful when you have a small number of
	clients and a large number of backend servers.
*/
package server

import (
	"LB/backend"
	"net/http/httputil"
	"net/url"
)

type IpHashPool struct {
	backends []*backend.Backend
	ipHashes []string
	port     int
}

func NewIpHashPool(port int) *IpHashPool {
	return &IpHashPool{port: port}
}

func (i *IpHashPool) AddBackend(serverUrl *url.URL, weight int, proxy *httputil.ReverseProxy) {
	i.backends = append(i.backends, backend.NewBackend(serverUrl, proxy))
}

func (i *IpHashPool) GetNextPeer() *backend.Backend {
	//TODO implement me
	panic("implement me")
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
