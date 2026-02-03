package oauth

import "github.com/google/wire"

var (
	OAuthRepoProviders = wire.NewSet(
		ProvideOAuthRepository,
	)
)
