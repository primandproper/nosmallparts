package authentication

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/platform/v4/analytics"
	"github.com/verygoodsoftwarenotvirus/platform/v4/encoding"
	"github.com/verygoodsoftwarenotvirus/platform/v4/featureflags"
	"github.com/verygoodsoftwarenotvirus/platform/v4/internalerrors"
	"github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue"
	msgconfig "github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/routing"

	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

const (
	serviceName                = "auth_service"
	AuthProviderParamKey       = "auth_provider"
	rejectedRequestCounterName = "auth_service.rejected_requests"
)

// TODO: remove this when Goth can handle concurrency.
var useProvidersMutex = sync.Mutex{}

type (
	// service handles passwords service-wide.
	service struct {
		config                   *Config
		logger                   logging.Logger
		authenticator            authentication.Authenticator
		analyticsReporter        analytics.EventReporter
		featureFlagManager       featureflags.FeatureFlagManager
		userDataManager          identity.UserDataManager
		accountMembershipManager identity.AccountUserMembershipDataManager
		encoderDecoder           encoding.ServerEncoderDecoder
		authProviderFetcher      func(*http.Request) string
		tracer                   tracing.Tracer
		dataChangesPublisher     messagequeue.Publisher
		oauth2Server             *server.Server
		tokenIssuer              tokens.Issuer
	}
)

// ProvideService builds a new AuthDataService.
func ProvideService(
	ctx context.Context,
	logger logging.Logger,
	cfg *Config,
	authenticator authentication.Authenticator,
	oauthRepo oauth.Repository,
	identityRepo identity.Repository,
	encoder encoding.ServerEncoderDecoder,
	tracerProvider tracing.TracerProvider,
	publisherProvider messagequeue.PublisherProvider,
	featureFlagManager featureflags.FeatureFlagManager,
	analyticsReporter analytics.EventReporter,
	routeParamManager routing.RouteParamManager,
	queuesConfig *msgconfig.QueuesConfig,
) (auth.AuthDataService, error) {
	if queuesConfig == nil {
		return nil, internalerrors.NilConfigError("queuesConfig for AuthDataService")
	}

	dataChangesPublisher, publisherProviderErr := publisherProvider.ProvidePublisher(ctx, queuesConfig.DataChangesTopicName)
	if publisherProviderErr != nil {
		return nil, fmt.Errorf("setting up %s data changes publisher: %w", serviceName, publisherProviderErr)
	}

	signer, err := cfg.Tokens.ProvideTokenIssuer(logger, tracerProvider)
	if err != nil {
		return nil, fmt.Errorf("creating json web token signer: %w", err)
	}

	manager := ProvideOAuth2ClientManager(logger, tracerProvider, &cfg.OAuth2, oauthRepo)

	svc := &service{
		logger:                   logging.NewNamedLogger(logger, serviceName),
		encoderDecoder:           encoder,
		config:                   cfg,
		userDataManager:          identityRepo,
		accountMembershipManager: identityRepo,
		authenticator:            authenticator,
		tracer:                   tracing.NewNamedTracer(tracerProvider, serviceName),
		dataChangesPublisher:     dataChangesPublisher,
		featureFlagManager:       featureFlagManager,
		analyticsReporter:        analyticsReporter,
		tokenIssuer:              signer,
		authProviderFetcher:      routeParamManager.BuildRouteParamStringIDFetcher(AuthProviderParamKey),
		oauth2Server:             ProvideOAuth2ServerImplementation(logger, tracerProvider, identityRepo, authenticator, signer, manager),
	}

	useProvidersMutex.Lock()
	goth.UseProviders(
		google.New(
			svc.config.SSO.Google.ClientID,
			svc.config.SSO.Google.ClientSecret,
			svc.config.SSO.Google.CallbackURL,
		),
	)
	useProvidersMutex.Unlock()

	return svc, nil
}
