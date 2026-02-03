package uploadedmedia

import (
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	generated "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/uploadedmedia/generated"
)

const (
	o11yName = "uploaded_media_db_client"
)

// repository is the uploaded media repository client.
type repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	db                *sql.DB
}

// ProvideUploadedMediaRepository provides a new repository.
func ProvideUploadedMediaRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
) types.Repository {
	c := &repository{
		Client:            client,
		db:                client.DB(),
		tracer:            tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		generatedQuerier:  generated.New(),
		auditLogEntryRepo: auditLogEntryRepo,
		logger:            logging.EnsureLogger(logger).WithName(o11yName),
	}

	return c
}
