package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/waitlists"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/waitlists"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

const (
	o11yName = "waitlists_service"
)

var _ waitlistssvc.WaitlistsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		waitlistssvc.UnimplementedWaitlistsServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
		waitlistRepository        waitlists.Repository
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	waitlistRepository waitlists.Repository,
) waitlistssvc.WaitlistsServiceServer {
	return &serviceImpl{
		logger:                    logging.EnsureLogger(logger).WithName(o11yName),
		tracer:                    tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		sessionContextDataFetcher: sessions.FetchContextDataFromContext,
		waitlistRepository:        waitlistRepository,
	}
}
