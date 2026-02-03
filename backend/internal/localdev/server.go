package localdev

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication"
	apiserver "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/build/services/api"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	identityconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres"
	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres/testing"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/redis"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/random"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auth"
	identityrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/migrations"
	notificationsrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/notifications"
	oauthrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/oauth"
	settingsrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/settings"
	webhooksrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/pkg/client"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CreatePremadeAdminUser(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	identityRepo identity.Repository,
	dbClient database.Client,
	premadeAdminUser *identity.User,
) (*identity.User, error) {
	hasher := authentication.ProvideArgon2Authenticator(logger, tracerProvider)

	actuallyHashedPass, err := hasher.HashPassword(ctx, premadeAdminUser.HashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	premadeAdminUser.HashedPassword = actuallyHashedPass

	var user *identity.User
	if user, err = identityRepo.GetUserByUsername(ctx, premadeAdminUser.Username); err == nil {
		return user, nil
	}

	user, err = identityRepo.CreateUser(ctx, identityconverters.ConvertUserToUserDatabaseCreationInput(premadeAdminUser))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// one-off query because I really don't want to make this functionality concrete
	if _, err = dbClient.DB().Exec(fmt.Sprintf("UPDATE users SET service_role='service_admin' WHERE id='%s'", user.ID)); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if err = identityRepo.MarkUserTwoFactorSecretAsVerified(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("failed to mark user as verified: %w", err)
	}

	adminUser, err := identityRepo.GetAdminUserByUsername(ctx, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}

	return adminUser, nil
}

func CreateOAuth2ClientForService(ctx context.Context, pgc database.Client, dbCfg *databasecfg.Config, oauth2Input *oauth.OAuth2ClientDatabaseCreationInput) (*oauth.OAuth2Client, error) {
	auditRepo := auditlogentries.ProvideAuditLogRepository(nil, nil, pgc)
	oauth2ClientManager := oauthrepo.ProvideOAuthRepository(nil, nil, auditRepo, dbCfg, pgc)

	createdClient, err := oauth2ClientManager.CreateOAuth2Client(ctx, oauth2Input)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth2 client: %w", err)
	}

	return createdClient, nil
}

func BuildInProcessServer(ctx context.Context, cfg *config.APIServiceConfig) (server *apiserver.Server, databaseClient database.Client, dbCfg *databasecfg.Config, err error) {
	logger, _, _, err := cfg.Observability.ProvideThreePillars(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("setting up logger: %w", err)
	}

	redisConfig, _, err := redis.BuildContainerBackedRedisConfig(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("connecting to redis: %w", err)
	}
	cfg.Events.Publisher.Provider = msgconfig.ProviderRedis
	cfg.Events.Publisher.Redis = *redisConfig
	cfg.Events.Consumer.Redis = *redisConfig

	// set up a database container, migrate it, and build a connection client
	_, db, dbCfg, err := pgtesting.BuildDatabaseContainer(ctx, "integration_testing")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	cfg.Database = *dbCfg

	if err = migrations.NewMigrator(logger, tracing.NewNoopTracerProvider(), db, dbCfg).Migrate(ctx); err != nil {
		return nil, nil, nil, fmt.Errorf(": %w", err)
	}

	databaseClient, err = postgres.ProvideDatabaseClient(ctx, logger, tracing.NewNoopTracerProvider(), dbCfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("initializing database client: %w", err)
	}

	// create premade admin user
	server, err = apiserver.NewServer(ctx, logger, cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("building API server: %w", err)
	}

	return server, databaseClient, dbCfg, nil
}

// DatabaseInitFunc is a function that performs database initialization operations.
// It receives the database client, config, logger, and tracer to perform arbitrary operations.
type DatabaseInitFunc func(ctx context.Context, dbClient database.Client, dbCfg *databasecfg.Config, logger logging.Logger, tracerProvider tracing.TracerProvider) error

// WithIdentityRepository provides an identity repository for custom operations.
// The provided function receives a fully configured identity.Repository along with logger, tracer, and database client.
func WithIdentityRepository(fn func(ctx context.Context, repo identity.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider, dbClient database.Client) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *databasecfg.Config, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
		auditLogRepo := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, dbClient)
		identityRepo := identityrepo.ProvideIdentityRepository(logger, tracerProvider, auditLogRepo, dbClient)
		return fn(ctx, identityRepo, logger, tracerProvider, dbClient)
	}
}

// WithOAuth2Repository provides an OAuth2 repository for custom operations.
// The provided function receives a fully configured oauth.Repository along with logger and tracer.
func WithOAuth2Repository(fn func(ctx context.Context, repo oauth.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *databasecfg.Config, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
		auditLogRepo := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, dbClient)
		oauthRepo := oauthrepo.ProvideOAuthRepository(logger, tracerProvider, auditLogRepo, dbCfg, dbClient)
		return fn(ctx, oauthRepo, logger, tracerProvider)
	}
}

// WithAuthRepository provides an auth repository for custom operations.
// The provided function receives a fully configured auth.Repository along with logger and tracer.
func WithAuthRepository(fn func(ctx context.Context, repo auth.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *databasecfg.Config, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
		auditLogRepo := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, dbClient)
		authRepo := authrepo.ProvideAuthRepository(logger, tracerProvider, auditLogRepo, dbClient)
		return fn(ctx, authRepo, logger, tracerProvider)
	}
}

