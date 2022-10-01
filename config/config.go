package config

import (
	"os"
	"sync"
)

type Config struct {
	// 服务名称
	Name string
	// 服务端口
	Port int
	// 轮询类型
	RoundType string
	// 后端服务地址
	Servers map[string]int
}

var config []*Config

func GetConfig() []*Config {
	return config
}

// NewConfig 创建配置
func NewConfig(filePath string) {
	//todo 从配置文件中读取配置,并初始化
	once := sync.Once{}
	once.Do(func() {
		_, err := os.ReadFile(filePath)
		if err != nil {
			return
		}
	})
}
