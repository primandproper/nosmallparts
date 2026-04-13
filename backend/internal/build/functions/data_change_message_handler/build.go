//go:build wireinject
// +build wireinject

package datachangemessagehandler

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/functions/datachangemessagehandler"
	analyticscfg "github.com/verygoodsoftwarenotvirus/platform/v4/analytics/config"
	databasecfg "github.com/verygoodsoftwarenotvirus/platform/v4/database/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/postgres"
	emailcfg "github.com/verygoodsoftwarenotvirus/platform/v4/email/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/encoding"
	msgconfig "github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	loggingcfg "github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging/config"
	metricscfg "github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	tracingcfg "github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/uploads/objectstorage"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity"
	issue_reports "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/issuereports"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/notifications"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/settings"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/waitlists"
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
		databasecfg.ClientConfigProviders,
		postgres.PGProviders,
		auditlogentries.AuditRepoProviders,
		dataprivacy.DataPrivProviders,
		identity.IDRepoProviders,
		issue_reports.IssueReportsRepoProviders,
		notifications.NotifRepoProviders,
		settings.SettingsRepoProviders,
		uploadedmedia.UploadedMediaRepoProviders,
		waitlists.WaitlistsRepoProviders,
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
