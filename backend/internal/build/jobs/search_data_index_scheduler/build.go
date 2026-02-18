//go:build wireinject
// +build wireinject

package searchdataindexscheduler

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	loggingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging/config"
	metricscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics/config"
	tracingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/text/indexing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity"

	"github.com/google/wire"
)

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.SearchDataIndexSchedulerConfig,
) (*indexing.IndexScheduler, error) {
	wire.Build(
		indexing.ProvidersIndexing,
		tracingcfg.TracingConfigProviders,
		observability.O11yProviders,
		msgconfig.MessageQueueProviders,
		databasecfg.ClientConfigProviders,
		postgres.PGProviders,
		loggingcfg.LogConfigProviders,
		metricscfg.MetricsConfigProviders,
		auditlogentries.AuditRepoProviders,
		identity.IDRepoProviders,
		ProvideIndexFunctions,
		ConfigProviders,
	)

	return nil, nil
}
