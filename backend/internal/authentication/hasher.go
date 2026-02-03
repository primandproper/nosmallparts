package authentication

import (
	"context"
	"errors"
)

var (
	// ErrPasswordDoesNotMatch indicates that a provided passwords does not match.
	ErrPasswordDoesNotMatch = errors.New("password does not match")
)

type (
	// Hasher hashes passwords.
	Hasher interface {
		HashPassword(ctx context.Context, password string) (string, error)
	}
)
