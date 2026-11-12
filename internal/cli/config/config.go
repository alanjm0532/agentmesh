package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user configuration directory: %w", err)
	}
	return filepath.Join(dir, "agentmesh", "config.toml"), nil
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

func Set(path, key, value string) (resultErr error) {
	if err := validateSetting(key, value); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("read CLI configuration file %q: %w", path, err)
		}
		initDefaults(v)
	}
	v.Set(key, value)

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create CLI configuration directory: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".agentmesh-config-*.toml")
	if err != nil {
		return fmt.Errorf("create temporary CLI configuration file: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(file.Name()); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			resultErr = errors.Join(resultErr, fmt.Errorf("remove temporary CLI configuration file %q: %w", file.Name(), removeErr))
		}
	}()
	if err = file.Close(); err != nil {
		return fmt.Errorf("close temporary CLI configuration file: %w", err)
	}
	if err = v.WriteConfigAs(file.Name()); err != nil {
		return fmt.Errorf("write CLI configuration file: %w", err)
	}
	if err = os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("replace CLI configuration file %q: %w", path, err)
	}
	return nil
}

func initDefaults(v *viper.Viper) {
	v.Set("timeout", "5s")
}
