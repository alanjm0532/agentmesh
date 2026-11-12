package main

import (
	"flag"
	"os"

	"github.com/alanjm0532/agentmesh/internal/server"
)

func main() {
	configPath := flag.String("config", "deployments/server.toml", "TOML 配置文件路径")
	flag.Parse()

	os.Exit(server.Run(*configPath))
}
