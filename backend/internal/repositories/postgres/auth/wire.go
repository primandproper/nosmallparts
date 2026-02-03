package auth

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"

	"github.com/google/wire"
)

var (
	AuthRepoProviders = wire.NewSet(
		ProvideAuthRepository,
		ProvidePasswordResetTokenDataManager,
	)
)

func ProvidePasswordResetTokenDataManager(r auth.Repository) auth.PasswordResetTokenDataManager {
	return r
}
