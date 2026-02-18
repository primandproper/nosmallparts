package oauth

import (
	"github.com/google/wire"
)

var (
	Providers = wire.NewSet(
		ProvideOAuth2ClientDataManagerFromRepository,
		ProvideOAuth2ClientTokenDataManagerFromRepository,
	)
)

func ProvideOAuth2ClientDataManagerFromRepository(r Repository) OAuth2ClientDataManager {
	return r
}

func ProvideOAuth2ClientTokenDataManagerFromRepository(r Repository) OAuth2ClientTokenDataManager {
	return r
}
