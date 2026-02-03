package messagequeue

import (
	"context"
)

type (
	// Consumer produces events onto a queue.
	Consumer interface {
		Consume(stopChan chan bool, errors chan error)
	}

	ConsumerFunc func(context.Context, []byte) error

	// ConsumerProvider is a function that provides a Consumer for a given topic.
	ConsumerProvider interface {
		ProvideConsumer(ctx context.Context, topic string, handlerFunc ConsumerFunc) (Consumer, error)
	}
)

var _ Consumer = (*NoopConsumer)(nil)

// NoopConsumer is a no-op [messagequeue.Consumer].
type NoopConsumer struct{}

// Consume does nothing.
func (n *NoopConsumer) Consume(_ chan bool, _ chan error) {}

// NewNoopConsumer is a noop Consumer.
func NewNoopConsumer() *NoopConsumer {
	return &NoopConsumer{}
}

type noopConsumerProvider struct{}

// ProvideConsumer does nothing.
func (n *noopConsumerProvider) ProvideConsumer(context.Context, string, ConsumerFunc) (Consumer, error) {
	return NewNoopConsumer(), nil
}

// NewNoopConsumerProvider is a noop ConsumerProvider.
func NewNoopConsumerProvider() ConsumerProvider {
	return &noopConsumerProvider{}
}
