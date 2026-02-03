//go:build wireinject
// +build wireinject

package datachangemessagehandler

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/functions/datachangemessagehandler"
	analyticscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres"
	emailcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/email/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	loggingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging/config"
	metricscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	tracingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/uploads/objectstorage"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/webhooks"
	identityindexing "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/indexing"

	"github.com/google/wire"
)

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.AsyncMessageHandlerConfig,
) (*datachangemessagehandler.AsyncDataChangeMessageHandler, error) {
	wire.Build(
		datachangemessagehandler.Providers,
		msgconfig.MessageQueueProviders,
		postgres.PGProviders,
		auditlogentries.AuditRepoProviders,
		identity.IDRepoProviders,
		webhooks.WebhookProviders,
		analyticscfg.Providers,
		emailcfg.Providers,
		metricscfg.MetricsConfigProviders,
		encoding.Providers,
		loggingcfg.LogConfigProviders,
		tracingcfg.TracingConfigProviders,
		observability.O11yProviders,
		tracing.ProvidersTracing,
		objectstorage.Providers,
		identityindexing.Providers,
		ConfigProviders,
		SearcherProviders,
	)

	return nil, nil
}
