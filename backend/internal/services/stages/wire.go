package stages

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	pusherpkg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/pusher"

	"github.com/google/wire"
)

var (
	StagesProviders = wire.NewSet(
		ProvideHandler,
	)
)

// ProvideHandler provides a stages Handler.
func ProvideHandler(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	stagesRepo stagesdomain.Repository,
	playScriptsRepo playscripts.Repository,
	pusher *pusherpkg.Client,
) *Handler {
	return NewHandler(logger, tracerProvider, stagesRepo, playScriptsRepo, pusher)
}
