package searchdataindexscheduler

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"

	"github.com/google/wire"
)

var (
	// ConfigProviders represents this package's offering to the dependency injector.
	ConfigProviders = wire.NewSet(
		wire.FieldsOf(
			new(*config.SearchDataIndexSchedulerConfig),
			"Queues",
			"Events",
			"Observability",
			"Database",
		),
	)
)
