//go:build wireinject
// +build wireinject

package api

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	analyticscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	featureflagscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/featureflags/config"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	loggingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging/config"
	metricscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	tracingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing/config"
	routingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/http"
	auditrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity"
	repomigrations "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/migrations"
	oauthrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/oauth"
	authservice "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/handlers/authentication"

	"github.com/google/wire"
)

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.APIServiceConfig,
) (http.Server, error) {
	wire.Build(
		authentication.AuthProviders,
		encoding.Providers,
		msgconfig.MessageQueueProviders,
		analyticscfg.Providers,
		featureflagscfg.ProvidersFeatureFlags,
		tracing.ProvidersTracing,
		tracingcfg.TracingConfigProviders,
		observability.O11yProviders,
		postgres.PGProviders,
		repomigrations.MigrationProviders,
		loggingcfg.LogConfigProviders,
		authservice.AuthHTTPServiceProviders,
		metricscfg.MetricsConfigProviders,
		http.ProvidersHTTP,
		routingcfg.RoutingConfigProviders,
		// repos
		auditrepo.AuditRepoProviders,
		identityrepo.IDRepoProviders,
		oauthrepo.OAuthRepoProviders,
		ConfigProviders,
		ProvideAPIRouter,
	)

	return nil, nil
}
