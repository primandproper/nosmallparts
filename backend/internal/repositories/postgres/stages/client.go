package stages

import (
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
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
		db:               client.WriteDB(),
		tracer:           tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier: generated.New(),
		logger:           logging.NewNamedLogger(logger, o11yName),
	}

	return c
}
