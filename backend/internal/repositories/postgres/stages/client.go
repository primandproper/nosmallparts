package stages

import (
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/stages/generated"
)

const (
	o11yName = "stages_db_client"
)

// repository is the stages repository client.
type repository struct {
	database.Client
	tracer           tracing.Tracer
	logger           logging.Logger
	generatedQuerier generated.Querier
	db               *sql.DB
}

// ProvideStagesRepository provides a new repository.
func ProvideStagesRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	client database.Client,
) stages.Repository {
	c := &repository{
		Client:           client,
		db:               client.DB(),
		tracer:           tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		generatedQuerier: generated.New(),
		logger:           logging.EnsureLogger(logger).WithName(o11yName),
	}

	return c
}
