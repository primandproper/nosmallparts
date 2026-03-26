package pusher

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/google/wire"
)

var (
	Providers = wire.NewSet(
		ProvideClient,
	)
)

// ProvideClient provides a Pusher Client.
func ProvideClient(
	cfg *Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) *Client {
	return NewClient(cfg, logger, tracerProvider)
}
