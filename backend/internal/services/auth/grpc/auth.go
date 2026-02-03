package grpc

import (
	"context"
	"database/sql"
	"errors"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/grpc/converters"
	identityconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/grpc/converters"

	"google.golang.org/grpc/codes"
)

func (s *serviceImpl) GetAuthStatus(ctx context.Context, _ *authsvc.GetAuthStatusRequest) (*authsvc.GetAuthStatusResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	x := &authsvc.GetAuthStatusResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UserId:                   sessionContextData.GetUserID(),
		AccountStatus:            sessionContextData.Requester.AccountStatus,
		AccountStatusExplanation: sessionContextData.Requester.AccountStatusExplanation,
		ActiveAccount:            sessionContextData.GetActiveAccountID(),
	}

	return x, nil
}

func (s *serviceImpl) ExchangeToken(ctx context.Context, request *authsvc.ExchangeTokenRequest) (*authsvc.ExchangeTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}

	newToken, err := s.authenticationManager.ExchangeTokenForUser(ctx, request.RefreshToken)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to exchange token")
	}

	x := &authsvc.ExchangeTokenResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UserId:       sessionContextData.GetUserID(),
		AccountId:    sessionContextData.GetActiveAccountID(),
		RefreshToken: newToken.RefreshToken,
		AccessToken:  newToken.AccessToken,
		ExpiresUtc:   grpcconverters.ConvertTimeToPBTimestamp(newToken.ExpiresUTC),
	}

	return x, nil
}

func (s *serviceImpl) LoginForToken(ctx context.Context, request *authsvc.LoginForTokenRequest) (*authsvc.LoginForTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	return s.loginForToken(ctx, false, converters.ConvertGRPCUserLoginInputToUserLoginInput(request.Input))
}

func (s *serviceImpl) AdminLoginForToken(ctx context.Context, request *authsvc.AdminLoginForTokenRequest) (*authsvc.LoginForTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	return s.loginForToken(ctx, true, converters.ConvertGRPCUserLoginInputToUserLoginInput(request.Input))
}

func (s *serviceImpl) loginForToken(ctx context.Context, admin bool, input *auth.UserLoginInput) (*authsvc.LoginForTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	tokenResponse, err := s.authenticationManager.ProcessLogin(ctx, admin, input)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to process login request")
	}

	x := &authsvc.LoginForTokenResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertTokenResponseToGRPCTokenResponse(tokenResponse),
	}

	return x, nil
}

func (s *serviceImpl) CheckPermissions(ctx context.Context, request *authsvc.UserPermissionsRequestInput) (*authsvc.UserPermissionsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(keys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCCheckPermissionsRequestToUserPermissionsRequestInput(request)
	result, err := s.authManager.CheckUserPermissions(ctx, input)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "checking user permissions")
	}

	x := &authsvc.UserPermissionsResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Permissions: result.Permissions,
	}

	return x, nil
}

func (s *serviceImpl) GetActiveAccount(ctx context.Context, request *authsvc.GetActiveAccountRequest) (*authsvc.GetActiveAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(keys.UserIDKey, sessionContextData.GetUserID())

	account, err := s.identityRepository.GetAccount(ctx, sessionContextData.GetActiveAccountID())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "failed to get active account")
		}
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get active account")
	}

	x := &authsvc.GetActiveAccountResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: identityconverters.ConvertAccountToGRPCAccount(account),
	}

	return x, nil
}

func (s *serviceImpl) GetSelf(ctx context.Context, request *authsvc.GetSelfRequest) (*authsvc.GetSelfResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(keys.UserIDKey, sessionContextData.GetUserID())

	user, err := s.identityRepository.GetUser(ctx, sessionContextData.GetUserID())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "failed to get user")
		}
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to get user")
	}

	x := &authsvc.GetSelfResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: identityconverters.ConvertUserToGRPCUser(user),
	}

	return x, nil
}

