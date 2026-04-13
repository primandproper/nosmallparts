package pusher

import (
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

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
