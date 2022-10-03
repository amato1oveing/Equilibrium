package main

import (
	"LB/config"
	"LB/server"
	"flag"
	"log"
)

func main() {
	var filePath string
	flag.StringVar(&filePath, "config", "config-example.conf", "config file path")
	flag.Parse()

	config.NewConfig(filePath)
	configs := config.GetConfig()

	for _, cfg := range configs {
		log.Printf("Starting server: %s\n", cfg.ServiceName)
		go server.Start(cfg)
	}

	select {}
}
