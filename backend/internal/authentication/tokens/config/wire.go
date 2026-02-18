package tokenscfg

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/google/wire"
)

var (
	// TokenIssuerProviders are what we provide to dependency injection.
	TokenIssuerProviders = wire.NewSet(
		ProvideTokenIssuer,
	)
)

// ProvideTokenIssuer provides a tokens.Issuer from a config.
func ProvideTokenIssuer(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider) (tokens.Issuer, error) {
	return cfg.ProvideTokenIssuer(logger, tracerProvider)
}
