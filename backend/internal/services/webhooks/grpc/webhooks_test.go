package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	webhookfakes "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks/fakes"
	webhookmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks/mock"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/filtering"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/reflection"
	"github.com/verygoodsoftwarenotvirus/platform/v4/testutils"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/webhooks/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) (*serviceImpl, *webhookmock.Repository) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	webhookRepo := &webhookmock.Repository{}

	service := &serviceImpl{
		tracer: tracer,
		logger: logger,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				ActiveAccountID: "test-account-id",
			}, nil
		},
		webhookRepository: webhookRepo,
	}

	return service, webhookRepo
}

func buildTestServiceWithSessionError(t *testing.T) *serviceImpl {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())

	service := &serviceImpl{
		tracer: tracer,
		logger: logger,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return nil, errors.New("session error")
		},
		webhookRepository: &webhookmock.Repository{},
	}

	return service
}

func TestServiceImpl_CreateWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		fakeWebhook := webhookfakes.BuildFakeWebhook()
		fakeInput := webhookfakes.BuildFakeWebhookCreationRequestInput()

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateWebhook), testutils.ContextMatcher, mock.AnythingOfType("*webhooks.WebhookDatabaseCreationInput")).Return(fakeWebhook, nil)

		request := &webhookssvc.CreateWebhookRequest{
			Input: &webhookssvc.WebhookCreationRequestInput{
				Name:        fakeInput.Name,
				ContentType: converters.ConvertStringToWebhookContentType(fakeInput.ContentType),
				Method:      converters.ConvertStringToWebhookMethod(fakeInput.Method),
				Url:         fakeInput.URL,
				Events:      fakeInput.Events,
			},
		}

		response, err := service.CreateWebhook(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeWebhook.ID, response.Created.Id)
		assert.Equal(t, fakeWebhook.Name, response.Created.Name)
		assert.Equal(t, fakeWebhook.URL, response.Created.Url)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service := buildTestServiceWithSessionError(t)

		request := &webhookssvc.CreateWebhookRequest{
			Input: &webhookssvc.WebhookCreationRequestInput{
				Name:        "test webhook",
				Url:         "https://example.com/webhook",
				Method:      webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST,
				ContentType: webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON,
				Events:      []string{"test_event"},
			},
		}

		response, err := service.CreateWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, _ := buildTestService(t)

		// Invalid request with empty name
		request := &webhookssvc.CreateWebhookRequest{
			Input: &webhookssvc.WebhookCreationRequestInput{
				Name:        "", // Invalid empty name
				Method:      webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST,
				ContentType: webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON,
				Url:         "https://example.com/webhook",
				Events:      []string{"test_event"},
			},
		}

		response, err := service.CreateWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		fakeInput := webhookfakes.BuildFakeWebhookCreationRequestInput()

		mockRepo.On(reflection.GetMethodName(mockRepo.CreateWebhook), testutils.ContextMatcher, mock.AnythingOfType("*webhooks.WebhookDatabaseCreationInput")).Return(nil, errors.New("repository error"))

		request := &webhookssvc.CreateWebhookRequest{
			Input: &webhookssvc.WebhookCreationRequestInput{
				Name:        fakeInput.Name,
				Url:         fakeInput.URL,
				Method:      webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST,
				ContentType: webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON,
				Events:      fakeInput.Events,
			},
		}

		response, err := service.CreateWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_AddWebhookTriggerEvent(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		fakeEvent := webhookfakes.BuildFakeWebhookTriggerEvent()

		mockRepo.On(reflection.GetMethodName(mockRepo.AddWebhookTriggerEvent), testutils.ContextMatcher, "test-account-id", mock.AnythingOfType("*webhooks.WebhookTriggerEventDatabaseCreationInput")).Return(fakeEvent, nil)

		request := &webhookssvc.AddWebhookTriggerEventRequest{
			Input: &webhookssvc.WebhookTriggerEventCreationRequestInput{
				BelongsToWebhook: fakeEvent.BelongsToWebhook,
				TriggerEvent:     fakeEvent.TriggerEvent,
			},
		}

		response, err := service.AddWebhookTriggerEvent(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Created)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeEvent.ID, response.Created.Id)
		assert.Equal(t, fakeEvent.TriggerEvent, response.Created.TriggerEvent)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service := buildTestServiceWithSessionError(t)

		request := &webhookssvc.AddWebhookTriggerEventRequest{
			Input: &webhookssvc.WebhookTriggerEventCreationRequestInput{
				BelongsToWebhook: "test-webhook-id",
				TriggerEvent:     "test_event",
			},
		}

		response, err := service.AddWebhookTriggerEvent(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		mockRepo.On(reflection.GetMethodName(mockRepo.AddWebhookTriggerEvent), testutils.ContextMatcher, "test-account-id", mock.AnythingOfType("*webhooks.WebhookTriggerEventDatabaseCreationInput")).Return(nil, errors.New("repository error"))

		request := &webhookssvc.AddWebhookTriggerEventRequest{
			Input: &webhookssvc.WebhookTriggerEventCreationRequestInput{
				BelongsToWebhook: "test-webhook-id",
				TriggerEvent:     "test_event",
			},
		}

		response, err := service.AddWebhookTriggerEvent(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		fakeWebhook := webhookfakes.BuildFakeWebhook()
		webhookID := "test-webhook-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWebhook), testutils.ContextMatcher, webhookID, "test-account-id").Return(fakeWebhook, nil)

		request := &webhookssvc.GetWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.GetWebhook(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Result)
		assert.NotNil(t, response.ResponseDetails)
		assert.Equal(t, fakeWebhook.ID, response.Result.Id)
		assert.Equal(t, fakeWebhook.Name, response.Result.Name)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service := buildTestServiceWithSessionError(t)

		request := &webhookssvc.GetWebhookRequest{
			WebhookId: "test-webhook-id",
		}

		response, err := service.GetWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		webhookID := "test-webhook-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWebhook), testutils.ContextMatcher, webhookID, "test-account-id").Return(nil, errors.New("repository error"))

		request := &webhookssvc.GetWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.GetWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetWebhooks(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		fakeWebhooks := webhookfakes.BuildFakeWebhooksList()

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWebhooks), testutils.ContextMatcher, "test-account-id", testutils.QueryFilterMatcher).Return(fakeWebhooks, nil)

		request := &webhookssvc.GetWebhooksRequest{
			Filter: &grpcfiltering.QueryFilter{
				// Add any filter fields as needed
			},
		}

		response, err := service.GetWebhooks(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeWebhooks.Data))
		assert.Equal(t, fakeWebhooks.Data[0].ID, response.Results[0].Id)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service := buildTestServiceWithSessionError(t)

		request := &webhookssvc.GetWebhooksRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWebhooks(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		mockRepo.On(reflection.GetMethodName(mockRepo.GetWebhooks), testutils.ContextMatcher, "test-account-id", testutils.QueryFilterMatcher).Return(nil, errors.New("repository error"))

		request := &webhookssvc.GetWebhooksRequest{
			Filter: &grpcfiltering.QueryFilter{},
		}

		response, err := service.GetWebhooks(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_ArchiveWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		webhookID := "test-webhook-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWebhook), testutils.ContextMatcher, webhookID, "test-account-id").Return(nil)

		request := &webhookssvc.ArchiveWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.ArchiveWebhook(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service := buildTestServiceWithSessionError(t)

		request := &webhookssvc.ArchiveWebhookRequest{
			WebhookId: "test-webhook-id",
		}

		response, err := service.ArchiveWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		webhookID := "test-webhook-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWebhook), testutils.ContextMatcher, webhookID, "test-account-id").Return(errors.New("repository error"))

		request := &webhookssvc.ArchiveWebhookRequest{
			WebhookId: webhookID,
		}

		response, err := service.ArchiveWebhook(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_ArchiveWebhookTriggerEvent(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		webhookID := "test-webhook-id"
		eventID := "test-event-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWebhookTriggerEvent), testutils.ContextMatcher, webhookID, eventID).Return(nil)

		request := &webhookssvc.ArchiveWebhookTriggerEventRequest{
			WebhookId:             webhookID,
			WebhookTriggerEventId: eventID,
		}

		response, err := service.ArchiveWebhookTriggerEvent(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		service, mockRepo := buildTestService(t)

		webhookID := "test-webhook-id"
		eventID := "test-event-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.ArchiveWebhookTriggerEvent), testutils.ContextMatcher, webhookID, eventID).Return(errors.New("repository error"))

		request := &webhookssvc.ArchiveWebhookTriggerEventRequest{
			WebhookId:             webhookID,
			WebhookTriggerEventId: eventID,
		}

		response, err := service.ArchiveWebhookTriggerEvent(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	t.Run("implements WebhooksServiceServer", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		assert.Implements(t, (*webhookssvc.WebhooksServiceServer)(nil), service)
	})
}
