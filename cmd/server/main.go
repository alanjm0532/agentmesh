package main

import (
	"flag"
	"log"
)

func main() {
	configPath := flag.String("config", "configs/config.toml", "TOML 配置文件路径")
	flag.Parse()

	server, err := wireServer(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	if err = server.Run(); err != nil {
		log.Fatal(err)
	}
}
