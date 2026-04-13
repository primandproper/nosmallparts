package grpc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy"
	dataprivacymock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	mockuploads "github.com/verygoodsoftwarenotvirus/platform/v4/uploads/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService(t *testing.T) (*serviceImpl, *dataprivacymock.Repository, *mockuploads.MockUploadManager) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	mockRepo := &dataprivacymock.Repository{}
	mockUploads := &mockuploads.MockUploadManager{}

	exampleUserID := identifiers.New()
	sessionFetcher := func(ctx context.Context) (*sessions.ContextData, error) {
		return &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID: exampleUserID,
			},
		}, nil
	}

	service := &serviceImpl{
		tracer:                    tracer,
		logger:                    logger,
		sessionContextDataFetcher: sessionFetcher,
		dataPrivacyRepo:           mockRepo,
		uploadManager:             mockUploads,
	}

	return service, mockRepo, mockUploads
}

func TestNewDataPrivacyService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		mockRepo := &dataprivacymock.Repository{}
		mockUploads := &mockuploads.MockUploadManager{}
		sessionFetcher := func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{}, nil
		}

		service := NewDataPrivacyService(logger, tracerProvider, sessionFetcher, mockRepo, mockUploads)

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
		service, mockRepo, mockUploads := buildTestService(t)

		collection := &dataprivacy.UserDataCollection{
			Identity: identity.UserDataCollection{
				User: identity.User{ID: identifiers.New()},
			},
		}

		mockRepo.On("FetchUserDataCollection", mock.Anything, mock.AnythingOfType("string")).Return(collection, nil)
		mockUploads.On("SaveFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)

		request := &dataprivacysvc.AggregateUserDataReportRequest{}

		response, err := service.AggregateUserDataReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.NotEmpty(t, response.ReportId)

		mockRepo.AssertExpectations(t)
		mockUploads.AssertExpectations(t)
	})
}

func TestServiceImpl_DestroyAllUserData(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo, _ := buildTestService(t)

		mockRepo.On("DeleteUser", mock.Anything, mock.AnythingOfType("string")).Return(nil)

		request := &dataprivacysvc.DestroyAllUserDataRequest{}

		response, err := service.DestroyAllUserData(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)
		assert.True(t, response.Successful)

		mockRepo.AssertExpectations(t)
	})
}

func TestServiceImpl_FetchUserDataReport(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, mockUploads := buildTestService(t)

		collection := &dataprivacy.UserDataCollection{
			Identity: identity.UserDataCollection{
				User: identity.User{ID: identifiers.New()},
			},
		}
		collectionBytes, _ := json.Marshal(collection)

		mockUploads.On("ReadFile", mock.Anything, mock.AnythingOfType("string")).Return(collectionBytes, nil)

		request := &dataprivacysvc.FetchUserDataReportRequest{
			UserDataAggregationReportId: identifiers.New(),
		}

		response, err := service.FetchUserDataReport(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotEmpty(t, response.ResponseDetails.TraceId)

		mockUploads.AssertExpectations(t)
	})
}
