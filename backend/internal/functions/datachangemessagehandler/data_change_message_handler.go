package datachangemessagehandler

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/platform/v4/analytics"
	"github.com/verygoodsoftwarenotvirus/platform/v4/email"
	"github.com/verygoodsoftwarenotvirus/platform/v4/encoding"
	"github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue"
	msgconfig "github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/uploads"
	identityindexing "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/indexing"
)

const (
	o11yName = "async_data_change_message_handler"
)

var (
	errRequiredDataIsNil = errors.New("required data is nil")
)

type AsyncDataChangeMessageHandler struct {
	searchDataIndexPublisher                  messagequeue.Publisher
	identityRepo                              identity.Repository
	dataPrivacyRepo                           dataprivacy.Repository
	logger                                    logging.Logger
	decoder                                   encoding.ServerEncoderDecoder
	webhookExecutionTimestampHistogram        metrics.Float64Histogram
	userDataAggregationExecutionTimeHistogram metrics.Float64Histogram
	outboundEmailsPublisher                   messagequeue.Publisher
	webhookRepo                               webhooks.Repository
	consumerProvider                          messagequeue.ConsumerProvider
	analyticsEventReporter                    analytics.EventReporter
	dataChangesExecutionTimeHistogram         metrics.Float64Histogram
	webhookExecutionRequestPublisher          messagequeue.Publisher
	emailer                                   email.Emailer
	uploadManager                             uploads.UploadManager
	tracer                                    tracing.Tracer
	outboundEmailsExecutionTimeHistogram      metrics.Float64Histogram
	searchIndexRequestsExecutionTimeHistogram metrics.Float64Histogram
	userDataIndexer                           *identityindexing.UserDataIndexer
	queuesConfig                              msgconfig.QueuesConfig
	nonWebhookEventTypes                      []string
	nonWebhookEventTypesHat                   sync.RWMutex
}

func (a *AsyncDataChangeMessageHandler) SetNonWebhookEventTypes(nonWebhookEventTypes []string) {
	a.nonWebhookEventTypesHat.Lock()
	defer a.nonWebhookEventTypesHat.Unlock()
	a.nonWebhookEventTypes = nonWebhookEventTypes
}

func NewAsyncDataChangeMessageHandler(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	cfg *config.AsyncMessageHandlerConfig,
	identityRepo identity.Repository,
	dataPrivacyRepo dataprivacy.Repository,
	webhookRepo webhooks.Repository,
	consumerProvider messagequeue.ConsumerProvider,
	publisherProvider messagequeue.PublisherProvider,
	analyticsEventReporter analytics.EventReporter,
	emailer email.Emailer,
	uploadManager uploads.UploadManager,
	metricsProvider metrics.Provider,
	decoder encoding.ServerEncoderDecoder,
	coreDataIndexer *identityindexing.UserDataIndexer,
) (*AsyncDataChangeMessageHandler, error) {
	dataChangesExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("data_changes_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up dataChanges execution time histogram: %w", err)
	}

	outboundEmailsExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("outbound_emails_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up outboundEmails execution time histogram: %w", err)
	}

	searchIndexRequestsExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("search_index_requests_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up searchIndexRequests execution time histogram: %w", err)
	}

	userDataAggregationExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("user_data_aggregation_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up userDataAggregation execution time histogram: %w", err)
	}

	webhookExecutionTimestampHistogram, err := metricsProvider.NewFloat64Histogram("webhook_requests_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up webhookExecutionRequests execution time histogram: %w", err)
	}

	outboundEmailsPublisher, err := publisherProvider.ProvidePublisher(ctx, cfg.Queues.OutboundEmailsTopicName)
	if err != nil {
		return nil, fmt.Errorf("configuring outbound emails publisher: %w", err)
	}

	searchDataIndexPublisher, err := publisherProvider.ProvidePublisher(ctx, cfg.Queues.SearchIndexRequestsTopicName)
	if err != nil {
		return nil, fmt.Errorf("configuring search indexing publisher: %w", err)
	}

	webhookExecutionRequestPublisher, err := publisherProvider.ProvidePublisher(ctx, cfg.Queues.WebhookExecutionRequestsTopicName)
	if err != nil {
		return nil, fmt.Errorf("configuring webhook execution requests publisher: %w", err)
	}

	return &AsyncDataChangeMessageHandler{
		tracer:                               tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:                               logging.NewNamedLogger(logger, o11yName),
		nonWebhookEventTypes:                 []string{},
		identityRepo:                         identityRepo,
		dataPrivacyRepo:                      dataPrivacyRepo,
		webhookRepo:                          webhookRepo,
		consumerProvider:                     consumerProvider,
		analyticsEventReporter:               analyticsEventReporter,
		outboundEmailsPublisher:              outboundEmailsPublisher,
		searchDataIndexPublisher:             searchDataIndexPublisher,
		queuesConfig:                         cfg.Queues,
		webhookExecutionRequestPublisher:     webhookExecutionRequestPublisher,
		emailer:                              emailer,
		uploadManager:                        uploadManager,
		dataChangesExecutionTimeHistogram:    dataChangesExecutionTimeHistogram,
		outboundEmailsExecutionTimeHistogram: outboundEmailsExecutionTimeHistogram,
		searchIndexRequestsExecutionTimeHistogram: searchIndexRequestsExecutionTimeHistogram,
		userDataAggregationExecutionTimeHistogram: userDataAggregationExecutionTimeHistogram,
		webhookExecutionTimestampHistogram:        webhookExecutionTimestampHistogram,
		decoder:                                   decoder,
		userDataIndexer:                           coreDataIndexer,
	}, nil
}

