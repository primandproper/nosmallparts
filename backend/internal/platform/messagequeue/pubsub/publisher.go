package pubsub

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
)

type (
	messagePublisher interface {
		Stop()
		Publish(context.Context, *pubsub.Message) *pubsub.PublishResult
	}

	pubSubPublisher struct {
		tracer    tracing.Tracer
		encoder   encoding.ClientEncoder
		logger    logging.Logger
		publisher messagePublisher
		topic     string
	}
)

// buildPubSubPublisher provides a Pub/Sub-backed pubSubPublisher.
func buildPubSubPublisher(logger logging.Logger, pubsubClient *pubsub.Publisher, tracerProvider tracing.TracerProvider, topic string) *pubSubPublisher {
	return &pubSubPublisher{
		topic:     topic,
		encoder:   encoding.ProvideClientEncoder(logger, tracerProvider, encoding.ContentTypeJSON),
		logger:    logging.EnsureLogger(logger),
		publisher: pubsubClient,
		tracer:    tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(fmt.Sprintf("%s_publisher", topic))),
	}
}

// Stop calls Stop on the topic.
func (p *pubSubPublisher) Stop() {
	p.publisher.Stop()
}

type publisherProvider struct {
	logger            logging.Logger
	publisherCache    map[string]messagequeue.Publisher
	pubsubClient      *pubsub.Client
	tracerProvider    tracing.TracerProvider
	publisherCacheHat sync.RWMutex
}

// ProvidePubSubPublisherProvider returns a PublisherProvider for a given address.
func ProvidePubSubPublisherProvider(logger logging.Logger, tracerProvider tracing.TracerProvider, client *pubsub.Client) messagequeue.PublisherProvider {
	return &publisherProvider{
		logger:         logging.EnsureLogger(logger),
		pubsubClient:   client,
		publisherCache: map[string]messagequeue.Publisher{},
		tracerProvider: tracerProvider,
	}
}

// Close closes the connection topic.
func (p *publisherProvider) Close() {
	if err := p.pubsubClient.Close(); err != nil {
		p.logger.Error("closing pubsub connection", err)
	}
}

// ProvidePublisher returns a pubSubPublisher for a given topic.
func (p *publisherProvider) ProvidePublisher(ctx context.Context, topicName string) (messagequeue.Publisher, error) {
	if topicName == "" {
		return nil, messagequeue.ErrEmptyTopicName
	}

	logger := logging.EnsureLogger(p.logger.Clone())

	p.publisherCacheHat.Lock()
	defer p.publisherCacheHat.Unlock()
	if cachedPub, ok := p.publisherCache[topicName]; ok {
		return cachedPub, nil
	}

	topic, err := p.pubsubClient.TopicAdminClient.GetTopic(ctx, &pubsubpb.GetTopicRequest{Topic: topicName})
	if err != nil {
		return nil, fmt.Errorf("error getting topic admin client: %w", err)
	}

	publisher := p.pubsubClient.Publisher(topic.GetName())

	pub := buildPubSubPublisher(logger, publisher, p.tracerProvider, topicName)
	p.publisherCache[topicName] = pub

	return pub, nil
}

func (p *pubSubPublisher) Publish(ctx context.Context, data any) error {
	_, span := p.tracer.StartSpan(ctx)
	defer span.End()

	logger := p.logger.Clone()

	var b bytes.Buffer
	if err := p.encoder.Encode(ctx, &b, data); err != nil {
		return observability.PrepareError(err, span, "encoding topic message")
	}

	msg := &pubsub.Message{Data: b.Bytes()}
	result := p.publisher.Publish(ctx, msg)

	<-result.Ready()

	// The Get method blocks until a server-generated ID or an error is returned for the published message.
	if _, err := result.Get(ctx); err != nil {
		observability.AcknowledgeError(err, logger, span, "publishing pubsub message")
	}

	logger.Debug("published message")

	return nil
}

func (p *pubSubPublisher) PublishAsync(ctx context.Context, data any) {
	_, span := p.tracer.StartSpan(ctx)
	defer span.End()

	logger := p.logger.Clone()

	var b bytes.Buffer
	if err := p.encoder.Encode(ctx, &b, data); err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding topic message")
	}

	msg := &pubsub.Message{Data: b.Bytes()}
	result := p.publisher.Publish(ctx, msg)
	<-result.Ready()

	// The Get method blocks until a server-generated ID or an error is returned for the published message.
	if _, err := result.Get(ctx); err != nil {
		observability.AcknowledgeError(err, logger, span, "publishing pubsub message")
	}

	logger.Debug("published message")
}
