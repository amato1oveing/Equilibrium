package server

import (
	"LB/config"
	"LB/util"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	Attempts int = iota // 尝试次数
	Retry               // 重试次数
)

var serverPoolList []*ServerPool

// NewServerPool 初始化ServerPool,并返回所在的索引
func addServerPool(roundType string, port int) *ServerPool {
	serverPool := NewServerPool(roundType, port)
	serverPoolList = append(serverPoolList, serverPool)
	return serverPool
}

// GetAttemptsFromContext 返回请求的尝试次数
func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

// GetRetryFromContext 返回重试次数
func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

func GetServerPool(port int) *ServerPool {
	for i, pool := range serverPoolList {
		if (*pool).GetPort() == port {
			return serverPoolList[i]
		}
	}
	log.Printf("No server pool found for port %d\n", port)
	return nil
}

// lb 负载均衡
func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	serverPool := GetServerPool(util.GetPortFromHost(r.Host))
	if serverPool == nil {
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}
	peer := (*serverPool).GetNextPeer()
	if peer != nil {
		peer.GetProxy().ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

// Start 开启一个负载均衡服务
func Start(config *config.Config) {
	serverPool := addServerPool(config.RoundType, config.Port)
	// 解析后端服务器地址
	for _, node := range config.Nodes {
		serverUrl, err := url.Parse(node.Host)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
			retries := GetRetryFromContext(request)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(request.Context(), Retry, retries+1)
					proxy.ServeHTTP(writer, request.WithContext(ctx))
				}
				return
			}

			// 重试3次后，将此后端标记为不可用
			(*serverPool).MarkBackendStatus(serverUrl, false)

			// 如果同一请求在不同后端之间路由，增加计数
			attempts := GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lb(writer, request.WithContext(ctx))
		}

		(*serverPool).AddBackend(serverUrl, node.Weight, proxy)
		log.Printf("Configured server: %s\n", serverUrl)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: http.HandlerFunc(lb),
	}

	// 启动健康检查
	go util.HealthCheck((*serverPool).GetBackends())

	log.Printf("Load Balancer started at :%d\n", config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
