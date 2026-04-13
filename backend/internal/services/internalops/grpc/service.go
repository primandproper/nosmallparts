package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	settingssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/internalops"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/types"
	msgconfig "github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"google.golang.org/grpc/codes"
)

const (
	o11yName = "internalops_service"
)

var _ settingssvc.InternalOperationsServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		settingssvc.UnimplementedInternalOperationsServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		msgConfig                 *msgconfig.Config
		sessionContextDataFetcher func(context.Context) (sessions.ContextData, error)
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	msgConfig *msgconfig.Config,
) settingssvc.InternalOperationsServer {
	return &serviceImpl{
		msgConfig: msgConfig,
		logger:    logging.NewNamedLogger(logger, o11yName),
		tracer:    tracing.NewNamedTracer(tracerProvider, o11yName),
	}
}

func (s serviceImpl) PublishArbitraryQueueMessage(ctx context.Context, request *settingssvc.PublishArbitraryQueueMessageRequest) (*settingssvc.PublishArbitraryQueueMessageResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "fetching session context data")
	}

	if !sessionContextData.ServiceRolePermissionChecker().HasPermission(authorization.PublishArbitraryQueueMessagePermission) {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "user is not permitted")
	}

	pp, err := msgconfig.ProvidePublisherProvider(ctx, s.logger, tracingnoop.NewTracerProvider(), s.msgConfig)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "establishing publisher provider")
	}

	publisher, err := pp.ProvidePublisher(ctx, request.QueueName)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "initializing publisher")
	}

	if err = publisher.Publish(ctx, []byte(request.Body)); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "publishing message")
	}

	x := &settingssvc.PublishArbitraryQueueMessageResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Success: true,
	}

	return x, nil
}
