package grpc

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/fakes"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	"github.com/verygoodsoftwarenotvirus/platform/v4/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildFakeSessionContextData() *sessions.ContextData {
	return &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:                   identityfakes.BuildFakeID(),
			AccountStatus:            identity.GoodStandingUserAccountStatus.String(),
			AccountStatusExplanation: "",
			ServicePermissions:       authorization.NewServiceRolePermissionChecker("service_user"),
		},
		ActiveAccountID: identityfakes.BuildFakeID(),
		AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
			identityfakes.BuildFakeID(): authorization.NewAccountRolePermissionChecker("account_member"),
		},
	}
}

func buildContextWithSessionData(t *testing.T) context.Context {
	t.Helper()
	sessionData := buildFakeSessionContextData()
	sessionData.AccountPermissions[sessionData.ActiveAccountID] = authorization.NewAccountRolePermissionChecker("account_member")
	return context.WithValue(t.Context(), sessions.SessionContextDataKey, sessionData)
}

func TestServiceImpl_GetAuthStatus(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		request := &authsvc.GetAuthStatusRequest{}

		response, err := service.GetAuthStatus(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotEmpty(t, response.UserId)
		assert.NotEmpty(t, response.ActiveAccount)
		assert.Equal(t, identity.GoodStandingUserAccountStatus.String(), response.AccountStatus)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.GetAuthStatusRequest{}

		response, err := service.GetAuthStatus(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})
}

func TestServiceImpl_ExchangeToken(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		fakeTokenResponse := &auth.TokenResponse{
			UserID:       identityfakes.BuildFakeID(),
			AccountID:    identityfakes.BuildFakeID(),
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresUTC:   time.Now().Add(time.Hour),
		}

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ExchangeTokenForUser), mock.Anything, "refresh-token").Return(fakeTokenResponse, nil)

		request := &authsvc.ExchangeTokenRequest{
			RefreshToken: "refresh-token",
		}

		response, err := service.ExchangeToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.Equal(t, fakeTokenResponse.AccessToken, response.AccessToken)
		assert.Equal(t, fakeTokenResponse.RefreshToken, response.RefreshToken)
		assert.NotEmpty(t, response.UserId)
		assert.NotEmpty(t, response.AccountId)

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.ExchangeTokenRequest{
			RefreshToken: "refresh-token",
		}

		response, err := service.ExchangeToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error exchanging token", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ExchangeTokenForUser), mock.Anything, "refresh-token").Return((*auth.TokenResponse)(nil), errors.New("exchange failed"))

		request := &authsvc.ExchangeTokenRequest{
			RefreshToken: "refresh-token",
		}

		response, err := service.ExchangeToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})
}

func TestServiceImpl_LoginForToken(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager := buildTestService(t)
		ctx := t.Context()

		fakeTokenResponse := &auth.TokenResponse{
			UserID:       identityfakes.BuildFakeID(),
			AccountID:    identityfakes.BuildFakeID(),
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresUTC:   time.Now().Add(time.Hour),
		}

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ProcessLogin), mock.Anything, false, mock.AnythingOfType("*auth.UserLoginInput")).Return(fakeTokenResponse, nil)

		request := &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  "testuser",
				Password:  "password123",
				TotpToken: "123456",
			},
		}

		response, err := service.LoginForToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeTokenResponse.AccessToken, response.Result.AccessToken)
		assert.Equal(t, fakeTokenResponse.RefreshToken, response.Result.RefreshToken)

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})

	t.Run("error processing login", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager := buildTestService(t)
		ctx := t.Context()

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ProcessLogin), mock.Anything, false, mock.AnythingOfType("*auth.UserLoginInput")).Return((*auth.TokenResponse)(nil), errors.New("login failed"))

		request := &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  "testuser",
				Password:  "wrongpassword",
				TotpToken: "123456",
			},
		}

		response, err := service.LoginForToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})
}

