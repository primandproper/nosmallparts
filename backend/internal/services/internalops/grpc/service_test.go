package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	settingssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/internalops"
	msgconfig "github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) *serviceImpl {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	msgConfig := &msgconfig.Config{}

	service := &serviceImpl{
		tracer:    tracer,
		logger:    logger,
		msgConfig: msgConfig,
		sessionContextDataFetcher: func(ctx context.Context) (sessions.ContextData, error) {
			return sessions.ContextData{
				Requester: sessions.RequesterInfo{
					ServicePermissions: nil,
				},
			}, nil
		},
	}

	return service
}

func buildTestServiceWithSessionError(t *testing.T) *serviceImpl {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	msgConfig := &msgconfig.Config{}

	service := &serviceImpl{
		tracer:    tracer,
		logger:    logger,
		msgConfig: msgConfig,
		sessionContextDataFetcher: func(ctx context.Context) (sessions.ContextData, error) {
			return sessions.ContextData{}, errors.New("session error")
		},
	}

	return service
}

func buildTestServiceWithInsufficientPermissions(t *testing.T) *serviceImpl {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	msgConfig := &msgconfig.Config{}

	service := &serviceImpl{
		tracer:    tracer,
		logger:    logger,
		msgConfig: msgConfig,
		sessionContextDataFetcher: func(ctx context.Context) (sessions.ContextData, error) {
			return sessions.ContextData{
				Requester: sessions.RequesterInfo{
					ServicePermissions: nil, // No permissions
				},
			}, nil
		},
	}

	return service
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		msgConfig := &msgconfig.Config{}

		service := NewService(logger, tracerProvider, msgConfig)

		assert.NotNil(t, service)
		assert.Implements(t, (*settingssvc.InternalOperationsServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.Equal(t, msgConfig, impl.msgConfig)
	})
}

func TestServiceImpl_PublishArbitraryQueueMessage(t *testing.T) {
	t.Parallel()

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &settingssvc.PublishArbitraryQueueMessageRequest{
			QueueName: "test-queue",
			Body:      "test message",
		}

		response, err := service.PublishArbitraryQueueMessage(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}
