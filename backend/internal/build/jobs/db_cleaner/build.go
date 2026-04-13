//go:build wireinject
// +build wireinject

package dbcleaner

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	databasecfg "github.com/verygoodsoftwarenotvirus/platform/v4/database/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/postgres"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	loggingcfg "github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging/config"
	metricscfg "github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics/config"
	tracingcfg "github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/maintenance"
	dbcleaner "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/oauth/workers/db_cleaner"

	"github.com/google/wire"
)

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.DBCleanerConfig,
) (*dbcleaner.Job, error) {
	wire.Build(
		dbcleaner.ProvidersDBCleaner,
		tracingcfg.TracingConfigProviders,
		observability.O11yProviders,
		databasecfg.ClientConfigProviders,
		postgres.PGProviders,
		loggingcfg.LogConfigProviders,
		metricscfg.MetricsConfigProviders,
		maintenance.Providers,
		ConfigProviders,
	)

	return nil, nil
}
