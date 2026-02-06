package waitlists

import (
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/waitlists/generated"
)

const (
	o11yName = "waitlists_db_client"
)

// repository is the waitlists repository implementation.
type repository struct {
	database.Client
	tracer           tracing.Tracer
	logger           logging.Logger
	generatedQuerier generated.Querier
	readDB           *sql.DB
	writeDB          *sql.DB
}

// ProvideWaitlistsRepository provides a new repository.
func ProvideWaitlistsRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	client database.Client,
) waitlists.Repository {
	c := &repository{
		Client:           client,
		readDB:           client.ReadDB(),
		writeDB:          client.WriteDB(),
		tracer:           tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		generatedQuerier: generated.New(),
		logger:           logging.EnsureLogger(logger).WithName(o11yName),
	}

	return c
}
