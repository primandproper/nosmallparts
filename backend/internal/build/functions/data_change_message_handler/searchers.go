package datachangemessagehandler

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	textsearchcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/text/config"
	identityindexing "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/indexing"

	"github.com/google/wire"
)

var SearcherProviders = wire.NewSet(
	ProvideUserTextSearcher,
)

func ProvideUserTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (identityindexing.UserTextSearcher, error) {
	return textsearchcfg.ProvideIndex[identityindexing.UserSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		identityindexing.IndexTypeUsers,
	)
}
