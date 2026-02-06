package databasecfg

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"

	"github.com/google/wire"
)

// ProvideClientConfig converts Config to database.ClientConfig for wire.
// Wire extracts fields by value from config structs, so we accept a value and return a pointer.
//
//nolint:gocritic // hugeParam: intentionally accepts value because wire extracts fields by value
func ProvideClientConfig(cfg Config) database.ClientConfig {
	return &cfg
}

var (
	// ClientConfigProviders provides the conversion from Config to database.ClientConfig.
	// Include this in wire builds that need to use postgres.PGProviders.
	ClientConfigProviders = wire.NewSet(
		ProvideClientConfig,
	)

	// DatabaseConfigProviders are what we provide to dependency injection.
	DatabaseConfigProviders = wire.NewSet(
		ProvideDatabase,
		ProvideClientConfig,
	)
)
