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
		ProvideStagesManager,
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

// ProvideStagesManager provides a StagesManager that delegates to the given repository.
func ProvideStagesManager(repo stagesdomain.Repository) stagesdomain.StagesManager {
	return NewManager(repo)
}

// ProvideHandler provides a WebSocket Handler.
func ProvideHandler(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	stagesManager stagesdomain.StagesManager,
	playScriptsRepo playscripts.Repository,
	connectionManager *platformws.ConnectionManager,
) *Handler {
	return NewHandler(logger, tracerProvider, stagesManager, playScriptsRepo, connectionManager)
}
