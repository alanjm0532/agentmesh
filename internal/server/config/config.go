// Package config 管理服务配置。
package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Addr string `mapstructure:"addr" validate:"notblank"`
	Mode string `mapstructure:"mode" validate:"oneof=debug release test"`
}

func (c Config) Validate() error {
	// 根据结构体标签统一校验配置。
	check := validator.New()
	if err := check.RegisterValidation("notblank", validators.NotBlank); err != nil {
		return err
	}
	return check.Struct(c)
}

func Load(path string) (Config, error) {
	// 从指定 TOML 文件读取配置。
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("读取配置文件 %q: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("解析配置文件 %q: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("校验配置文件 %q: %w", path, err)
	}
	return cfg, nil
}
