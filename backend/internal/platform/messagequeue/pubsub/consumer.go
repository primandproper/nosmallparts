package pubsub

import (
	"context"
	"fmt"
	"sync"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
)

type (
	pubSubConsumer struct {
		logger      logging.Logger
		consumer    *pubsub.Client
		handlerFunc func(context.Context, []byte) error
		topic       string
	}
)

// buildPubSubConsumer provides a Pub/Sub-backed pubSubConsumer.
func buildPubSubConsumer(
	logger logging.Logger,
	pubsubClient *pubsub.Client,
	topic string,
	handlerFunc func(context.Context, []byte) error,
) messagequeue.Consumer {
	return &pubSubConsumer{
		topic:       topic,
		logger:      logging.EnsureLogger(logger),
		consumer:    pubsubClient,
		handlerFunc: handlerFunc,
	}
}

func (c *pubSubConsumer) Consume(stopChan chan bool, errors chan error) {
	if stopChan == nil {
		stopChan = make(chan bool, 1)
	}

	ctx := context.Background()
	sub, err := c.consumer.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name:  c.topic,
		Topic: c.topic,
	})
	if err != nil {
		c.logger.Error(fmt.Sprintf("creating %s subscription", c.topic), err)
		errors <- err
		return
	}

	subscriber := c.consumer.Subscriber(sub.GetName())

	go func() {
		<-stopChan
		if err = c.consumer.SubscriptionAdminClient.DeleteSubscription(ctx, &pubsubpb.DeleteSubscriptionRequest{Subscription: sub.GetName()}); err != nil {
			c.logger.Error(fmt.Sprintf("deleting %s subscription", c.topic), err)
			errors <- err
		}
	}()

	if err = subscriber.Receive(ctx, func(receivedContext context.Context, m *pubsub.Message) {
		if handleErr := c.handlerFunc(receivedContext, m.Data); handleErr != nil {
			errors <- err
		} else {
			m.Ack()
		}
	}); err != nil {
		c.logger.Error(fmt.Sprintf("receiving %s pub/sub data", c.topic), err)
	}
}

type pubsubConsumerProvider struct {
	logger           logging.Logger
	consumerCache    map[string]messagequeue.Consumer
	pubsubClient     *pubsub.Client
	consumerCacheHat sync.RWMutex
}

// ProvidePubSubConsumerProvider returns a ConsumerProvider for a given address.
func ProvidePubSubConsumerProvider(logger logging.Logger, client *pubsub.Client) messagequeue.ConsumerProvider {
	return &pubsubConsumerProvider{
		logger:        logging.EnsureLogger(logger),
		pubsubClient:  client,
		consumerCache: map[string]messagequeue.Consumer{},
	}
}

// Close closes the connection topic.
func (p *pubsubConsumerProvider) Close() {
	if err := p.pubsubClient.Close(); err != nil {
		p.logger.Error("closing pubsub connection", err)
	}
}

// ProvideConsumer returns a pubSubConsumer for a given topic.
func (p *pubsubConsumerProvider) ProvideConsumer(_ context.Context, topic string, handlerFunc messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
	if topic == "" {
		return nil, messagequeue.ErrEmptyTopicName
	}

	logger := logging.EnsureLogger(p.logger.Clone())

	p.consumerCacheHat.Lock()
	defer p.consumerCacheHat.Unlock()
	if cachedPub, ok := p.consumerCache[topic]; ok {
		return cachedPub, nil
	}

	pub := buildPubSubConsumer(logger, p.pubsubClient, topic, handlerFunc)
	p.consumerCache[topic] = pub

	return pub, nil
}
