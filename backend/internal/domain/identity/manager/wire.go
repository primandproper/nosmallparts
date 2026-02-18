package manager

import "github.com/google/wire"

var (
	IDManagerProviders = wire.NewSet(
		NewIdentityDataManager,
	)
)
