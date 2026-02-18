package internalerrors

import (
	"errors"
	"fmt"
)

var (
	// ErrCircuitBroken is returned when a circuit breaker has tripped.
	ErrCircuitBroken = errors.New("service circuit broken")
	// ErrNilInputParameter is returned when an input parameter is nil.
	ErrNilInputParameter = errors.New("provided input parameter is nil")
	// ErrEmptyInputParameter is returned when an input parameter is empty.
	ErrEmptyInputParameter = errors.New("provided input parameter is empty")
)

// NilConfigError returns a nil config error.
func NilConfigError(name string) error {
	return fmt.Errorf("nil config provided for %s", name)
}
