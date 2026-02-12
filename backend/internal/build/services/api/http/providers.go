package api

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	identitymanager "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/manager"
	oauthdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing"
	textsearchcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/text/config"
	identityindexing "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/indexing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/handlers/authentication"
)

// ProvideCookiesConfig returns the cookies config from the API service config.
func ProvideCookiesConfig(cfg *config.APIServiceConfig) *cookies.Config {
	return &cfg.Cookies
}

// ProvideOAuth2ManagerForSession returns an OAuth2 manager that uses the same token storage as the auth service, for validating Bearer tokens when issuing session cookies.
func ProvideOAuth2ManagerForSession(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	cfg *config.APIServiceConfig,
	oauthRepo oauthdomain.Repository,
) *oauth2ManagerForSession {
	return &oauth2ManagerForSession{authentication.ProvideOAuth2ClientManager(logger, tracerProvider, &cfg.Services.Auth.OAuth2, oauthRepo)}
}

// ProvideSessionFromCookieMiddleware returns middleware that decodes the session from header/cookie and attaches ContextData to the request context.
func ProvideSessionFromCookieMiddleware(
	cookieManager cookies.Manager,
	cookiesCfg *cookies.Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) routing.Middleware {
	cfg := sessions.SessionFromCookieConfig{
		HeaderName: sessions.DefaultSessionHeaderName,
		CookieName: cookiesCfg.CookieName,
	}
	return sessions.SessionFromCookieMiddleware(cookieManager, cookiesCfg.CookieName, cfg, logger, tracerProvider)
}

// ProvideSessionCookieHandler builds the handler that issues the session cookie given a Bearer token.
func ProvideSessionCookieHandler(
	oauth2Manager *oauth2ManagerForSession,
	identityRepo identity.Repository,
	cookieManager cookies.Manager,
	cookiesCfg *cookies.Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) *SessionCookieHandler {
	return NewSessionCookieHandler(oauth2Manager, identityRepo, cookieManager, cookiesCfg.CookieName, logger, tracerProvider)
}

// ProvideTextSearchConfig returns the text search config from the API service config.
func ProvideTextSearchConfig(cfg *config.APIServiceConfig) *textsearchcfg.Config {
	return &cfg.TextSearch
}

// ProvideQueuesConfig returns the queues config from the API service config.
func ProvideQueuesConfig(cfg *config.APIServiceConfig) *msgconfig.QueuesConfig {
	return &cfg.Queues
}

// ProvideUserTextSearcher provides the identity user search index for the identity manager.
func ProvideUserTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (identityindexing.UserTextSearcher, error) {
	return textsearchcfg.ProvideIndex[identityindexing.UserSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		identityindexing.IndexTypeUsers,
	)
}

// ProvideRegisterHandler builds the HTTP handler for POST /api/register.
func ProvideRegisterHandler(
	identityManager identitymanager.IdentityDataManager,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) *RegisterHandler {
	return NewRegisterHandler(identityManager, logger, tracerProvider)
}
