package managers

import "github.com/google/wire"

var (
	AuthManagerProviders = wire.NewSet(
		ProvideAuthManager,
	)
)
