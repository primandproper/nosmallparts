package authentication

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication"
	mockauthn "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthenticationService_validateLogin(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		authenticator := &mockauthn.Authenticator{}
		authenticator.On(
			"CredentialsAreValid",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = authenticator

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.True(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, authenticator)
	})

	T.Run("with invalid two factor code", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		authenticator := &mockauthn.Authenticator{}
		authenticator.On(
			"CredentialsAreValid",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(true, authentication.ErrInvalidTOTPToken)
		helper.service.authenticator = authenticator

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, authenticator)
	})

	T.Run("with error returned from validator", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		expectedErr := errors.New("arbitrary")

		authenticator := &mockauthn.Authenticator{}
		authenticator.On(
			"CredentialsAreValid",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(false, expectedErr)
		helper.service.authenticator = authenticator

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, authenticator)
	})

	T.Run("with invalid login", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		authenticator := &mockauthn.Authenticator{}
		authenticator.On(
			"CredentialsAreValid",
			testutils.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleLoginInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleLoginInput.TOTPToken,
		).Return(false, nil)
		helper.service.authenticator = authenticator

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.False(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, authenticator)
	})
}
