package dataprivacy

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

const (
	o11yName = "webhook_db_client"
)

// repository is the webhook repository client.
type repository struct {
	database.Client
	tracer       tracing.Tracer
	logger       logging.Logger
	identityRepo identity.Repository
}

// ProvideDataPrivacyRepository provides a new repository.
func ProvideDataPrivacyRepository(logger logging.Logger, tracerProvider tracing.TracerProvider, identityRepo identity.Repository, client database.Client) dataprivacy.Repository {
	c := &repository{
		Client:       client,
		tracer:       tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		identityRepo: identityRepo,
		logger:       logging.EnsureLogger(logger).WithName(o11yName),
	}

	return c
}
