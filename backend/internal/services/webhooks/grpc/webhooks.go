package grpc

import (
	"context"

	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/webhooks/grpc/converters"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreateWebhook(ctx context.Context, request *webhookssvc.CreateWebhookRequest) (*webhookssvc.CreateWebhookResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	input := converters.ConvertGRPCWebhookCreationRequestInputToWebhookDatabaseCreationInput(request.Input, sessionContextData.ActiveAccountID)
	if err = input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate webhook creation request")
	}

	created, err := s.webhookRepository.CreateWebhook(ctx, input)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create webhook")
	}

	x := &webhookssvc.CreateWebhookResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertWebhookToGRPCWebhook(created),
	}

	return x, nil
}

func (s *serviceImpl) AddWebhookTriggerEvent(ctx context.Context, request *webhookssvc.AddWebhookTriggerEventRequest) (*webhookssvc.AddWebhookTriggerEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	input := converters.ConvertGRPCWebhookTriggerEventDatabaseCreationInputToWebhookTriggerEventDatabaseCreationInput(request.Input)
	created, err := s.webhookRepository.AddWebhookTriggerEvent(ctx, sessionContextData.ActiveAccountID, input)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to add webhook trigger event")
	}

	x := &webhookssvc.AddWebhookTriggerEventResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertWebhookTriggerEventToGRPCWebhookTriggerEvent(created),
	}

	return x, nil
}

func (s *serviceImpl) GetWebhook(ctx context.Context, request *webhookssvc.GetWebhookRequest) (*webhookssvc.GetWebhookResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(keys.WebhookIDKey, request.WebhookId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	webhook, err := s.webhookRepository.GetWebhook(ctx, request.WebhookId, sessionContextData.GetActiveAccountID())
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch webhook")
	}

	x := &webhookssvc.GetWebhookResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertWebhookToGRPCWebhook(webhook),
	}

	return x, nil
}

func (s *serviceImpl) GetWebhooks(ctx context.Context, request *webhookssvc.GetWebhooksRequest) (*webhookssvc.GetWebhooksResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)
	retrieved, err := s.webhookRepository.GetWebhooks(ctx, sessionContextData.ActiveAccountID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch webhooks")
	}

	x := &webhookssvc.GetWebhooksResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(retrieved.Pagination, filter),
	}

	for _, webhook := range retrieved.Data {
		x.Results = append(x.Results, converters.ConvertWebhookToGRPCWebhook(webhook))
	}

	return x, nil
}

func (s *serviceImpl) ArchiveWebhook(ctx context.Context, request *webhookssvc.ArchiveWebhookRequest) (*webhookssvc.ArchiveWebhookResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	if err = s.webhookRepository.ArchiveWebhook(ctx, request.WebhookId, sessionContextData.ActiveAccountID); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive webhook")
	}

	x := &webhookssvc.ArchiveWebhookResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) ArchiveWebhookTriggerEvent(ctx context.Context, request *webhookssvc.ArchiveWebhookTriggerEventRequest) (*webhookssvc.ArchiveWebhookTriggerEventResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.WebhookIDKey, request.WebhookId).WithValue(keys.WebhookTriggerEventIDKey, request.WebhookTriggerEventId)

	if err := s.webhookRepository.ArchiveWebhookTriggerEvent(ctx, request.WebhookId, request.WebhookTriggerEventId); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive webhook trigger event")
	}

	x := &webhookssvc.ArchiveWebhookTriggerEventResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}
