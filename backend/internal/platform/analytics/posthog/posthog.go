package posthog

import (
	"context"
	"errors"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/internalerrors"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/posthog/posthog-go"
)

const (
	name = "posthog_event_reporter"
)

var (
	// ErrEmptyAPIToken indicates an empty API token was provided.
	ErrEmptyAPIToken = errors.New("empty Posthog API token")
)

type (
	// EventReporter is a PostHog-backed EventReporter.
	EventReporter struct {
		tracer         tracing.Tracer
		logger         logging.Logger
		client         posthog.Client
		circuitBreaker circuitbreaking.CircuitBreaker
	}
)

// NewPostHogEventReporter returns a new PostHog-backed EventReporter.
func NewPostHogEventReporter(logger logging.Logger, tracerProvider tracing.TracerProvider, apiKey string, circuitBreaker circuitbreaking.CircuitBreaker, configModifiers ...func(*posthog.Config)) (analytics.EventReporter, error) {
	if apiKey == "" {
		return nil, ErrEmptyAPIToken
	}

	phc := posthog.Config{Endpoint: "https://app.posthog.com"}
	for _, f := range configModifiers {
		f(&phc)
	}

	client, err := posthog.NewWithConfig(apiKey, phc)
	if err != nil {
		return nil, err
	}

	c := &EventReporter{
		tracer:         tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(name)),
		logger:         logging.EnsureLogger(logger).WithName(name),
		client:         client,
		circuitBreaker: circuitBreaker,
	}

	return c, nil
}

// Close wraps the internal client's Close method.
func (c *EventReporter) Close() {
	if err := c.client.Close(); err != nil {
		c.logger.Error("closing connection", err)
	}
}

// AddUser upsert's a user's identity.
func (c *EventReporter) AddUser(ctx context.Context, userID string, properties map[string]any) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if c.circuitBreaker.CannotProceed() {
		return internalerrors.ErrCircuitBroken
	}

	props := posthog.NewProperties()
	for k, v := range properties {
		props.Set(k, v)
	}

	err := c.client.Enqueue(posthog.Identify{
		DistinctId: userID,
		Properties: props,
	})
	if err != nil {
		c.circuitBreaker.Failed()
		return err
	}

	c.circuitBreaker.Succeeded()
	return nil
}

// EventOccurred associates events with a user.
func (c *EventReporter) EventOccurred(ctx context.Context, event, userID string, properties map[string]any) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if c.circuitBreaker.CannotProceed() {
		return internalerrors.ErrCircuitBroken
	}

	props := posthog.NewProperties()
	for k, v := range properties {
		props.Set(k, v)
	}

	err := c.client.Enqueue(posthog.Capture{
		DistinctId: userID,
		Event:      event,
		Properties: props,
	})
	if err != nil {
		c.circuitBreaker.Failed()
		return err
	}

	c.circuitBreaker.Succeeded()
	return nil
}
