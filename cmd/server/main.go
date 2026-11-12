package main

import (
	"flag"
	"os"

	"github.com/alanjm0532/agentmesh/internal/server"
)

func main() {
	configPath := flag.String("config", "deployments/server.toml", "TOML configuration file path")
	flag.Parse()

	os.Exit(server.Run(*configPath))
}
