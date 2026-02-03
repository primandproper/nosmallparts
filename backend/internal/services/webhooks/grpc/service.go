package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

const (
	o11yName = "configuration_service"
)

var _ webhookssvc.WebhooksServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		webhookssvc.UnimplementedWebhooksServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
		webhookRepository         webhooks.Repository
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	webhookRepository webhooks.Repository,
) webhookssvc.WebhooksServiceServer {
	return &serviceImpl{
		logger:                    logging.EnsureLogger(logger).WithName(o11yName),
		tracer:                    tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		sessionContextDataFetcher: sessions.FetchContextDataFromContext,
		webhookRepository:         webhookRepository,
	}
}
