package stages

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	platformws "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/websocket"

	"github.com/google/wire"
)

var (
	StagesWebSocketProviders = wire.NewSet(
		ProvideConnectionManager,
		ProvideHandler,
	)
)

// ProvideConnectionManager provides a ConnectionManager.
func ProvideConnectionManager(
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
) *platformws.ConnectionManager {
	return platformws.NewConnectionManager(tracerProvider, logger)
}

// ProvideHandler provides a WebSocket Handler.
func ProvideHandler(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	stagesRepo stagesdomain.Repository,
	playScriptsRepo playscripts.Repository,
	connectionManager *platformws.ConnectionManager,
) *Handler {
	return NewHandler(logger, tracerProvider, stagesRepo, playScriptsRepo, connectionManager)
}
