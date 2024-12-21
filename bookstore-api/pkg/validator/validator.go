package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

type BookValidator struct {
	validator *validator.Validate
}

func NewBookValidator() *BookValidator {
	v := validator.New()
	v.RegisterValidation("isbn", validateISBN)
	return &BookValidator{validator: v}
}

func (v *BookValidator) Validate(i interface{}) []ValidationError {
	var errors []ValidationError

	err := v.validator.Struct(i)
	if err == nil {
		return nil
	}

	validationErrors := err.(validator.ValidationErrors)
	for _, e := range validationErrors {
		errors = append(errors, ValidationError{
			Field: e.Field(),
			Tag:   e.Tag(),
			Value: e.Param(),
		})
	}

	return errors
}

func validateISBN(fl validator.FieldLevel) bool {
	isbn := fl.Field().String()
	isbnRegex := regexp.MustCompile(`^(?:\d{3}-?)?\d{1,5}-?\d{1,7}-?\d{1,6}-?\d$`)
	cleanISBN := regexp.MustCompile(`-`).ReplaceAllString(isbn, "")

	if !isbnRegex.MatchString(isbn) || len(cleanISBN) != 13 {
		return false
	}

	var sum int
	for i, digit := range cleanISBN[:12] {
		digit := int(digit - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	checksum := (10 - (sum % 10)) % 10
	return int(cleanISBN[12]-'0') == checksum
}