func (a *AsyncDataChangeMessageHandler) ConsumeMessages(
	ctx context.Context,
	stopChan chan bool,
	errorsChan chan error,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	// set up myriad publishers

	dataChangesConsumer, err := a.consumerProvider.ProvideConsumer(
		ctx,
		a.queuesConfig.DataChangesTopicName,
		a.DataChangesEventHandler,
	)
	if err != nil {
		return observability.PrepareAndLogError(err, a.logger, span, "configuring data changes consumer")
	}

	outboundEmailsConsumer, err := a.consumerProvider.ProvideConsumer(
		ctx,
		a.queuesConfig.OutboundEmailsTopicName,
		a.OutboundEmailsEventHandler,
	)
	if err != nil {
		return observability.PrepareAndLogError(err, a.logger, span, "configuring outbound emails consumer")
	}

	searchIndexRequestsConsumer, err := a.consumerProvider.ProvideConsumer(
		ctx,
		a.queuesConfig.SearchIndexRequestsTopicName,
		a.SearchIndexRequestsEventHandler,
	)
	if err != nil {
		return observability.PrepareAndLogError(err, a.logger, span, "configuring search index requests consumer")
	}

	webhookExecutionRequestsConsumer, err := a.consumerProvider.ProvideConsumer(
		ctx,
		a.queuesConfig.WebhookExecutionRequestsTopicName,
		a.WebhookExecutionRequestsEventHandler,
	)
	if err != nil {
		return observability.PrepareAndLogError(err, a.logger, span, "configuring webhook execution requests consumer")
	}

	userDataAggregationConsumer, err := a.consumerProvider.ProvideConsumer(
		ctx,
		a.queuesConfig.WebhookExecutionRequestsTopicName,
		a.UserDataAggregationEventHandler,
	)
	if err != nil {
		return observability.PrepareAndLogError(err, a.logger, span, "configuring user data aggregation requests consumer")
	}

	go dataChangesConsumer.Consume(ctx, stopChan, errorsChan)
	go outboundEmailsConsumer.Consume(ctx, stopChan, errorsChan)
	go searchIndexRequestsConsumer.Consume(ctx, stopChan, errorsChan)
	go webhookExecutionRequestsConsumer.Consume(ctx, stopChan, errorsChan)
	go userDataAggregationConsumer.Consume(ctx, stopChan, errorsChan)

	go func() {
		for e := range errorsChan {
			a.logger.Error("consuming message", e)
		}
	}()

	return nil
}
