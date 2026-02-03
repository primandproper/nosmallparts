package analyticscfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/segment"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/stretchr/testify/require"
)

func TestProvideCollector(T *testing.T) {
	T.Parallel()

	T.Run("noop", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		logger := logging.NewNoopLogger()

		actual, err := ProvideEventReporter(cfg, logger, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		require.NoError(t, err)
		require.NotNil(t, actual)
	})

	T.Run("with segment", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderSegment,
			Segment: &segment.Config{
				APIToken: t.Name(),
			},
		}
		logger := logging.NewNoopLogger()

		actual, err := ProvideEventReporter(cfg, logger, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		require.NoError(t, err)
		require.NotNil(t, actual)
	})
}