func TestServiceImpl_AdminLoginForToken(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager := buildTestService(t)
		ctx := t.Context()

		fakeTokenResponse := &auth.TokenResponse{
			UserID:       identityfakes.BuildFakeID(),
			AccountID:    identityfakes.BuildFakeID(),
			AccessToken:  "admin-access-token",
			RefreshToken: "admin-refresh-token",
			ExpiresUTC:   time.Now().Add(time.Hour),
		}

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ProcessLogin), mock.Anything, true, mock.AnythingOfType("*auth.UserLoginInput")).Return(fakeTokenResponse, nil)

		request := &authsvc.AdminLoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  "adminuser",
				Password:  "adminpassword123",
				TotpToken: "123456",
			},
		}

		response, err := service.AdminLoginForToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeTokenResponse.AccessToken, response.Result.AccessToken)
		assert.Equal(t, fakeTokenResponse.RefreshToken, response.Result.RefreshToken)

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})

	t.Run("error processing admin login", func(t *testing.T) {
		t.Parallel()

		service, _, _, authenticationManager := buildTestService(t)
		ctx := t.Context()

		authenticationManager.On(reflection.GetMethodName(authenticationManager.ProcessLogin), mock.Anything, true, mock.AnythingOfType("*auth.UserLoginInput")).Return((*auth.TokenResponse)(nil), errors.New("admin login failed"))

		request := &authsvc.AdminLoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  "adminuser",
				Password:  "wrongpassword",
				TotpToken: "123456",
			},
		}

		response, err := service.AdminLoginForToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authenticationManager)
	})
}

func TestServiceImpl_CheckPermissions(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		fakePermissionsResponse := &auth.UserPermissionsResponse{
			Permissions: map[string]bool{
				"read_users":   true,
				"create_users": false,
			},
		}

		authManager.On(reflection.GetMethodName(authManager.CheckUserPermissions), mock.Anything, mock.AnythingOfType("*auth.UserPermissionsRequestInput")).Return(fakePermissionsResponse, nil)

		request := &authsvc.UserPermissionsRequestInput{
			Permissions: []string{"read_users", "create_users"},
		}

		response, err := service.CheckPermissions(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.Equal(t, fakePermissionsResponse.Permissions, response.Permissions)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.UserPermissionsRequestInput{
			Permissions: []string{"read_users"},
		}

		response, err := service.CheckPermissions(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error checking permissions", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.CheckUserPermissions), mock.Anything, mock.AnythingOfType("*auth.UserPermissionsRequestInput")).Return((*auth.UserPermissionsResponse)(nil), errors.New("permission check failed"))

		request := &authsvc.UserPermissionsRequestInput{
			Permissions: []string{"read_users"},
		}

		response, err := service.CheckPermissions(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_GetActiveAccount(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		fakeAccount := identityfakes.BuildFakeAccount()

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, sessionData.GetActiveAccountID()).Return(fakeAccount, nil)

		request := &authsvc.GetActiveAccountRequest{}

		response, err := service.GetActiveAccount(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeAccount.ID, response.Result.Id)

		mock.AssertExpectationsForObjects(t, identityRepo)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.GetActiveAccountRequest{}

		response, err := service.GetActiveAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("account not found", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, sessionData.GetActiveAccountID()).Return((*identity.Account)(nil), sql.ErrNoRows)

		request := &authsvc.GetActiveAccountRequest{}

		response, err := service.GetActiveAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, identityRepo)
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, sessionData.GetActiveAccountID()).Return((*identity.Account)(nil), errors.New("database error"))

		request := &authsvc.GetActiveAccountRequest{}

		response, err := service.GetActiveAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, identityRepo)
	})
}

func TestServiceImpl_GetSelf(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		fakeUser := identityfakes.BuildFakeUser()

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetUser), mock.Anything, sessionData.GetUserID()).Return(fakeUser, nil)

		request := &authsvc.GetSelfRequest{}

		response, err := service.GetSelf(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeUser.ID, response.Result.Id)

		mock.AssertExpectationsForObjects(t, identityRepo)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.GetSelfRequest{}

		response, err := service.GetSelf(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetUser), mock.Anything, sessionData.GetUserID()).Return((*identity.User)(nil), sql.ErrNoRows)

		request := &authsvc.GetSelfRequest{}

		response, err := service.GetSelf(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, identityRepo)
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		service, identityRepo, _, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		sessionData := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
		identityRepo.On(reflection.GetMethodName(identityRepo.GetUser), mock.Anything, sessionData.GetUserID()).Return((*identity.User)(nil), errors.New("database error"))

		request := &authsvc.GetSelfRequest{}

		response, err := service.GetSelf(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, identityRepo)
	})
}

