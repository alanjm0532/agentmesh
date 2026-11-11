// Package configs 管理服务配置。
package configs

import "os"

type Config struct {
	Addr string
}

// Load 从环境变量读取配置，未设置时使用本地演示地址。
func Load() Config {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	return Config{Addr: addr}
}
