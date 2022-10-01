package backend

import (
	"net/http/httputil"
	"net/url"
)

type Backend interface {
	// GetURL returns the URL of the backend
	GetURL() *url.URL
	// IsAlive returns true if the backend is alive
	IsAlive() bool
	// SetAlive sets the status of the backend
	SetAlive(alive bool)
	// GetProxy returns the proxy of the backend
	GetProxy() *httputil.ReverseProxy
}

func NewBackend(roundType string, serverUrl *url.URL, proxy *httputil.ReverseProxy) *Backend {
	var b Backend
	switch roundType {
	case "Robin":
		b = &RobinBackend{URL: serverUrl, Alive: true, ReverseProxy: proxy}
	default:
		b = &RobinBackend{URL: serverUrl, Alive: true, ReverseProxy: proxy}
	}
	return &b
}
