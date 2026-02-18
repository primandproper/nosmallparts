package dbcleaner

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"

	"github.com/google/wire"
)

var (
	// ConfigProviders represents this package's offering to the dependency injector.
	ConfigProviders = wire.NewSet(
		wire.FieldsOf(
			new(*config.DBCleanerConfig),
			"Observability",
			"Database",
		),
	)
)
