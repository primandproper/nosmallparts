package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth/managers"

	"github.com/stretchr/testify/mock"
)

var (
	_ managers.AuthManagerInterface = (*AuthManager)(nil)
)

// AuthManager is a mock implementation of the auth manager.
type AuthManager struct {
	mock.Mock
}

// CheckUserPermissions is a mock method.
func (m *AuthManager) CheckUserPermissions(ctx context.Context, input *auth.UserPermissionsRequestInput) (*auth.UserPermissionsResponse, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*auth.UserPermissionsResponse), args.Error(1)
}

// TOTPSecretVerification is a mock method.
func (m *AuthManager) TOTPSecretVerification(ctx context.Context, input *auth.TOTPSecretVerificationInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// NewTOTPSecret is a mock method.
func (m *AuthManager) NewTOTPSecret(ctx context.Context, input *auth.TOTPSecretRefreshInput) (*auth.TOTPSecretRefreshResponse, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*auth.TOTPSecretRefreshResponse), args.Error(1)
}

// UpdatePassword is a mock method.
func (m *AuthManager) UpdatePassword(ctx context.Context, input *auth.PasswordUpdateInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// RequestUsernameReminder is a mock method.
func (m *AuthManager) RequestUsernameReminder(ctx context.Context, input *auth.UsernameReminderRequestInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// CreatePasswordResetToken is a mock method.
func (m *AuthManager) CreatePasswordResetToken(ctx context.Context, input *auth.PasswordResetTokenCreationRequestInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// PasswordResetTokenRedemption is a mock method.
func (m *AuthManager) PasswordResetTokenRedemption(ctx context.Context, input *auth.PasswordResetTokenRedemptionRequestInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// RequestEmailVerificationEmail is a mock method.
func (m *AuthManager) RequestEmailVerificationEmail(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// VerifyUserEmailAddress is a mock method.
func (m *AuthManager) VerifyUserEmailAddress(ctx context.Context, input *auth.EmailAddressVerificationRequestInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
