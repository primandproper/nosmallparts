package vectorcfg

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/vector"

	"github.com/google/wire"
)

var (
	VectorProviders = wire.NewSet(
		ProvideVectorSearcher,
	)
)

func ProvideVectorSearcher(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider) (vector.Searcher, error) {
	return cfg.ProvideVectorSearcher(logger, tracerProvider)
}
