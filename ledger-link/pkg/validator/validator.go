package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(s interface{}) error {
	return validate.Struct(s)
}