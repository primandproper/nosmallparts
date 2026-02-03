package analyticscfg

import (
	"context"
	"fmt"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/posthog"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/rudderstack"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/segment"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ProviderSegment represents Segment.
	ProviderSegment = "segment"
	// ProviderRudderstack represents Rudderstack.
	ProviderRudderstack = "rudderstack"
	// ProviderPostHog represents PostHog.
	ProviderPostHog = "posthog"
)

type (
	// Config is the configuration structure.
	Config struct {
		Segment        *segment.Config        `env:"init"     envPrefix:"SEGMENT_"         json:"segment"`
		Posthog        *posthog.Config        `env:"init"     envPrefix:"POSTHOG_"         json:"posthog"`
		Rudderstack    *rudderstack.Config    `env:"init"     envPrefix:"RUDDERSTACK_"     json:"rudderstack"`
		Provider       string                 `env:"PROVIDER" json:"provider"`
		CircuitBreaker circuitbreaking.Config `env:"init"     envPrefix:"CIRCUIT_BREAKER_" json:"circuitBreaker"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Provider, validation.In(ProviderSegment, ProviderRudderstack, ProviderPostHog)),
		validation.Field(&cfg.Segment, validation.When(cfg.Provider == ProviderSegment, validation.Required), validation.When(cfg.Provider != ProviderSegment, validation.Nil)),
		validation.Field(&cfg.Posthog, validation.When(cfg.Provider == ProviderPostHog, validation.Required), validation.When(cfg.Provider != ProviderPostHog, validation.Nil)),
		validation.Field(&cfg.Rudderstack, validation.When(cfg.Provider == ProviderRudderstack, validation.Required), validation.When(cfg.Provider != ProviderRudderstack, validation.Nil)),
	)
}

// ProvideCollector provides a collector.
func (cfg *Config) ProvideCollector(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
) (analytics.EventReporter, error) {
	cb, err := cfg.CircuitBreaker.ProvideCircuitBreaker(logger, metricsProvider)
	if err != nil {
		return nil, fmt.Errorf("could not create analytics circuit breaker: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case ProviderSegment:
		return segment.NewSegmentEventReporter(logger, tracerProvider, cfg.Segment.APIToken, cb)
	case ProviderRudderstack:
		return rudderstack.NewRudderstackEventReporter(logger, tracerProvider, cfg.Rudderstack, cb)
	case ProviderPostHog:
		return posthog.NewPostHogEventReporter(logger, tracerProvider, cfg.Posthog.APIKey, cb)
	default:
		return analytics.NewNoopEventReporter(), nil
	}
}
