package grpc

import (
	"context"

	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/grpc/converters"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) CreateUser(ctx context.Context, request *identitysvc.CreateUserRequest) (*identitysvc.CreateUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	input := converters.ConvertGRPCUserRegistrationInputToUserRegistrationInput(request.Input)

	created, err := s.identityDataManager.CreateUser(ctx, input)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to create user")
	}

	x := &identitysvc.CreateUserResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Created:         converters.ConvertUserCreationResponseToGRPCUserCreationResponse(created),
	}

	return x, nil
}

func (s *serviceImpl) ArchiveUser(ctx context.Context, request *identitysvc.ArchiveUserRequest) (*identitysvc.ArchiveUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if err := s.identityDataManager.ArchiveUser(ctx, request.UserId); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Internal, "failed to archive user")
	}

	x := &identitysvc.ArchiveUserResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) GetUser(ctx context.Context, request *identitysvc.GetUserRequest) (*identitysvc.GetUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		keys.UserIDKey: request.UserId,
	}, span, s.logger)

	user, err := s.identityDataManager.GetUser(ctx, request.UserId)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch users from database")
	}

	x := &identitysvc.GetUserResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Result:          converters.ConvertUserToGRPCUser(user),
	}

	return x, nil
}

func (s *serviceImpl) GetUsers(ctx context.Context, request *identitysvc.GetUsersRequest) (*identitysvc.GetUsersResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	users, err := s.identityDataManager.GetUsers(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch users from database")
	}

	x := &identitysvc.GetUsersResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(users.Pagination, filter),
	}

	for _, user := range users.Data {
		x.Results = append(x.Results, converters.ConvertUserToGRPCUser(user))
	}

	return x, nil
}

func (s *serviceImpl) GetUsersForAccount(ctx context.Context, request *identitysvc.GetUsersForAccountRequest) (*identitysvc.GetUsersForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	users, err := s.identityDataManager.GetUsersForAccount(ctx, request.AccountId, filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to fetch users from database")
	}

	x := &identitysvc.GetUsersForAccountResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(users.Pagination, filter),
	}

	for _, user := range users.Data {
		x.Results = append(x.Results, converters.ConvertUserToGRPCUser(user))
	}

	return x, nil
}

func (s *serviceImpl) SearchForUsers(ctx context.Context, request *identitysvc.SearchForUsersRequest) (*identitysvc.SearchForUsersResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		keys.SearchQueryKey: request.Query,
	}, span, s.logger)

	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	users, err := s.identityDataManager.SearchForUsers(ctx, request.Query, request.UseSearchService, filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to search for users")
	}

	x := &identitysvc.SearchForUsersResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
		Pagination:      grpcconverters.ConvertPaginationToGRPCPagination(users.Pagination, filter),
	}

	for _, user := range users.Data {
		x.Results = append(x.Results, converters.ConvertUserToGRPCUser(user))
	}

	return x, nil
}

func (s *serviceImpl) UpdateUserDetails(ctx context.Context, request *identitysvc.UpdateUserDetailsRequest) (*identitysvc.UpdateUserDetailsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	input := converters.ConvertGRPCUserDetailsUpdateRequestInputToUserDetailsUpdateRequestInput(request.Input)

	if err = s.identityDataManager.UpdateUserDetails(ctx, sessionContextData.GetUserID(), input); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update user details")
	}

	x := &identitysvc.UpdateUserDetailsResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) UpdateUserEmailAddress(ctx context.Context, request *identitysvc.UpdateUserEmailAddressRequest) (*identitysvc.UpdateUserEmailAddressResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	if err = s.identityDataManager.UpdateUserEmailAddress(ctx, sessionContextData.GetUserID(), request.NewEmailAddress); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update user email address")
	}

	x := &identitysvc.UpdateUserEmailAddressResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) UpdateUserUsername(ctx context.Context, request *identitysvc.UpdateUserUsernameRequest) (*identitysvc.UpdateUserUsernameResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	if err = s.identityDataManager.UpdateUserUsername(ctx, sessionContextData.GetUserID(), request.NewUsername); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update user username")
	}

	x := &identitysvc.UpdateUserUsernameResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}

func (s *serviceImpl) UploadUserAvatar(ctx context.Context, request *identitysvc.UploadUserAvatarRequest) (*identitysvc.UploadUserAvatarResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	if err = s.identityDataManager.UploadUserAvatar(ctx, sessionContextData.GetUserID(), request.Base64EncodedData); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to upload user avatar")
	}

	x := &identitysvc.UploadUserAvatarResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}
