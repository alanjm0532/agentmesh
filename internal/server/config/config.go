// Package config loads and validates server configuration.
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
	// Validate configuration using struct tags.
	check := validator.New()
	if err := check.RegisterValidation("notblank", validators.NotBlank); err != nil {
		return err
	}
	return check.Struct(c)
}

func Load(path string) (Config, error) {
	// Read configuration from the specified TOML file.
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read configuration file %q: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode configuration file %q: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate configuration file %q: %w", path, err)
	}
	return cfg, nil
}
