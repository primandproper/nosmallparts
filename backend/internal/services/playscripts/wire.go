package playscripts

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/google/wire"
)

var (
	// HTTPProviders provides the PlayScriptsManager and HTTP handler for play scripts.
	HTTPProviders = wire.NewSet(
		ProvidePlayScriptsManager,
		ProvideHTTPHandler,
	)
)

// ProvidePlayScriptsManager provides a PlayScriptsManager that delegates to the given repository.
func ProvidePlayScriptsManager(repo playscripts.Repository) playscripts.PlayScriptsManager {
	return NewManager(repo)
}

// ProvideHTTPHandler provides an HTTP handler for play scripts.
func ProvideHTTPHandler(
	manager playscripts.PlayScriptsManager,
	encoderDecoder encoding.ServerEncoderDecoder,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) *HTTPHandler {
	return NewHTTPHandler(manager, encoderDecoder, logger, tracerProvider)
}
