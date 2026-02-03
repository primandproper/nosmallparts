package oauth

import (
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/cryptography/encryption"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/cryptography/encryption/salsa20"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/oauth/generated"
)

const (
	o11yName = "oauth_db_client"
)

// repository is the oauth2 client and token repo implemenation.
type repository struct {
	database.Client
	tracer                  tracing.Tracer
	logger                  logging.Logger
	generatedQuerier        generated.Querier
	auditLogEntryRepo       audit.Repository
	oauth2ClientTokenEncDec encryption.EncryptorDecryptor
	db                      *sql.DB
}

// ProvideOAuthRepository provides a new repository.
func ProvideOAuthRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
	cfg *databasecfg.Config,
	client database.Client,
) oauth.Repository {
	encDec, err := salsa20.NewEncryptorDecryptor(tracerProvider, logger, []byte(cfg.OAuth2TokenEncryptionKey))
	if err != nil {
		return nil
	}

	c := &repository{
		Client:                  client,
		db:                      client.DB(),
		tracer:                  tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		generatedQuerier:        generated.New(),
		auditLogEntryRepo:       auditLogEntryRepo,
		oauth2ClientTokenEncDec: encDec,
		logger:                  logging.EnsureLogger(logger).WithName(o11yName),
	}

	return c
}
