package featureflagscfg

import (
	"fmt"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/featureflags"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/google/wire"
)

var (
	ProvidersFeatureFlags = wire.NewSet(
		ProvideFeatureFlagManager,
	)
)

func ProvideFeatureFlagManager(c *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider, httpClient *http.Client) (featureflags.FeatureFlagManager, error) {
	circuitBreaker, err := c.CircuitBreaker.ProvideCircuitBreaker(logger, metricsProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize feature flag circuit breaker: %w", err)
	}

	return c.ProvideFeatureFlagManager(logger, tracerProvider, httpClient, circuitBreaker)
}
