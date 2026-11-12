package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

var settingValidators = map[string]func(string) error{
	"server":  validateServer,
	"timeout": validateTimeout,
}

func validateSetting(key, value string) error {
	validate, ok := settingValidators[key]
	if !ok {
		return fmt.Errorf("unsupported setting %q", key)
	}
	return validate(value)
}

func validateServer(value string) error {
	if err := validator.New().Var(value, "required,http_url"); err != nil {
		return fmt.Errorf("invalid server URL: expected an HTTP or HTTPS URL: %w", err)
	}
	return nil
}

func validateTimeout(value string) error {
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fmt.Errorf("invalid timeout %q: expected a positive Go duration, such as 5s", value)
	}
	return nil
}
