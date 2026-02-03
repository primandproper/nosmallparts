package analyticscfg

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/google/wire"
)

var (
	// Providers are what we provide to dependency injection.
	Providers = wire.NewSet(
		ProvideEventReporter,
	)
)

// ProvideEventReporter provides an analytics.EventReporter from a config.
func ProvideEventReporter(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider) (analytics.EventReporter, error) {
	return cfg.ProvideCollector(logger, tracerProvider, metricsProvider)
}
