package grpc

import (
	"context"

	identityconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/converters"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) AdminUpdateUserStatus(ctx context.Context, request *identitysvc.AdminUpdateUserStatusRequest) (*identitysvc.AdminUpdateUserStatusResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to fetch session context data")
	}

	if !sessionContextData.Requester.ServicePermissions.CanUpdateUserAccountStatuses() {
		return nil, observability.PrepareAndLogGRPCStatus(nil, logger, span, codes.Unauthenticated, "user account status update requester does not have permission")
	}

	if err = s.identityDataManager.AdminUpdateUserStatus(ctx, identityconverters.ConvertGRPCAdminUpdateUserStatusRequestToUserAccountStatusUpdateInput(request)); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating user account status")
	}

	x := &identitysvc.AdminUpdateUserStatusResponse{
		ResponseDetails: s.buildResponseDetails(ctx, span),
	}

	return x, nil
}
