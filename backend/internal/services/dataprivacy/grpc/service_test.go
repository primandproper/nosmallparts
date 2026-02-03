package grpc

import (
	"testing"

	dataprivacysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func buildTestService(t *testing.T) *serviceImpl {
	t.Helper()

	logger := logging.NewNoopLogger()
	tracer := tracing.NewTracerForTest(t.Name())

	service := &serviceImpl{
		tracer: tracer,
		logger: logger,
	}

	return service
}

func TestNewDataPrivacyService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()

		service := NewDataPrivacyService(logger, tracerProvider)

		assert.NotNil(t, service)
		assert.Implements(t, (*dataprivacysvc.DataPrivacyServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
	})
}

func TestServiceImpl_AggregateUserDataReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestService(t)

		request := &dataprivacysvc.AggregateUserDataReportRequest{}

		response, err := service.AggregateUserDataReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
	})
}

func TestServiceImpl_DestroyAllUserData(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestService(t)

		request := &dataprivacysvc.DestroyAllUserDataRequest{}

		response, err := service.DestroyAllUserData(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
	})
}

func TestServiceImpl_FetchUserDataReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestService(t)

		request := &dataprivacysvc.FetchUserDataReportRequest{}

		response, err := service.FetchUserDataReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
	})
}
