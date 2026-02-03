package observability

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	loggingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"
	metricscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	tracingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Config contains settings about how we report our metrics.
	Config struct {
		_ struct{} `json:"-"`

		Logging loggingcfg.Config `envPrefix:"LOGGING_" json:"logging"`
		Metrics metricscfg.Config `envPrefix:"METRICS_" json:"metrics"`
		Tracing tracingcfg.Config `envPrefix:"TRACING_" json:"tracing"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Logging),
		validation.Field(&cfg.Metrics),
		validation.Field(&cfg.Tracing),
	)
}

func (cfg *Config) ProvideThreePillars(ctx context.Context) (logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider, err error) {
	logger, err = cfg.Logging.ProvideLogger(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("setting up logger: %w", err)
	}

	tracerProvider, err = cfg.Tracing.ProvideTracerProvider(ctx, logger)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("setting up tracer provider: %w", err)
	}

	metricsProvider, err = cfg.Metrics.ProvideMetricsProvider(ctx, logger)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("setting up metrics provider: %w", err)
	}

	return logger, tracerProvider, metricsProvider, nil
}
