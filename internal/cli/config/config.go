package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

func New() *viper.Viper {
	cfg := viper.New()
	cfg.SetEnvPrefix("AGENTMESH")
	cfg.AutomaticEnv()
	cfg.SetConfigName("cli")
	cfg.SetConfigType("toml")
	cfg.AddConfigPath("deployments")
	return cfg
}

func Load(cfg *viper.Viper, path string, explicit bool) error {
	if explicit {
		cfg.SetConfigFile(path)
	}
	if err := cfg.ReadInConfig(); err != nil {
		// 默认配置文件可省略；显式指定的文件必须能够读取。
		var notFound viper.ConfigFileNotFoundError
		if explicit || !errors.As(err, &notFound) {
			return fmt.Errorf("读取 CLI 配置失败: %w", err)
		}
	}
	return nil
}
