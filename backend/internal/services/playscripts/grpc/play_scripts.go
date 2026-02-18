package grpc

import (
	"context"
	"errors"

	playscripts "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/play_scripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/playscripts/grpc/converters"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreatePlayScript(ctx context.Context, request *playscriptssvc.CreatePlayScriptRequest) (*playscriptssvc.CreatePlayScriptResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID).WithValue(keys.UserIDKey, sessionContextData.Requester.UserID)

	playScriptInput, castMemberInputs, lineInputs := converters.ConvertGRPCPlayScriptCreationRequestInputToInputs(request.Input, sessionContextData.Requester.UserID, sessionContextData.GetActiveAccountID())
	if err = playScriptInput.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate play script creation request")
	}

	created, err := s.playScriptRepository.CreatePlayScript(ctx, playScriptInput, castMemberInputs, lineInputs)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create play script")
	}

	x := &playscriptssvc.CreatePlayScriptResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Created: converters.ConvertPlayScriptToGRPCPlayScript(created),
	}

	return x, nil
}

func (s *serviceImpl) GetPlayScript(ctx context.Context, request *playscriptssvc.GetPlayScriptRequest) (*playscriptssvc.GetPlayScriptResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(keys.PlayScriptIDKey, request.PlayScriptId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	playScript, err := s.playScriptRepository.GetPlayScript(ctx, request.PlayScriptId, request.IncludeLinkedScripts)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch play script")
	}

	// Allow access if:
	// 1. The play script is public, OR
	// 2. The play script belongs to the user's account, OR
	// 3. The user is a service admin (e.g. admin app users)
	if !playScript.IsPublic && playScript.BelongsToAccount != sessionContextData.GetActiveAccountID() {
		if !sessionContextData.ServiceRolePermissionChecker().IsServiceAdmin() {
			return nil, observability.PrepareAndLogGRPCStatus(errors.New("permission denied"), logger, span, codes.PermissionDenied, "play script does not belong to account")
		}
	}

	x := &playscriptssvc.GetPlayScriptResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Result: converters.ConvertPlayScriptToGRPCPlayScript(playScript),
	}

	return x, nil
}

func (s *serviceImpl) GetPlayScripts(ctx context.Context, request *playscriptssvc.GetPlayScriptsRequest) (*playscriptssvc.GetPlayScriptsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	playScripts, err := s.playScriptRepository.GetPlayScripts(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch play scripts")
	}

	x := &playscriptssvc.GetPlayScriptsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(playScripts.Pagination, filter),
	}

	s.enrichPlayScriptsWithRatings(ctx, playScripts.Data, sessionContextData.Requester.UserID)

	for _, playScript := range playScripts.Data {
		x.Results = append(x.Results, converters.ConvertPlayScriptToGRPCPlayScript(playScript))
	}

	return x, nil
}

func (s *serviceImpl) GetPlayScriptsForAccount(ctx context.Context, request *playscriptssvc.GetPlayScriptsForAccountRequest) (*playscriptssvc.GetPlayScriptsForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	playScripts, err := s.playScriptRepository.GetPlayScriptsForAccount(ctx, sessionContextData.ActiveAccountID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch play scripts for account")
	}

	x := &playscriptssvc.GetPlayScriptsForAccountResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(playScripts.Pagination, filter),
	}

	s.enrichPlayScriptsWithRatings(ctx, playScripts.Data, sessionContextData.Requester.UserID)

	for _, playScript := range playScripts.Data {
		x.Results = append(x.Results, converters.ConvertPlayScriptToGRPCPlayScript(playScript))
	}

	return x, nil
}

func (s *serviceImpl) GetPublicPlayScripts(ctx context.Context, request *playscriptssvc.GetPublicPlayScriptsRequest) (*playscriptssvc.GetPublicPlayScriptsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	playScripts, err := s.playScriptRepository.GetPublicPlayScripts(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch public play scripts")
	}

	x := &playscriptssvc.GetPublicPlayScriptsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(playScripts.Pagination, filter),
	}

	s.enrichPlayScriptsWithRatings(ctx, playScripts.Data, sessionContextData.Requester.UserID)

	for _, playScript := range playScripts.Data {
		x.Results = append(x.Results, converters.ConvertPlayScriptToGRPCPlayScript(playScript))
	}

	return x, nil
}

func (s *serviceImpl) SearchPlayScriptsForXParticipants(ctx context.Context, request *playscriptssvc.SearchPlayScriptsForXParticipantsRequest) (*playscriptssvc.SearchPlayScriptsForXParticipantsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	if request.ParticipantCount < 1 {
		return nil, observability.PrepareAndLogGRPCStatus(nil, logger, span, codes.InvalidArgument, "participant_count must be at least 1")
	}

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	playScripts, err := s.playScriptRepository.SearchPlayScriptsForXParticipants(ctx, int(request.ParticipantCount), sessionContextData.ActiveAccountID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to search play scripts for participant count")
	}

	x := &playscriptssvc.SearchPlayScriptsForXParticipantsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(playScripts.Pagination, filter),
	}

	s.enrichPlayScriptsWithRatings(ctx, playScripts.Data, sessionContextData.Requester.UserID)

	for _, playScript := range playScripts.Data {
		x.Results = append(x.Results, converters.ConvertPlayScriptToGRPCPlayScript(playScript))
	}

	return x, nil
}

