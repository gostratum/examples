package usecase

import "errors"

// Typed errors for clean HTTP mapping
var (
	ErrUnavailable = errors.New("service unavailable")
	ErrNotFound    = errors.New("resource not found")
	ErrInvalid     = errors.New("invalid input")
)
