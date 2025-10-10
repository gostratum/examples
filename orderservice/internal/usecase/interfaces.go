package usecase

import (
	"errors"

	"github.com/gostratum/examples/orderservice/internal/domain"
)

// Application-level errors for use case layer
// These are used to communicate failures to the presentation layer
var (
	// ErrUnavailable indicates the service is temporarily unavailable (infrastructure failure)
	ErrUnavailable = errors.New("service unavailable")

	// ErrNotFound wraps domain.ErrNotFound for application layer
	ErrNotFound = domain.ErrNotFound

	// ErrInvalid wraps domain.ErrInvalidInput for application layer
	ErrInvalid = domain.ErrInvalidInput

	// ErrConflict wraps domain.ErrConflict for application layer
	ErrConflict = domain.ErrConflict
)
