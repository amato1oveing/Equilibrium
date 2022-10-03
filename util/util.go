package util

import (
	"LB/backend"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// HealthCheck 启动健康检查，每两分钟检查一次
func HealthCheck(backends []*backend.Backend) {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			healthCheck(backends)
			log.Println("Health check completed")
		}
	}
}

// HealthCheck 检查后端服务是否存活
func healthCheck(backends []*backend.Backend) {
	for _, b := range backends {
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

// GetPortFromHost 从host中获取端口号
func GetPortFromHost(host string) int {
	sli := strings.Split(host, ":")
	if len(sli) == 2 {
		port, err := strconv.Atoi(sli[1])
		if err != nil {
			log.Println("Error parsing port from host: ", err)
			return 0
		}
		return port
	}
	return 0
}
