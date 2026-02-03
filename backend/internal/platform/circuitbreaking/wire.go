package circuitbreaking

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"

	"github.com/google/wire"
)

var (
	Providers = wire.NewSet(
		ProvideCircuitBreaker,
	)
)

func ProvideCircuitBreaker(cfg *Config, logger logging.Logger, metricsProvider metrics.Provider) (CircuitBreaker, error) {
	return cfg.ProvideCircuitBreaker(logger, metricsProvider)
}
