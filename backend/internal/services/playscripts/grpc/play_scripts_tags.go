package grpc

import (
	"context"

	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/play_scripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/playscripts/grpc/converters"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) GetPlayScriptsWithTag(ctx context.Context, request *playscriptssvc.GetPlayScriptsWithTagRequest) (*playscriptssvc.GetPlayScriptsWithTagResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	playScripts, err := s.playScriptRepository.GetPlayScriptsWithTag(ctx, request.TagName, filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch play scripts with tag")
	}

	x := &playscriptssvc.GetPlayScriptsWithTagResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(playScripts.Pagination, filter),
	}

	for _, ps := range playScripts.Data {
		x.Results = append(x.Results, converters.ConvertPlayScriptToGRPCPlayScript(ps))
	}

	return x, nil
}

func (s *serviceImpl) CreateScriptTag(ctx context.Context, request *playscriptssvc.CreateScriptTagRequest) (*playscriptssvc.CreateScriptTagResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	if _, err := s.sessionContextDataFetcher(ctx); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}

	created, err := s.playScriptRepository.CreateScriptTag(ctx, identifiers.New(), request.Name, request.Description)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create script tag")
	}

	return &playscriptssvc.CreateScriptTagResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Created: converters.ConvertScriptTagToGRPC(created),
	}, nil
}

func (s *serviceImpl) AddTagToPlayScript(ctx context.Context, request *playscriptssvc.AddTagToPlayScriptRequest) (*playscriptssvc.AddTagToPlayScriptResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(keys.PlayScriptIDKey, request.PlayScriptId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}

	if err = s.playScriptRepository.AddTagToPlayScript(ctx, request.PlayScriptId, request.ScriptTagId); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to add tag to play script")
	}

	return &playscriptssvc.AddTagToPlayScriptResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
	}, nil
}

func (s *serviceImpl) RemoveTagFromPlayScript(ctx context.Context, request *playscriptssvc.RemoveTagFromPlayScriptRequest) (*playscriptssvc.RemoveTagFromPlayScriptResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(keys.PlayScriptIDKey, request.PlayScriptId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}

	if err = s.playScriptRepository.RemoveTagFromPlayScript(ctx, request.PlayScriptId, request.ScriptTagId); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to remove tag from play script")
	}

	return &playscriptssvc.RemoveTagFromPlayScriptResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
	}, nil
}