func TestServiceImpl_RedeemPasswordResetToken(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.PasswordResetTokenRedemption), mock.Anything, mock.AnythingOfType("*auth.PasswordResetTokenRedemptionRequestInput")).Return(nil)

		request := &authsvc.RedeemPasswordResetTokenRequest{
			Token:       "reset-token",
			NewPassword: "newpassword123",
		}

		response, err := service.RedeemPasswordResetToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.RedeemPasswordResetTokenRequest{
			Token:       "reset-token",
			NewPassword: "newpassword123",
		}

		response, err := service.RedeemPasswordResetToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error redeeming token", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.PasswordResetTokenRedemption), mock.Anything, mock.AnythingOfType("*auth.PasswordResetTokenRedemptionRequestInput")).Return(errors.New("redemption failed"))

		request := &authsvc.RedeemPasswordResetTokenRequest{
			Token:       "invalid-token",
			NewPassword: "newpassword123",
		}

		response, err := service.RedeemPasswordResetToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}
func TestServiceImpl_RefreshTOTPSecret(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		fakeTOTPResponse := &auth.TOTPSecretRefreshResponse{
			TwoFactorSecret: "new-secret",
			TwoFactorQRCode: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
		}

		authManager.On(reflection.GetMethodName(authManager.NewTOTPSecret), mock.Anything, mock.AnythingOfType("*auth.TOTPSecretRefreshInput")).Return(fakeTOTPResponse, nil)

		request := &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: "password123",
			TotpToken:       "123456",
		}

		response, err := service.RefreshTOTPSecret(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeTOTPResponse.TwoFactorSecret, response.Result.TwoFactorSecret)
		assert.Equal(t, fakeTOTPResponse.TwoFactorQRCode, response.Result.TwoFactorQrCode)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: "password123",
			TotpToken:       "123456",
		}

		response, err := service.RefreshTOTPSecret(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error refreshing TOTP secret", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.NewTOTPSecret), mock.Anything, mock.AnythingOfType("*auth.TOTPSecretRefreshInput")).Return((*auth.TOTPSecretRefreshResponse)(nil), errors.New("refresh failed"))

		request := &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: "wrongpassword",
			TotpToken:       "123456",
		}

		response, err := service.RefreshTOTPSecret(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_RequestEmailVerificationEmail(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.RequestEmailVerificationEmail), mock.Anything).Return(nil)

		request := &authsvc.RequestEmailVerificationEmailRequest{}

		response, err := service.RequestEmailVerificationEmail(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.RequestEmailVerificationEmailRequest{}

		response, err := service.RequestEmailVerificationEmail(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error requesting email verification", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.RequestEmailVerificationEmail), mock.Anything).Return(errors.New("email request failed"))

		request := &authsvc.RequestEmailVerificationEmailRequest{}

		response, err := service.RequestEmailVerificationEmail(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_RequestPasswordResetToken(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.CreatePasswordResetToken), mock.Anything, mock.AnythingOfType("*auth.PasswordResetTokenCreationRequestInput")).Return(nil)

		request := &authsvc.RequestPasswordResetTokenRequest{
			EmailAddress: "test@example.com",
		}

		response, err := service.RequestPasswordResetToken(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.RequestPasswordResetTokenRequest{
			EmailAddress: "test@example.com",
		}

		response, err := service.RequestPasswordResetToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error creating password reset token", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.CreatePasswordResetToken), mock.Anything, mock.AnythingOfType("*auth.PasswordResetTokenCreationRequestInput")).Return(errors.New("token creation failed"))

		request := &authsvc.RequestPasswordResetTokenRequest{
			EmailAddress: "invalid@example.com",
		}

		response, err := service.RequestPasswordResetToken(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_RequestUsernameReminder(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.RequestUsernameReminder), mock.Anything, mock.AnythingOfType("*auth.UsernameReminderRequestInput")).Return(nil)

		request := &authsvc.RequestUsernameReminderRequest{
			EmailAddress: "test@example.com",
		}

		response, err := service.RequestUsernameReminder(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.RequestUsernameReminderRequest{
			EmailAddress: "test@example.com",
		}

		response, err := service.RequestUsernameReminder(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error requesting username reminder", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.RequestUsernameReminder), mock.Anything, mock.AnythingOfType("*auth.UsernameReminderRequestInput")).Return(errors.New("reminder request failed"))

		request := &authsvc.RequestUsernameReminderRequest{
			EmailAddress: "invalid@example.com",
		}

		response, err := service.RequestUsernameReminder(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_VerifyEmailAddress(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.VerifyUserEmailAddress), mock.Anything, mock.AnythingOfType("*auth.EmailAddressVerificationRequestInput")).Return(nil)

		request := &authsvc.VerifyEmailAddressRequest{
			Token: "verification-token",
		}

		response, err := service.VerifyEmailAddress(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.True(t, response.Verified)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.VerifyEmailAddressRequest{
			Token: "verification-token",
		}

		response, err := service.VerifyEmailAddress(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error verifying email address", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.VerifyUserEmailAddress), mock.Anything, mock.AnythingOfType("*auth.EmailAddressVerificationRequestInput")).Return(errors.New("verification failed"))

		request := &authsvc.VerifyEmailAddressRequest{
			Token: "invalid-token",
		}

		response, err := service.VerifyEmailAddress(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_VerifyTOTPSecret(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := t.Context()

		authManager.On(reflection.GetMethodName(authManager.TOTPSecretVerification), mock.Anything, mock.AnythingOfType("*auth.TOTPSecretVerificationInput")).Return(nil)

		request := &authsvc.VerifyTOTPSecretRequest{
			UserId:    identityfakes.BuildFakeID(),
			TotpToken: "123456",
		}

		response, err := service.VerifyTOTPSecret(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.True(t, response.Accepted)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error verifying TOTP secret", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := t.Context()

		authManager.On(reflection.GetMethodName(authManager.TOTPSecretVerification), mock.Anything, mock.AnythingOfType("*auth.TOTPSecretVerificationInput")).Return(errors.New("verification failed"))

		request := &authsvc.VerifyTOTPSecretRequest{
			UserId:    identityfakes.BuildFakeID(),
			TotpToken: "invalid",
		}

		response, err := service.VerifyTOTPSecret(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}

func TestServiceImpl_UpdatePassword(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.UpdatePassword), mock.Anything, mock.AnythingOfType("*auth.PasswordUpdateInput")).Return(nil)

		request := &authsvc.UpdatePasswordRequest{
			NewPassword:     "newpassword123",
			CurrentPassword: "oldpassword123",
			TotpToken:       "123456",
		}

		response, err := service.UpdatePassword(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mock.AssertExpectationsForObjects(t, authManager)
	})

	t.Run("error fetching session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context() // No session context data

		request := &authsvc.UpdatePasswordRequest{
			NewPassword:     "newpassword123",
			CurrentPassword: "oldpassword123",
			TotpToken:       "123456",
		}

		response, err := service.UpdatePassword(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("error updating password", func(t *testing.T) {
		t.Parallel()

		service, _, authManager, _ := buildTestService(t)
		ctx := buildContextWithSessionData(t)

		authManager.On(reflection.GetMethodName(authManager.UpdatePassword), mock.Anything, mock.AnythingOfType("*auth.PasswordUpdateInput")).Return(errors.New("password update failed"))

		request := &authsvc.UpdatePasswordRequest{
			NewPassword:     "newpassword123",
			CurrentPassword: "wrongpassword",
			TotpToken:       "123456",
		}

		response, err := service.UpdatePassword(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())

		mock.AssertExpectationsForObjects(t, authManager)
	})
}
