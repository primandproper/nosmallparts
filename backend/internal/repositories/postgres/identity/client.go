package identity

import (
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/random"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity/generated"
)

const (
	o11yName = "identity_db_client"
)

var _ identity.Repository = (*repository)(nil)

// repository is the identity repository implementation.
type repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	secretGenerator   random.Generator
	db                *sql.DB
}

// ProvideIdentityRepository provides a new repository.
func ProvideIdentityRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
) identity.Repository {
	c := &repository{
		Client:            client,
		db:                client.DB(),
		tracer:            tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		generatedQuerier:  generated.New(),
		auditLogEntryRepo: auditLogEntryRepo,
		secretGenerator:   random.NewGenerator(logger, tracerProvider),
		logger:            logging.EnsureLogger(logger).WithName(o11yName),
	}

	return c
}
