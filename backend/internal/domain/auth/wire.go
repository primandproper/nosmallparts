package auth

import (
	"github.com/google/wire"
)

var (
	Providers = wire.NewSet(
		ProvidePasswordResetTokenDataManagerFromRepository,
	)
)

func ProvidePasswordResetTokenDataManagerFromRepository(r Repository) PasswordResetTokenDataManager {
	return r
}
