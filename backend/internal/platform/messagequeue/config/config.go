package msgconfig

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/pubsub"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/redis"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/sqs"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	ps "cloud.google.com/go/pubsub/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ProviderRedis is used to refer to redis.
	ProviderRedis provider = "redis"
	// ProviderSQS is used to refer to sqs.
	ProviderSQS provider = "sqs"
	// ProviderPubSub is used to refer to GCP Pub/Sub.
	ProviderPubSub provider = "pubsub"
)

var (
	ErrNilConfig = errors.New("nil config provided")
)

type (
	// provider is used to indicate what messaging provider we'll use.
	provider string

	// MessageQueueConfig is used to indicate how the messaging provider should be configured.
	MessageQueueConfig struct {
		_ struct{} `json:"-"`

		Provider provider      `env:"PROVIDER"      json:"provider,omitempty"`
		SQS      sqs.Config    `envPrefix:"SQS_"    json:"sqs,omitempty"`
		PubSub   pubsub.Config `envPrefix:"PUBSUB_" json:"pubSub,omitempty"`
		Redis    redis.Config  `envPrefix:"REDIS_"  json:"redis,omitempty"`
	}

	// Config is used to indicate how the messaging provider should be configured.
	Config struct {
		_ struct{} `json:"-"`

		Consumer  MessageQueueConfig `envPrefix:"CONSUMER_"  json:"consumers,omitempty"`
		Publisher MessageQueueConfig `envPrefix:"PUBLISHER_" json:"publishers,omitempty"`
	}

	// QueuesConfig contains the various queue names.
	QueuesConfig struct {
		_ struct{} `json:"-"`

		DataChangesTopicName              string `env:"DATA_CHANGES_TOPIC_NAME"               json:"dataChangesTopicName"`
		OutboundEmailsTopicName           string `env:"OUTBOUND_EMAILS_TOPIC_NAME"            json:"outboundEmailsTopicName"`
		SearchIndexRequestsTopicName      string `env:"SEARCH_INDEX_REQUESTS_TOPIC_NAME"      json:"searchIndexRequestsTopicName"`
		UserDataAggregationTopicName      string `env:"USER_DATA_AGGREGATION_TOPIC_NAME"      json:"userDataAggregationTopicName"`
		WebhookExecutionRequestsTopicName string `env:"WEBHOOK_EXECUTION_REQUESTS_TOPIC_NAME" json:"webhookExecutionRequestsTopicName"`
	}
)

var _ validation.ValidatableWithContext = (*QueuesConfig)(nil)

// ValidateWithContext validates a QueuesConfig struct.
func (c *QueuesConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.DataChangesTopicName, validation.Required),
		validation.Field(&c.OutboundEmailsTopicName, validation.Required),
		validation.Field(&c.SearchIndexRequestsTopicName, validation.Required),
		validation.Field(&c.UserDataAggregationTopicName, validation.Required),
		validation.Field(&c.WebhookExecutionRequestsTopicName, validation.Required),
	)
}

func cleanString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ProvideConsumerProvider provides a PublisherProvider.
func ProvideConsumerProvider(ctx context.Context, logger logging.Logger, c *Config) (messagequeue.ConsumerProvider, error) {
	if c == nil {
		return nil, ErrNilConfig
	}

	switch cleanString(string(c.Consumer.Provider)) {
	case string(ProviderRedis):
		return redis.ProvideRedisConsumerProvider(logger, c.Consumer.Redis), nil
	case string(ProviderPubSub):
		client, err := ps.NewClient(ctx, c.Consumer.PubSub.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("establishing PubSub client: %w", err)
		}

		return pubsub.ProvidePubSubConsumerProvider(logger, client), nil
	default:
		return nil, fmt.Errorf("invalid provider: %q", c.Consumer.Provider)
	}
}

// ProvidePublisherProvider provides a PublisherProvider.
func ProvidePublisherProvider(ctx context.Context, logger logging.Logger, tracerProvider tracing.TracerProvider, c *Config) (messagequeue.PublisherProvider, error) {
	if c == nil {
		return nil, ErrNilConfig
	}

	switch cleanString(string(c.Publisher.Provider)) {
	case string(ProviderRedis):
		return redis.ProvideRedisPublisherProvider(logger, tracerProvider, c.Publisher.Redis), nil
	case string(ProviderSQS):
		return sqs.ProvideSQSPublisherProvider(logger, tracerProvider), nil
	case string(ProviderPubSub):
		client, err := ps.NewClient(ctx, c.Publisher.PubSub.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("establishing PubSub client: %w", err)
		}

		return pubsub.ProvidePubSubPublisherProvider(logger, tracerProvider, client), nil
	default:
		return nil, fmt.Errorf("invalid publisher provider: %q", c.Publisher.Provider)
	}
}