func (s *serviceImpl) RedeemPasswordResetToken(ctx context.Context, request *authsvc.RedeemPasswordResetTokenRequest) (*authsvc.RedeemPasswordResetTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(keys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCRedeemPasswordResetTokenRequestToPasswordResetTokenRedemptionRequestInput(request)
	if err = s.authManager.PasswordResetTokenRedemption(ctx, input); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "refreshing totp secret")
	}

	x := &authsvc.RedeemPasswordResetTokenResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) RefreshTOTPSecret(ctx context.Context, request *authsvc.RefreshTOTPSecretRequest) (*authsvc.RefreshTOTPSecretResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(keys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCRefreshTOTPSecretRequestToTOTPSecretRefreshInput(request)
	totpSecretResponse, err := s.authManager.NewTOTPSecret(ctx, input)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "refreshing totp secret")
	}

	x := &authsvc.RefreshTOTPSecretResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: converters.ConvertTOTPSecretRefreshResponseToGRPCTOTPSecretRefreshResponse(totpSecretResponse),
	}

	return x, nil
}

func (s *serviceImpl) RequestEmailVerificationEmail(ctx context.Context, request *authsvc.RequestEmailVerificationEmailRequest) (*authsvc.RequestEmailVerificationEmailResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(keys.UserIDKey, sessionContextData.GetUserID())

	if err = s.authManager.RequestEmailVerificationEmail(ctx); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to request email verification email")
	}

	x := &authsvc.RequestEmailVerificationEmailResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) RequestPasswordResetToken(ctx context.Context, request *authsvc.RequestPasswordResetTokenRequest) (*authsvc.RequestPasswordResetTokenResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(keys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCRequestPasswordResetTokenRequestToPasswordResetTokenCreationRequestInput(request)
	if err = s.authManager.CreatePasswordResetToken(ctx, input); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to create password reset token")
	}

	x := &authsvc.RequestPasswordResetTokenResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) RequestUsernameReminder(ctx context.Context, request *authsvc.RequestUsernameReminderRequest) (*authsvc.RequestUsernameReminderResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(keys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCRequestUsernameReminderRequestToUsernameReminderRequestInput(request)
	if err = s.authManager.RequestUsernameReminder(ctx, input); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to issue username reminder")
	}

	x := &authsvc.RequestUsernameReminderResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}

func (s *serviceImpl) VerifyEmailAddress(ctx context.Context, request *authsvc.VerifyEmailAddressRequest) (*authsvc.VerifyEmailAddressResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(keys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCVerifyEmailAddressRequestToEmailAddressVerificationRequestInput(request)
	if err = s.authManager.VerifyUserEmailAddress(ctx, input); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to verify user's email address")
	}

	x := &authsvc.VerifyEmailAddressResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Verified: true,
	}

	return x, nil
}

func (s *serviceImpl) VerifyTOTPSecret(ctx context.Context, request *authsvc.VerifyTOTPSecretRequest) (*authsvc.VerifyTOTPSecretResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(keys.UserIDKey, request.UserId)

	input := converters.ConvertGRPCVerifyTOTPSecretRequestToTOTPSecretVerificationInput(request)
	if err := s.authManager.TOTPSecretVerification(ctx, input); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to verify TOTP secret")
	}

	x := &authsvc.VerifyTOTPSecretResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Accepted: true,
	}

	return x, nil
}

func (s *serviceImpl) UpdatePassword(ctx context.Context, request *authsvc.UpdatePasswordRequest) (*authsvc.UpdatePasswordResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span)

	sessionContextData, err := s.fetchSessionContext(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Unauthenticated, "failed to get session context data")
	}
	logger = logger.WithValue(keys.UserIDKey, sessionContextData.GetUserID())

	input := converters.ConvertGRPCUpdatePasswordRequestToPasswordUpdateInput(request)
	if err = s.authManager.UpdatePassword(ctx, input); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "failed to update password")
	}

	x := &authsvc.UpdatePasswordResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
	}

	return x, nil
}
