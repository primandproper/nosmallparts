package manager

import (
	"github.com/google/wire"
)

var (
	OAuthManagerProviders = wire.NewSet(
		NewOAuth2Manager,
	)
)
