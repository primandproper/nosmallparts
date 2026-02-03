package authentication

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	mockauthn "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens/paseto"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/fakes"
	identitymock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/mock"
	oauthmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/random"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/testutils"

	oauth2errors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProvideOAuth2ClientManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		cfg := &OAuth2Config{
			Domain: "example.com",
		}
		dataManager := &oauthmock.RepositoryMock{}

		manager := ProvideOAuth2ClientManager(logger, tracerProvider, cfg, dataManager)

		assert.NotNil(t, manager)
	})
}

func TestProvideOAuth2ServerImplementation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		identityRepo := &identitymock.RepositoryMock{}
		authenticator := &mockauthn.Authenticator{}

		ctx := t.Context()
		signingKey := random.MustGenerateRawBytes(ctx, 32)
		tokenIssuer, err := paseto.NewPASETOSigner(logger, tracerProvider, t.Name(), signingKey)
		require.NoError(t, err)

		cfg := &OAuth2Config{
			Domain: "example.com",
		}
		dataManager := &oauthmock.RepositoryMock{}
		manager := ProvideOAuth2ClientManager(logger, tracerProvider, cfg, dataManager)

		server := ProvideOAuth2ServerImplementation(logger, tracerProvider, identityRepo, authenticator, tokenIssuer, manager)

		assert.NotNil(t, server)
	})
}

func TestBuildOAuth2ErrorHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		handler := buildOAuth2ErrorHandler(logger)

		assert.NotNil(t, handler)

		// Test that the handler doesn't panic
		assert.NotPanics(t, func() {
			handler(&oauth2errors.Response{
				Error: errors.New("test error"),
			})
		})
	})
}

func TestBuildInternalErrorHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		handler := buildInternalErrorHandler(logger)

		assert.NotNil(t, handler)

		testErr := errors.New("test error")
		result := handler(testErr)

		assert.NotNil(t, result)
		assert.Equal(t, testErr, result.Error)
		assert.Equal(t, -1, result.ErrorCode)
		assert.Equal(t, testErr.Error(), result.Description)
		assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	})
}

func TestBuildClientInfoHandler(T *testing.T) {
	T.Parallel()

	T.Run("with form data", func(t *testing.T) {
		t.Parallel()

		handler := buildClientInfoHandler()
		assert.NotNil(t, handler)

		req := &http.Request{
			Form: url.Values{
				"client_id":     []string{"test-client-id"},
				"client_secret": []string{"test-client-secret"},
			},
		}

		clientID, clientSecret, err := handler(req)

		assert.NoError(t, err)
		assert.Equal(t, "test-client-id", clientID)
		assert.Equal(t, "test-client-secret", clientSecret)
	})

	T.Run("with basic auth", func(t *testing.T) {
		t.Parallel()

		handler := buildClientInfoHandler()
		assert.NotNil(t, handler)

		req := &http.Request{
			Header: http.Header{
				"Authorization": []string{"Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1jbGllbnQtc2VjcmV0"},
			},
			Form: url.Values{},
		}

		clientID, clientSecret, err := handler(req)

		assert.NoError(t, err)
		assert.Equal(t, "test-client-id", clientID)
		assert.Equal(t, "test-client-secret", clientSecret)
	})

	T.Run("with no auth", func(t *testing.T) {
		t.Parallel()

		handler := buildClientInfoHandler()
		assert.NotNil(t, handler)

		req := &http.Request{
			Form: url.Values{},
		}

		clientID, clientSecret, err := handler(req)

		assert.Error(t, err)
		assert.Equal(t, oauth2errors.ErrInvalidClient, err)
		assert.Empty(t, clientID)
		assert.Empty(t, clientSecret)
	})
}

func TestBuildPasswordAuthorizationHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		user := fakes.BuildFakeUser()

		authenticator := &mockauthn.Authenticator{}
		authenticator.On(
			"CredentialsAreValid",
			testutils.ContextMatcher,
			user.HashedPassword,
			"password",
			"",
			"",
		).Return(true, nil)

		dataManager := &identitymock.RepositoryMock{}
		dataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			user.Username,
		).Return(user, nil)

		handler := buildPasswordAuthorizationHandler(logger, authenticator, dataManager)
		assert.NotNil(t, handler)

		userID, err := handler(ctx, "client-id", user.Username, "password")

		assert.NoError(t, err)
		assert.Equal(t, user.ID, userID)

		mock.AssertExpectationsForObjects(t, authenticator, dataManager)
	})

	T.Run("with invalid username", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		authenticator := &mockauthn.Authenticator{}
		dataManager := &identitymock.RepositoryMock{}
		dataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			"unknown-user",
		).Return((*identity.User)(nil), errors.New("user not found"))

		handler := buildPasswordAuthorizationHandler(logger, authenticator, dataManager)
		assert.NotNil(t, handler)

		userID, err := handler(ctx, "client-id", "unknown-user", "password")

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "invalid username or password")

		mock.AssertExpectationsForObjects(t, authenticator, dataManager)
	})

	T.Run("with invalid credentials", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		user := fakes.BuildFakeUser()

		authenticator := &mockauthn.Authenticator{}
		authenticator.On(
			"CredentialsAreValid",
			testutils.ContextMatcher,
			user.HashedPassword,
			"wrong-password",
			"",
			"",
		).Return(false, nil)

		dataManager := &identitymock.RepositoryMock{}
		dataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			user.Username,
		).Return(user, nil)

		handler := buildPasswordAuthorizationHandler(logger, authenticator, dataManager)
		assert.NotNil(t, handler)

		userID, err := handler(ctx, "client-id", user.Username, "wrong-password")

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "invalid username or password")

		mock.AssertExpectationsForObjects(t, authenticator, dataManager)
	})

	T.Run("with validation error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		user := fakes.BuildFakeUser()

		authenticator := &mockauthn.Authenticator{}
		authenticator.On(
			"CredentialsAreValid",
			testutils.ContextMatcher,
			user.HashedPassword,
			"password",
			"",
			"",
		).Return(false, errors.New("validation error"))

		dataManager := &identitymock.RepositoryMock{}
		dataManager.On(
			"GetUserByUsername",
			testutils.ContextMatcher,
			user.Username,
		).Return(user, nil)

		handler := buildPasswordAuthorizationHandler(logger, authenticator, dataManager)
		assert.NotNil(t, handler)

		userID, err := handler(ctx, "client-id", user.Username, "password")

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "invalid username or password")

		mock.AssertExpectationsForObjects(t, authenticator, dataManager)
	})
}

func TestBuildUserAuthorizationHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracer := tracing.NewTracer(tracing.NewNoopTracerProvider().Tracer("test"))

		signingKey := random.MustGenerateRawBytes(ctx, 32)
		tokenIssuer, err := paseto.NewPASETOSigner(logger, tracing.NewNoopTracerProvider(), t.Name(), signingKey)
		require.NoError(t, err)

		user := fakes.BuildFakeUser()
		token, err := tokenIssuer.IssueToken(ctx, user, time.Hour)
		require.NoError(t, err)

		req := &http.Request{
			Header: http.Header{
				"Authorization": []string{"Bearer " + token},
			},
		}
		req = req.WithContext(ctx)

		handler := buildUserAuthorizationHandler(tracer, logger, tokenIssuer)
		assert.NotNil(t, handler)

		userID, err := handler(nil, req)

		assert.NoError(t, err)
		assert.Equal(t, user.ID, userID)
	})

	T.Run("with invalid token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracer := tracing.NewTracer(tracing.NewNoopTracerProvider().Tracer("test"))

		signingKey := random.MustGenerateRawBytes(ctx, 32)
		tokenIssuer, err := paseto.NewPASETOSigner(logger, tracing.NewNoopTracerProvider(), t.Name(), signingKey)
		require.NoError(t, err)

		req := &http.Request{
			Header: http.Header{
				"Authorization": []string{"Bearer invalid-token"},
			},
		}
		req = req.WithContext(ctx)

		handler := buildUserAuthorizationHandler(tracer, logger, tokenIssuer)
		assert.NotNil(t, handler)

		userID, err := handler(nil, req)

		assert.Error(t, err)
		assert.Equal(t, oauth2errors.ErrAccessDenied, err)
		assert.Empty(t, userID)
	})
}

func TestAuthorizeScopeHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		handler := AuthorizeScopeHandler(logger)
		assert.NotNil(t, handler)

		req := &http.Request{
			URL: &url.URL{
				RawQuery: "scope=read%20write",
			},
		}

		scope, err := handler(nil, req)

		assert.NoError(t, err)
		assert.Equal(t, "read write", scope)
	})
}

func TestAccessTokenExpHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		handler := AccessTokenExpHandler(logger)
		assert.NotNil(t, handler)

		duration, err := handler(nil, nil)

		assert.NoError(t, err)
		assert.Equal(t, 24*time.Hour, duration)
	})
}

func TestClientScopeHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		handler := ClientScopeHandler(logger)
		assert.NotNil(t, handler)

		allowed, err := handler(nil)

		assert.NoError(t, err)
		assert.True(t, allowed)
	})
}
