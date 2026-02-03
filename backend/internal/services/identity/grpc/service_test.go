package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/fakes"
	managermock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/manager/mock"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func buildTestService(t *testing.T) (*serviceImpl, *managermock.IdentityDataManager) {
	t.Helper()

	logger := logging.NewNoopLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	identityDataManager := &managermock.IdentityDataManager{}

	service := &serviceImpl{
		tracer:              tracer,
		logger:              logger,
		identityDataManager: identityDataManager,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID:             identityfakes.BuildFakeID(),
					AccountStatus:      identity.GoodStandingUserAccountStatus.String(),
					ServicePermissions: authorization.NewServiceRolePermissionChecker(authorization.ServiceUserRole.String()),
				},
				ActiveAccountID: identityfakes.BuildFakeID(),
				AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
					identityfakes.BuildFakeID(): authorization.NewAccountRolePermissionChecker(authorization.AccountMemberRole.String()),
				},
			}, nil
		},
	}

	return service, identityDataManager
}

func buildTestServiceWithSessionError(t *testing.T) *serviceImpl {
	t.Helper()

	logger := logging.NewNoopLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	identityDataManager := &managermock.IdentityDataManager{}

	service := &serviceImpl{
		tracer:              tracer,
		logger:              logger,
		identityDataManager: identityDataManager,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return nil, errors.New("session error")
		},
	}

	return service
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		sessionContextDataFetcher := func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{}, nil
		}
		identityDataManager := &managermock.IdentityDataManager{}

		service := NewService(logger, tracerProvider, sessionContextDataFetcher, identityDataManager)

		assert.NotNil(t, service)
		assert.Implements(t, (*identitysvc.IdentityServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.NotNil(t, impl.sessionContextDataFetcher)
		assert.Equal(t, identityDataManager, impl.identityDataManager)
	})
}

func TestServiceImpl_buildResponseDetails(t *testing.T) {
	t.Parallel()

	t.Run("with valid session context", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		ctx := t.Context()

		result := service.buildResponseDetails(ctx, nil)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.NotEmpty(t, result.CurrentAccountId)
	})

	t.Run("with span", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		ctx, span := service.tracer.StartSpan(t.Context())
		defer span.End()

		result := service.buildResponseDetails(ctx, span)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.NotEmpty(t, result.TraceId)
		assert.NotEmpty(t, result.CurrentAccountId)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)
		ctx := t.Context()

		result := service.buildResponseDetails(ctx, nil)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.Empty(t, result.CurrentAccountId)
	})

	t.Run("with nil span", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		ctx := t.Context()

		result := service.buildResponseDetails(ctx, nil)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.Empty(t, result.TraceId)
		assert.NotEmpty(t, result.CurrentAccountId)
	})
}
