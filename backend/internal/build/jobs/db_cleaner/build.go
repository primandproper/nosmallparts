//go:build wireinject
// +build wireinject

package dbcleaner

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	loggingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging/config"
	metricscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics/config"
	tracingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/maintenance"
	repomigrations "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/migrations"
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
		postgres.PGProviders,
		repomigrations.MigrationProviders,
		loggingcfg.LogConfigProviders,
		metricscfg.MetricsConfigProviders,
		maintenance.Providers,
		ConfigProviders,
	)

	return nil, nil
}
