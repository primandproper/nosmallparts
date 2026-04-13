package waitlists

import (
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
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
		tracer:           tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier: generated.New(),
		logger:           logging.NewNamedLogger(logger, o11yName),
	}

	return c
}
