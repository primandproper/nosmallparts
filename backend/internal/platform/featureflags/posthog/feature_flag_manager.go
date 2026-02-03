package posthog

import (
	"context"
	"errors"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/featureflags"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/internalerrors"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/posthog/posthog-go"
)

const (
	serviceName = "posthog_feature_flag_manager"
)

var (
	ErrNilConfig          = errors.New("missing config")
	ErrNilUser            = errors.New("missing user")
	ErrMissingCredentials = errors.New("missing PostHog credentials")
)

type (
	// featureFlagManager implements the feature flag interface.
	featureFlagManager struct {
		logger         logging.Logger
		tracer         tracing.Tracer
		posthogClient  posthog.Client
		circuitBreaker circuitbreaking.CircuitBreaker
	}
)

// NewFeatureFlagManager constructs a new featureFlagManager.
func NewFeatureFlagManager(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, circuitBreaker circuitbreaking.CircuitBreaker, configModifiers ...func(config *posthog.Config)) (featureflags.FeatureFlagManager, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	cfg.CircuitBreakerConfig.EnsureDefaults()

	if cfg.ProjectAPIKey == "" {
		return nil, fmt.Errorf("missing credential 'ProjectAPIKey': %w", ErrMissingCredentials)
	}

	if cfg.PersonalAPIKey == "" {
		return nil, fmt.Errorf("missing credential 'PersonalAPIKey': %w", ErrMissingCredentials)
	}

	phc := posthog.Config{
		PersonalApiKey: cfg.PersonalAPIKey,
	}

	for _, modifier := range configModifiers {
		modifier(&phc)
	}

	client, err := posthog.NewWithConfig(
		cfg.ProjectAPIKey,
		phc,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create posthog client: %w", err)
	}

	ffm := &featureFlagManager{
		posthogClient:  client,
		circuitBreaker: circuitBreaker,
		logger:         logging.EnsureLogger(logger).WithName(serviceName),
		tracer:         tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(serviceName)),
	}

	return ffm, nil
}

// CanUseFeature returns whether a user can use a feature or not.
func (f *featureFlagManager) CanUseFeature(ctx context.Context, userID, feature string) (bool, error) {
	_, span := f.tracer.StartSpan(ctx)
	defer span.End()

	logger := f.logger.WithValue(keys.UserIDKey, userID).WithValue("feature", feature)

	if !f.circuitBreaker.CanProceed() {
		return false, internalerrors.ErrCircuitBroken
	}

	flagEnabled, err := f.posthogClient.IsFeatureEnabled(posthog.FeatureFlagPayload{
		Key:        feature,
		DistinctId: userID,
	})
	if err != nil {
		f.circuitBreaker.Failed()
		return false, observability.PrepareAndLogError(err, logger, span, "checking feature flag eligibility")
	}

	if enabled, ok := flagEnabled.(bool); ok {
		f.circuitBreaker.Failed()
		return enabled, nil
	}

	return false, nil
}

// Identify identifies a user in PostHog.
func (f *featureFlagManager) Identify(ctx context.Context, user featureflags.User) error {
	_, span := f.tracer.StartSpan(ctx)
	defer span.End()

	if !f.circuitBreaker.CanProceed() {
		return internalerrors.ErrCircuitBroken
	}

	if user == nil {
		return ErrNilUser
	}

	logger := f.logger.WithValue(keys.UserIDKey, user.GetID())

	err := f.posthogClient.Enqueue(posthog.Identify{
		DistinctId: user.GetID(),
		Properties: map[string]any{
			"username":   user.GetUsername(),
			"first_name": user.GetFirstName(),
			"last_name":  user.GetLastName(),
		},
	})
	if err != nil {
		f.circuitBreaker.Failed()
		return observability.PrepareAndLogError(err, logger, span, "identifying user")
	}

	f.circuitBreaker.Succeeded()
	return nil
}

// Close closes the PostHog client.
func (f *featureFlagManager) Close() error {
	return f.posthogClient.Close()
}
