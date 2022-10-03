package config

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	ServiceName string `json:"service_name"`
	Port        int    `json:"port"`
	RoundType   string `json:"round_type"`
	Nodes       []Node `json:"nodes"`
}

type Node struct {
	Host   string `json:"host"`
	Weight int    `json:"weight"`
}

var config []*Config

func GetConfig() []*Config {
	return config
}

// NewConfig 创建配置
func NewConfig(filePath string) {
	once := sync.Once{}
	once.Do(func() {
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		parseConfig(file)
	})
}

// 解析配置文件,获取每个服务的配置定位
func parseConfig(file *os.File) {
	reader := bufio.NewReader(file)
	var begin []int
	var end []int
	for {
		b, err := reader.ReadBytes('{')
		if err != nil {
			break
		}

		offset := 0
		if len(begin)-1 >= 0 {
			offset = begin[len(begin)-1]
		}
		begin = append(begin, offset+len(b))
	}
	// 重置文件指针到文件头
	file.Seek(0, 0)

	for {
		b, err := reader.ReadBytes('}')
		if err != nil {
			break
		}
		offset := 0
		if len(end)-1 >= 0 {
			offset = end[len(end)-1]
		}
		end = append(end, offset+len(b))
	}
	file.Seek(0, 0)

	confHandler(file, begin, end)
}

func confHandler(file *os.File, begin, end []int) {
	reader := bufio.NewReader(file)
	if len(begin) != len(end) || len(begin) == 0 || len(begin)%2 != 0 {
		panic("配置文件格式错误")
	}

	for i := 0; i < len(begin); i += 2 {
		var cfg Config
		m := make(map[string]string)
		startPos := 0
		if i != 0 {
			startPos = end[i-1]
		}
		buf := make([]byte, begin[i]-startPos)
		_, err := reader.Read(buf)
		if err != nil {
			log.Fatalln("读取配置文件失败,err:", err)
		}
		serviceName := strings.TrimRight(string(buf), "{")
		serviceName = strings.TrimSpace(serviceName)
		cfg.ServiceName = serviceName

		buf = make([]byte, begin[i+1]-begin[i])
		_, err = reader.Read(buf)
		if err != nil {
			log.Fatalln("读取配置文件失败,err:", err)
		}
		body := strings.TrimRight(string(buf), "{")
		bodySli := strings.Split(body, ";")
		if strings.Contains(bodySli[len(bodySli)-1], "nodes") {
			bodySli = bodySli[:len(bodySli)-1]
		}
		for _, v := range bodySli {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			sli := strings.Fields(v)
			if len(sli) != 2 {
				panic("配置文件格式错误")
			}
			m[sli[0]] = sli[1]
		}

		buf = make([]byte, end[i]-begin[i+1])
		_, err = reader.Read(buf)
		if err != nil {
			log.Fatalln("读取配置文件失败,err:", err)
		}
		body = strings.TrimRight(string(buf), "}")
		bodySli = strings.Split(body, ";")
		for _, v := range bodySli {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			sli := strings.Fields(v)
			if len(sli) == 1 {
				cfg.Nodes = append(cfg.Nodes, Node{Host: sli[0]})
			} else if len(sli) == 2 {
				weight, _ := strconv.Atoi(sli[1])
				cfg.Nodes = append(cfg.Nodes, Node{Host: sli[0], Weight: weight})
			}
		}

		buf = make([]byte, end[i+1]-end[i])
		_, err = reader.Read(buf)
		if err != nil {
			log.Fatalln("读取配置文件失败,err:", err)
		}
		body = strings.TrimRight(string(buf), "}")
		bodySli = strings.Split(body, ";")
		if strings.Contains(bodySli[len(bodySli)-1], "nodes") {
			bodySli = bodySli[:len(bodySli)-1]
		}
		for _, v := range bodySli {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			sli := strings.Fields(v)
			if len(sli) != 2 {
				panic("配置文件格式错误")
			}
			m[sli[0]] = sli[1]
		}

		cfg.Port, _ = strconv.Atoi(m["port"])
		cfg.RoundType = m["round_type"]
		config = append(config, &cfg)
	}
}
