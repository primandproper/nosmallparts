//go:build wireinject
// +build wireinject

package api

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	identitymanager "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/manager"
	analyticscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
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
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/random"
	auditrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity"
	migrations "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/migrations"
	oauthrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/oauth"
	playscriptsrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/playscripts"
	stagesrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/stages"
	authservice "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/handlers/authentication"
	livekitsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/livekit"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/playscripts"
	stagessvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/stages"

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
		databasecfg.ProvideDatabase,
		migrations.NewMigrator,
		wire.Bind(new(database.Migrator), new(*migrations.Migrator)),
		loggingcfg.LogConfigProviders,
		authservice.AuthHTTPServiceProviders,
		metricscfg.MetricsConfigProviders,
		http.ProvidersHTTP,
		routingcfg.RoutingConfigProviders,
		random.RandProviders,
		identitymanager.IDManagerProviders,
		ProvideTextSearchConfig,
		ProvideQueuesConfig,
		ProvideUserTextSearcher,
		ProvideRegisterHandler,
		// repos
		auditrepo.AuditRepoProviders,
		identityrepo.IDRepoProviders,
		oauthrepo.OAuthRepoProviders,
		playscriptsrepo.PlayScriptsRepoProviders,
		stagesrepo.StagesRepoProviders,
		// services
		stagessvc.StagesWebSocketProviders,
		playscriptssvc.HTTPProviders,
		livekitsvc.Providers,
		ConfigProviders,
		ProvideCookiesConfig,
		cookies.NewCookieManager,
		ProvideOAuth2ManagerForSession,
		ProvideSessionFromCookieMiddleware,
		ProvideSessionCookieHandler,
		ProvideAPIRouter,
	)

	return nil, nil
}
