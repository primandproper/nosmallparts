package integration

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/fakes"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_LoginForToken(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		user := createServiceUserForTest(t, true, fakes.BuildFakeUserRegistrationInput())
		actual := fetchLoginTokenForUserForTest(t, user)

		assert.NotEmpty(t, actual)
	})

	T.Run("2FA is not required for non-admin users who haven't verified their secrets", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user := createServiceUserForTest(t, false, fakes.BuildFakeUserRegistrationInput())

		loginInput := &authsvc.UserLoginInput{
			Username: user.Username,
			Password: user.HashedPassword,
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: loginInput,
		})
		assert.NoError(t, err)
		assert.NotNil(t, tokenRes)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)
	})

	T.Run("with bogus input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		loginInput := &authsvc.UserLoginInput{
			Username:  " ",
			Password:  "1",
			TotpToken: "otp scode",
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: loginInput,
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})
}

func TestAuth_AdminLoginForToken(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		loginInput := &authsvc.UserLoginInput{
			Username:  premadeAdminUser.Username,
			Password:  adminUserPassword,
			TotpToken: generateTOTPCodeForUserForTest(t, premadeAdminUser),
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
			Input: loginInput,
		})
		assert.NoError(t, err)
		assert.NotNil(t, tokenRes)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)
	})

	T.Run("non-admin users cannot login via this route", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user := createServiceUserForTest(t, true, fakes.BuildFakeUserRegistrationInput())

		loginInput := &authsvc.UserLoginInput{
			Username:  user.Username,
			Password:  user.HashedPassword,
			TotpToken: generateTOTPCodeForUserForTest(t, user),
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
			Input: loginInput,
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})

	T.Run("with incorrect password", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		loginInput := &authsvc.UserLoginInput{
			Username:  premadeAdminUser.Username,
			Password:  adminUserPassword + "blah",
			TotpToken: generateTOTPCodeForUserForTest(t, premadeAdminUser),
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
			Input: loginInput,
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})

	T.Run("with incorrect TOTP Code", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		loginInput := &authsvc.UserLoginInput{
			Username:  premadeAdminUser.Username,
			Password:  adminUserPassword,
			TotpToken: "000000",
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
			Input: loginInput,
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})

	T.Run("with bogus input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		loginInput := &authsvc.UserLoginInput{
			Username:  " ",
			Password:  "1",
			TotpToken: "otp scode",
		}

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
			Input: loginInput,
		})
		assert.Error(t, err)
		assert.Nil(t, tokenRes)
	})
}

func TestAuth_GetAuthStatus(T *testing.T) {
	T.Parallel()

	T.Run("for unauthenticated user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		res, err := unauthedClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("for logged in user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		res, err := testClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, user.ID, res.UserId)
	})
}

func TestAuth_ChangingPassword(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		_, err := testClient.UpdatePassword(ctx, &authsvc.UpdatePasswordRequest{
			NewPassword:     user.HashedPassword + "blah",
			CurrentPassword: user.HashedPassword,
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		require.NoError(t, err)

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)

		tokenRes, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  user.Username,
				Password:  user.HashedPassword + "blah",
				TotpToken: generateTOTPCodeForUserForTest(t, user),
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, tokenRes)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)
	})

	T.Run("with inadequate new password", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		_, err := testClient.UpdatePassword(ctx, &authsvc.UpdatePasswordRequest{
			NewPassword:     "b",
			CurrentPassword: user.HashedPassword,
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		assert.Error(t, err)
	})
}

func TestAuth_ChangingTOTPSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		res, err := testClient.RefreshTOTPSecret(ctx, &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: user.HashedPassword,
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		require.NoError(t, err)
		res.Result.TwoFactorSecret = user.TwoFactorSecret

		_, err = testClient.VerifyTOTPSecret(ctx, &authsvc.VerifyTOTPSecretRequest{
			TotpToken: generateTOTPCodeForUserForTest(t, user),
			UserId:    user.ID,
		})
		assert.NoError(t, err)
	})

	T.Run("fails with invalid token", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		_, err := testClient.RefreshTOTPSecret(ctx, &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: user.HashedPassword,
			TotpToken:       "000000",
		})
		assert.Error(t, err)
	})

	T.Run("fails with invalid password", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		_, err := testClient.RefreshTOTPSecret(ctx, &authsvc.RefreshTOTPSecretRequest{
			CurrentPassword: user.HashedPassword + "blah",
			TotpToken:       generateTOTPCodeForUserForTest(t, user),
		})
		assert.Error(t, err)
	})
}

func TestAuth_RequestingPasswordReset(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(T)

		res, err := testClient.RequestPasswordResetToken(ctx, &authsvc.RequestPasswordResetTokenRequest{
			EmailAddress: user.EmailAddress,
		})
		require.NoError(t, err)
		assert.NotNil(t, res)

		// boo, hiss, we're talking directly to the database in an _integration test?_ for shame, for shame.
		var token string
		queryErr := databaseClient.DB().QueryRow(`SELECT token FROM password_reset_tokens WHERE belongs_to_user = $1`, user.ID).Scan(&token)
		require.NoError(t, queryErr)

		auditLogRepo := auditlogentries.ProvideAuditLogRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), databaseClient)
		authRepo := authrepo.ProvideAuthRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), auditLogRepo, databaseClient)

		resetToken, err := authRepo.GetPasswordResetTokenByToken(ctx, token)
		require.NoError(t, err)
		require.NotNil(t, resetToken)

		_, err = testClient.RedeemPasswordResetToken(ctx, &authsvc.RedeemPasswordResetTokenRequest{
			Token:       resetToken.Token,
			NewPassword: user.HashedPassword + "blah",
		})
		require.NoError(t, err)

		tokenRes, err := testClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  user.Username,
				Password:  user.HashedPassword + "blah",
				TotpToken: generateTOTPCodeForUserForTest(t, user),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, tokenRes)
		assert.NotEmpty(t, tokenRes.Result.AccessToken)

		// verify that we can't do it twice
		_, err = testClient.RedeemPasswordResetToken(ctx, &authsvc.RedeemPasswordResetTokenRequest{
			Token:       resetToken.Token,
			NewPassword: user.HashedPassword + "blah",
		})
		require.Error(t, err)
	})
}

// TODO section below this line

/*
somewhere here we should validate that a user can't just pretend to be an admin via OAuth somehow?
I feel like that's now how it presently works anyway, but the whole thing is haphazard and fucked up, so
just shore all of it iup.
*/

func TestAuth_InvalidateToken(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		/*
			there's currently no logout mechanism surfaced in the client, but there should be, and it should be tested
		*/
		t.Skipf("TODO")
	})
}
