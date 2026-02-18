package repositories

import (
	"github.com/google/wire"
)

var (
	// RepositoryProviders are what we provide to dependency injection.
	RepositoryProviders = wire.NewSet(
		ProvideMigrator,
	)
)
