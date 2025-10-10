package domain

import "errors"

// Domain errors represent business rule violations
var (
	// ErrNotFound indicates a requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput indicates the provided input violates business rules
	ErrInvalidInput = errors.New("invalid input")

	// ErrConflict indicates a conflict with existing data (e.g., duplicate email)
	ErrConflict = errors.New("resource conflict")
)
