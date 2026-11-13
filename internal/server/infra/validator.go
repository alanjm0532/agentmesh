package infra

import (
	"reflect"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
)

var configureValidation = sync.OnceFunc(func() {
	validate, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		panic("Gin validator engine does not support custom validation rules")
	}
	if err := validate.RegisterValidation("notblank", validators.NotBlank); err != nil {
		panic(err)
	}
	if err := validate.RegisterValidation("maxbytes", func(fl validator.FieldLevel) bool {
		limit, err := strconv.Atoi(fl.Param())
		// Bcrypt limits passwords by bytes, while the built-in max counts runes.
		return err == nil && fl.Field().Kind() == reflect.String && len(fl.Field().String()) <= limit
	}); err != nil {
		panic(err)
	}
})
