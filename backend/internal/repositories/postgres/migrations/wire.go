package migrations

import (
	"github.com/google/wire"
)

var (
	// MigrationProviders are what we offer to dependency injection.
	MigrationProviders = wire.NewSet(
		ProvideMigratorFactory,
	)
)