func (s *serviceImpl) UpdatePlayScript(ctx context.Context, request *playscriptssvc.UpdatePlayScriptRequest) (*playscriptssvc.UpdatePlayScriptResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(keys.PlayScriptIDKey, request.PlayScriptId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	// Fetch the existing play script
	playScript, err := s.playScriptRepository.GetPlayScript(ctx, request.PlayScriptId, false)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch play script")
	}

	// Verify the play script belongs to the user's account or user is service admin
	if playScript.BelongsToAccount != sessionContextData.GetActiveAccountID() {
		if !sessionContextData.ServiceRolePermissionChecker().IsServiceAdmin() {
			return nil, observability.PrepareAndLogGRPCStatus(errors.New("permission denied"), logger, span, codes.PermissionDenied, "play script does not belong to account")
		}
	}

	// Apply updates
	updateInput := converters.ConvertGRPCPlayScriptUpdateRequestInputToPlayScriptUpdateRequestInput(request.Input)
	if err = updateInput.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.InvalidArgument, "failed to validate play script update request")
	}

	playScript.Update(updateInput)

	if err = s.playScriptRepository.UpdatePlayScript(ctx, playScript); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update play script")
	}

	x := &playscriptssvc.UpdatePlayScriptResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Updated: converters.ConvertPlayScriptToGRPCPlayScript(playScript),
	}

	return x, nil
}

// enrichPlayScriptsWithRatings attaches rating stats and user ratings to each play script.
func (s *serviceImpl) enrichPlayScriptsWithRatings(ctx context.Context, playScripts []*playscripts.PlayScript, userID string) {
	if len(playScripts) == 0 {
		return
	}
	ids := make([]string, len(playScripts))
	for i, ps := range playScripts {
		ids[i] = ps.ID
	}
	stats, err := s.playScriptRepository.GetRatingStatsForPlayScripts(ctx, ids)
	if err != nil {
		return
	}
	userRatings, err := s.playScriptRepository.GetUserRatingsForPlayScripts(ctx, userID, ids)
	if err != nil {
		return
	}
	for _, ps := range playScripts {
		if s, ok := stats[ps.ID]; ok {
			ps.RatingStats = s
		}
		if r, ok := userRatings[ps.ID]; ok {
			ps.UserRating = r
		}
	}
}

func (s *serviceImpl) RatePlayScript(ctx context.Context, request *playscriptssvc.RatePlayScriptRequest) (*playscriptssvc.RatePlayScriptResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(keys.PlayScriptIDKey, request.PlayScriptId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	userID := sessionContextData.Requester.UserID
	logger = logger.WithValue(keys.UserIDKey, userID)

	rating := int(request.Rating)
	if rating < 1 || rating > 5 {
		return nil, observability.PrepareAndLogGRPCStatus(errors.New("rating must be between 1 and 5"), logger, span, codes.InvalidArgument, "invalid rating")
	}

	created, err := s.playScriptRepository.RatePlayScript(ctx, userID, request.PlayScriptId, rating)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to rate play script")
	}

	x := &playscriptssvc.RatePlayScriptResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
		Created: converters.ConvertPlayScriptRatingToGRPC(created),
	}

	return x, nil
}

func (s *serviceImpl) ArchivePlayScript(ctx context.Context, request *playscriptssvc.ArchivePlayScriptRequest) (*playscriptssvc.ArchivePlayScriptResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(keys.PlayScriptIDKey, request.PlayScriptId)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}
	logger = logger.WithValue(keys.AccountIDKey, sessionContextData.ActiveAccountID)

	// Fetch the existing play script to verify ownership
	playScript, err := s.playScriptRepository.GetPlayScript(ctx, request.PlayScriptId, false)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch play script")
	}

	// Verify the play script belongs to the user's account or user is service admin
	if playScript.BelongsToAccount != sessionContextData.GetActiveAccountID() {
		if !sessionContextData.ServiceRolePermissionChecker().IsServiceAdmin() {
			return nil, observability.PrepareAndLogGRPCStatus(errors.New("permission denied"), logger, span, codes.PermissionDenied, "play script does not belong to account")
		}
	}

	if err = s.playScriptRepository.ArchivePlayScript(ctx, request.PlayScriptId); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to archive play script")
	}

	x := &playscriptssvc.ArchivePlayScriptResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId:          span.SpanContext().TraceID().String(),
			CurrentAccountId: sessionContextData.ActiveAccountID,
		},
	}

	return x, nil
}
