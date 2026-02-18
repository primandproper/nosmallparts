package datachangemessagehandler

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	dataprivacymock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy/mock"
	identitymock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/mock"
	webhooksmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks/mock"
	analyticsmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/mock"
	emailmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/email/mock"
	encodingmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding/mock"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	msgqueuemock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/reflection"
	uploadsmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/uploads/mock"
	identityindexing "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/indexing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//nolint:gocritic // I know this returns too many things
func buildTestAsyncDataChangeMessageHandler(t *testing.T) (*AsyncDataChangeMessageHandler, *identitymock.RepositoryMock, *webhooksmock.Repository, *msgqueuemock.ConsumerProvider, *msgqueuemock.PublisherProvider, *analyticsmock.EventReporter, *emailmock.Emailer, *uploadsmock.MockUploadManager, *mockmetrics.MetricsProvider, *encodingmock.EncoderDecoder, *dataprivacymock.Repository) {
	t.Helper()

	logger := logging.NewNoopLogger()
	tracer := tracing.NewTracerForTest(t.Name())

	identityRepo := &identitymock.RepositoryMock{}
	webhookRepo := &webhooksmock.Repository{}
	consumerProvider := &msgqueuemock.ConsumerProvider{}
	publisherProvider := &msgqueuemock.PublisherProvider{}
	analyticsEventReporter := &analyticsmock.EventReporter{}
	emailer := &emailmock.Emailer{}
	uploadManager := &uploadsmock.MockUploadManager{}
	metricsProvider := &mockmetrics.MetricsProvider{}
	decoder := &encodingmock.EncoderDecoder{}
	dataPrivacyRepo := &dataprivacymock.Repository{}

	// Create mock indexers with noop implementations for testing
	userDataIndexer := &identityindexing.UserDataIndexer{}

	// Set up mock publishers for the indexers to prevent nil pointer dereferences
	mockPublisher := &msgqueuemock.Publisher{}
	publisherProvider.On(reflection.GetMethodName(publisherProvider.ProvidePublisher), mock.AnythingOfType("string")).Return(mockPublisher, nil).Maybe()

	// Set up mock histograms
	mockHistogram := metrics.NewNoopMetricsProvider()
	noop, _ := mockHistogram.NewFloat64Histogram("test")
	metricsProvider.On(reflection.GetMethodName(metricsProvider.NewFloat64Histogram), mock.AnythingOfType("string"), mock.Anything).Return(noop, nil).Maybe()

	handler := &AsyncDataChangeMessageHandler{
		identityRepo:                         identityRepo,
		webhookRepo:                          webhookRepo,
		consumerProvider:                     consumerProvider,
		analyticsEventReporter:               analyticsEventReporter,
		emailer:                              emailer,
		uploadManager:                        uploadManager,
		decoder:                              decoder,
		userDataIndexer:                      userDataIndexer,
		logger:                               logger,
		tracer:                               tracer,
		nonWebhookEventTypes:                 []string{},
		dataChangesExecutionTimeHistogram:    noop,
		outboundEmailsExecutionTimeHistogram: noop,
		webhookExecutionTimestampHistogram:   noop,
		userDataAggregationExecutionTimeHistogram: noop,
		queuesConfig: msgconfig.QueuesConfig{
			SearchIndexRequestsTopicName: "search-index-requests",
		},
		searchDataIndexPublisher:         mockPublisher,
		outboundEmailsPublisher:          mockPublisher,
		webhookExecutionRequestPublisher: mockPublisher,
		dataPrivacyRepo:                  dataPrivacyRepo,
	}

	return handler, identityRepo, webhookRepo, consumerProvider, publisherProvider, analyticsEventReporter, emailer, uploadManager, metricsProvider, decoder, dataPrivacyRepo
}

func TestNewAsyncDataChangeMessageHandler(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		cfg := &config.AsyncMessageHandlerConfig{
			Queues: msgconfig.QueuesConfig{
				OutboundEmailsTopicName:           "outbound-emails",
				SearchIndexRequestsTopicName:      "search-index-requests",
				WebhookExecutionRequestsTopicName: "webhook-execution-requests",
			},
		}
		identityRepo := &identitymock.RepositoryMock{}
		dataPrivacyRepo := &dataprivacymock.Repository{}
		webhookRepo := &webhooksmock.Repository{}
		consumerProvider := &msgqueuemock.ConsumerProvider{}
		publisherProvider := &msgqueuemock.PublisherProvider{}
		analyticsEventReporter := &analyticsmock.EventReporter{}
		emailer := &emailmock.Emailer{}
		uploadManager := &uploadsmock.MockUploadManager{}
		metricsProvider := &mockmetrics.MetricsProvider{}
		decoder := &encodingmock.EncoderDecoder{}
		coreDataIndexer := &identityindexing.UserDataIndexer{}

		// Set up metrics expectations
		mockHistogram := metrics.NewNoopMetricsProvider()
		noop, _ := mockHistogram.NewFloat64Histogram("test")
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewFloat64Histogram), "data_changes_execution_time", mock.Anything).Return(noop, nil)
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewFloat64Histogram), "outbound_emails_execution_time", mock.Anything).Return(noop, nil)
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewFloat64Histogram), "search_index_requests_execution_time", mock.Anything).Return(noop, nil)
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewFloat64Histogram), "user_data_aggregation_execution_time", mock.Anything).Return(noop, nil)
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewFloat64Histogram), "webhook_requests_execution_time", mock.Anything).Return(noop, nil)

		// Set up publisher expectations
		mockPublisher := &msgqueuemock.Publisher{}
		publisherProvider.On(reflection.GetMethodName(publisherProvider.ProvidePublisher), "outbound-emails").Return(mockPublisher, nil)
		publisherProvider.On(reflection.GetMethodName(publisherProvider.ProvidePublisher), "search-index-requests").Return(mockPublisher, nil)
		publisherProvider.On(reflection.GetMethodName(publisherProvider.ProvidePublisher), "webhook-execution-requests").Return(mockPublisher, nil)

		handler, err := NewAsyncDataChangeMessageHandler(
			ctx,
			logger,
			tracerProvider,
			cfg,
			identityRepo,
			dataPrivacyRepo,
			webhookRepo,
			consumerProvider,
			publisherProvider,
			analyticsEventReporter,
			emailer,
			uploadManager,
			metricsProvider,
			decoder,
			coreDataIndexer,
		)

		assert.NoError(t, err)
		assert.NotNil(t, handler)
		assert.Equal(t, identityRepo, handler.identityRepo)
		assert.Equal(t, webhookRepo, handler.webhookRepo)
		assert.Equal(t, consumerProvider, handler.consumerProvider)
		assert.Equal(t, analyticsEventReporter, handler.analyticsEventReporter)
		assert.Equal(t, emailer, handler.emailer)
		assert.Equal(t, uploadManager, handler.uploadManager)
		assert.Equal(t, decoder, handler.decoder)
		assert.Equal(t, coreDataIndexer, handler.userDataIndexer)

		mock.AssertExpectationsForObjects(t, metricsProvider, publisherProvider)
	})
}

func TestAsyncDataChangeMessageHandler_SetNonWebhookEventTypes(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		eventTypes := []string{"event1", "event2", "event3"}
		handler.SetNonWebhookEventTypes(eventTypes)

		handler.nonWebhookEventTypesHat.RLock()
		assert.Equal(t, eventTypes, handler.nonWebhookEventTypes)
		handler.nonWebhookEventTypesHat.RUnlock()
	})
}
