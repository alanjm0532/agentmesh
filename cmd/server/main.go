package main

import (
	"flag"
	"os"
)

func main() {
	configPath := flag.String("config", "configs/config.toml", "TOML 配置文件路径")
	flag.Parse()

	os.Exit(run(*configPath))
}