// WithSettingsRepository provides a settings repository for custom operations.
// The provided function receives a fully configured settings.Repository along with logger and tracer.
func WithSettingsRepository(fn func(ctx context.Context, repo settings.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *databasecfg.Config, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
		auditLogRepo := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, dbClient)
		settingsRepo := settingsrepo.ProvideSettingsRepository(logger, tracerProvider, auditLogRepo, dbClient)
		return fn(ctx, settingsRepo, logger, tracerProvider)
	}
}

// WithWebhooksRepository provides a webhooks repository for custom operations.
// The provided function receives a fully configured webhooks.Repository along with logger and tracer.
func WithWebhooksRepository(fn func(ctx context.Context, repo webhooks.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *databasecfg.Config, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
		auditLogRepo := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, dbClient)
		webhooksRepo := webhooksrepo.ProvideWebhooksRepository(logger, tracerProvider, auditLogRepo, dbClient)
		return fn(ctx, webhooksRepo, logger, tracerProvider)
	}
}

// WithNotificationsRepository provides a notifications repository for custom operations.
// The provided function receives a fully configured notifications.Repository along with logger and tracer.
func WithNotificationsRepository(fn func(ctx context.Context, repo notifications.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) error) DatabaseInitFunc {
	return func(ctx context.Context, dbClient database.Client, dbCfg *databasecfg.Config, logger logging.Logger, tracerProvider tracing.TracerProvider) error {
		auditLogRepo := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, dbClient)
		notificationsRepo := notificationsrepo.ProvideNotificationsRepository(logger, tracerProvider, auditLogRepo, dbClient)
		return fn(ctx, notificationsRepo, logger, tracerProvider)
	}
}

// AllInOne sets up a complete local development environment with a docker-backed server,
// database, and runs the provided database initialization functions.
func AllInOne(ctx context.Context, cfg *config.APIServiceConfig, initFuncs ...DatabaseInitFunc) (*apiserver.Server, error) {
	server, databaseClient, dbCfg, err := BuildInProcessServer(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("building in-process server: %w", err)
	}

	log.Printf("%sDATABASE CONNECTION URL: %s%s", strings.Repeat("\n", 10), dbCfg.ConnectionDetails.URI(), strings.Repeat("\n", 10))

	logger, tracerProvider, _, err := cfg.Observability.ProvideThreePillars(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting o11y pillars: %w", err)
	}

	// Run all database initialization functions
	for i, initFunc := range initFuncs {
		if err = initFunc(ctx, databaseClient, dbCfg, logger, tracerProvider); err != nil {
			return nil, fmt.Errorf("running database init function %d: %w", i, err)
		}
	}

	return server, nil
}

func BuildInsecureOAuthedGRPCClient(
	ctx context.Context,
	createdClientID,
	createdClientSecret,
	httpTestServerAddress,
	grpcServerAddress,
	token string,
) (client.Client, error) {
	state, err := random.GenerateBase64EncodedString(ctx, 32)
	if err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     createdClientID,
		ClientSecret: createdClientSecret,
		Scopes:       []string{"anything"}, // TODO: This should be nil-able
		RedirectURL:  httpTestServerAddress,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   httpTestServerAddress + "/oauth2/authorize",
			TokenURL:  httpTestServerAddress + "/oauth2/token",
		},
	}

	authCodeURL := oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge_method", "plain"),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		authCodeURL,
		http.NoBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth2 code: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Location", "localhost")

	httpClient := tracing.BuildTracedHTTPClient()
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth2 code: %w", err)
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			log.Println("failed to close oauth2 response body", err)
		}
	}()

	const (
		codeKey = "code"
	)

	rl, err := res.Location()
	if err != nil {
		return nil, fmt.Errorf("getting location value from response: %w", err)
	}

	code := rl.Query().Get(codeKey)
	if code == "" {
		return nil, fmt.Errorf("code not returned from oauth2 redirect")
	}

	oauth2Token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging OAuth2 code: %w", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(&insecureOAuth{
			TokenSource: oauth2Config.TokenSource(ctx, oauth2Token),
		}),
	}

	c, err := client.BuildClient(grpcServerAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("building client: %w", err)
	}

	return c, nil
}

// Custom insecure OAuth2 credentials that skip security checks.
type insecureOAuth struct {
	TokenSource oauth2.TokenSource
}

func (i *insecureOAuth) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	token, err := i.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	return map[string]string{"authorization": token.Type() + " " + token.AccessToken}, nil
}

func (i *insecureOAuth) RequireTransportSecurity() bool {
	return false // Explicitly allow insecure transport
}

func FetchLoginTokenForUser(ctx context.Context, grpcServerAddr string, loginInput *authsvc.UserLoginInput) (string, error) {
	unauthedClient, err := client.BuildUnauthenticatedGRPCClient(grpcServerAddr)
	if err != nil {
		return "", fmt.Errorf("initializing client: %w", err)
	}

	tokenRes, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
		Input: loginInput,
	})
	if err != nil {
		return "", fmt.Errorf("fetching login token: %w", err)
	}

	return tokenRes.Result.AccessToken, nil
}
