package maintenance

import (
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/maintenance"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/maintenance/generated"
)

const (
	o11yName = "maintenance_db_client"
)

// repository is the maintenance repository implementation.
type repository struct {
	database.Client
	tracer           tracing.Tracer
	logger           logging.Logger
	generatedQuerier generated.Querier
	readDB           *sql.DB
	writeDB          *sql.DB
}

// ProvideMaintenanceRepository provides a new repository.
func ProvideMaintenanceRepository(logger logging.Logger, tracerProvider tracing.TracerProvider, client database.Client) maintenance.MaintenanceDataManager {
	c := &repository{
		Client:           client,
		readDB:           client.ReadDB(),
		writeDB:          client.WriteDB(),
		tracer:           tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier: generated.New(),
		logger:           logging.NewNamedLogger(logger, o11yName),
	}

	// these are here for future use
	_, _ = c.readDB, c.writeDB

	return c
}
