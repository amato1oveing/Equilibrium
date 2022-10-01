package main

import (
	"LB/backend"
	"LB/config"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Attempts int = iota // 尝试次数
	Retry               // 重试次数
)

type ServerPool struct {
	backends  []*backend.Backend // 后端服务列表
	current   uint64             // 当前后端服务索引
	roundType string             // 轮询类型
}

var serverPool *ServerPool

// NewServerPool 初始化ServerPool
func NewServerPool(roundType string) {
	once := sync.Once{}
	once.Do(func() {
		serverPool = &ServerPool{roundType: roundType}
	})
}

// AddBackend 添加后端服务
func (s *ServerPool) AddBackend(serverUrl *url.URL, proxy *httputil.ReverseProxy) {
	s.backends = append(s.backends, backend.NewBackend(s.roundType, serverUrl, proxy))
}

// NextIndex 获取下一个后端服务的索引
func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

// MarkBackendStatus 标记后端服务状态
func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if (*b).GetURL().String() == backendUrl.String() {
			(*b).SetAlive(alive)
			break
		}
	}
}

// GetNextPeer 获取下一个可用的后端服务
func (s *ServerPool) GetNextPeer() backend.Backend {
	next := s.NextIndex()
	l := len(s.backends) + next // 避免无限循环
	for i := next; i < l; i++ {
		idx := i % len(s.backends) // 避免数组越界
		// 如果后端服务可用，则返回
		if (*s.backends[idx]).IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx)) // 更新当前后端服务索引
			}
			return *s.backends[idx]
		}
	}
	return nil
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

// lb 负载均衡
func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := serverPool.GetNextPeer()
	if peer != nil {
		peer.GetProxy().ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

// HealthCheck 检查后端服务是否存活
func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := isBackendAlive((*b).GetURL())
		(*b).SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", (*b).GetURL(), status)
	}
}

// 检查后端服务是否存活
func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	defer conn.Close()
	return true
}

// healthCheck 启动健康检查，每两分钟检查一次
func healthCheck() {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			serverPool.HealthCheck()
			log.Println("Health check completed")
		}
	}
}

// 开启一个负载均衡服务
func start(config *config.Config) {
	NewServerPool(config.RoundType)
	// 解析后端服务器地址
	for tok, _ := range config.Servers {
		serverUrl, err := url.Parse(tok)
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
			serverPool.MarkBackendStatus(serverUrl, false)

			// 如果同一请求在不同后端之间路由，增加计数
			attempts := GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lb(writer, request.WithContext(ctx))
		}

		serverPool.AddBackend(serverUrl, proxy)
		log.Printf("Configured server: %s\n", serverUrl)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: http.HandlerFunc(lb),
	}

	// 启动健康检查
	go healthCheck()

	log.Printf("Load Balancer started at :%d\n", config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "config", "config-example.conf", "config file path")
	flag.Parse()

	config.NewConfig(filePath)
	configs := config.GetConfig()

	for _, cfg := range configs {
		log.Printf("Starting server: %s\n", cfg.Name)
		go start(cfg)
	}

	select {}
}
