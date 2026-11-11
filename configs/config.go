// Package configs 管理服务配置。
package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Addr string `mapstructure:"addr"`
}

// Load 从指定 TOML 文件读取配置。
func Load(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")
	v.SetDefault("server.addr", "127.0.0.1:8080")
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("读取配置文件 %q: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("解析配置文件 %q: %w", path, err)
	}
	if strings.TrimSpace(cfg.Server.Addr) == "" {
		return Config{}, fmt.Errorf("配置 server.addr 不能为空")
	}
	return cfg, nil
}
