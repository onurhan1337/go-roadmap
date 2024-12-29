package models

import "errors"

var (
	ErrNotFound = errors.New("record not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden = errors.New("forbidden")
)