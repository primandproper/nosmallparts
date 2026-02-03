package authentication

import (
	"context"
	"errors"
)

var (
	// ErrInvalidTOTPToken indicates that a provided two-factor code is invalid.
	ErrInvalidTOTPToken = errors.New("invalid two factor code")
)

type (
	// Authenticator authenticates users.
	Authenticator interface {
		Hasher

		CredentialsAreValid(ctx context.Context, hash, password, totpSecret, totpCode string) (bool, error)
	}
)
