package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Server  string        `mapstructure:"server" validate:"required,http_url"`
	Timeout time.Duration `mapstructure:"timeout" validate:"required,gt=0"`
}

func (c Config) Validate() error {
	return validator.New().Struct(c)
}

func Load(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read CLI configuration file %q: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode CLI configuration file %q: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate CLI configuration file %q: %w", path, err)
	}
	return cfg, nil
}
