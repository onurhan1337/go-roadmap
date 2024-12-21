package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
	Validate.RegisterValidation("isbn", validateISBN)
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

type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

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