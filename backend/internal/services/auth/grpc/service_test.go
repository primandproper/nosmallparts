package grpc

import (
	"context"
	"testing"

	authenticationmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	authmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth/mock"
	identitymock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/mock"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func buildTestService(t *testing.T) (*serviceImpl, *identitymock.RepositoryMock, *authmock.AuthManager, *authenticationmock.Manager) {
	t.Helper()

	logger := logging.NewNoopLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	identityRepo := &identitymock.RepositoryMock{}
	authManager := &authmock.AuthManager{}
	authenticationManager := &authenticationmock.Manager{}

	service := &serviceImpl{
		tracer:                tracer,
		logger:                logger,
		identityRepository:    identityRepo,
		authManager:           authManager,
		authenticationManager: authenticationManager,
	}

	return service, identityRepo, authManager, authenticationManager
}

func TestNewAuthService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		identityRepo := &identitymock.RepositoryMock{}
		authManager := &authmock.AuthManager{}
		authenticationManager := &authenticationmock.Manager{}

		service := NewAuthService(logger, tracerProvider, identityRepo, authManager, authenticationManager)

		assert.NotNil(t, service)
		assert.Implements(t, (*authsvc.AuthServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.Equal(t, identityRepo, impl.identityRepository)
		assert.Equal(t, authManager, impl.authManager)
		assert.Equal(t, authenticationManager, impl.authenticationManager)
	})
}

func TestServiceImpl_fetchSessionContext(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)

		sessionContextData := &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID:                   "test-user-id",
				AccountStatus:            "active",
				AccountStatusExplanation: "",
				ServicePermissions:       authorization.NewServiceRolePermissionChecker("service_admin"),
			},
			ActiveAccountID: "test-account-id",
			AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
				"test-account-id": authorization.NewAccountRolePermissionChecker("account_admin"),
			},
		}

		ctx := context.WithValue(t.Context(), sessions.SessionContextDataKey, sessionContextData)

		result, err := service.fetchSessionContext(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sessionContextData, result)
		assert.Equal(t, "test-user-id", result.GetUserID())
		assert.Equal(t, "test-account-id", result.GetActiveAccountID())
	})

	t.Run("missing session context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := t.Context()

		result, err := service.fetchSessionContext(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "session context not found")
	})

	t.Run("wrong type in context", func(t *testing.T) {
		t.Parallel()

		service, _, _, _ := buildTestService(t)
		ctx := context.WithValue(t.Context(), sessions.SessionContextDataKey, "wrong-type")

		result, err := service.fetchSessionContext(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "session context not found")
	})
}
