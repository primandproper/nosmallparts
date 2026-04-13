package config

import (
	"github.com/verygoodsoftwarenotvirus/platform/v4/server/http"

	"github.com/google/wire"
)

var (
	// ServiceConfigProviders represents this package's offering to the dependency injector.
	ServiceConfigProviders = wire.NewSet(
		wire.FieldsOf(
			new(*APIServiceConfig),
			"Auth",
			"Observability",
			"Email",
			"Analytics",
			"FeatureFlags",
			"Encoding",
			"Routing",
			"Database",
			"Meta",
			"Events",
			"Queues",
			"TextSearch",
			"Services",
		),
		wire.FieldsOf(
			new(*AdminWebappConfig),
			"Cookies",
		),
		wire.FieldsOf(
			new(*ServicesConfig),
			"Users",
			"DataPrivacy",
			"OAuth2Clients",
			"OAuth2Clients",
			"UploadedMedia",
		),
		ProvideHTTPServerConfigFromAPIServiceConfig,
		ProvideHTTPServerConfigFromAdminWebappConfig,
	)
)

func ProvideHTTPServerConfigFromAPIServiceConfig(cfg *APIServiceConfig) http.Config {
	return cfg.HTTPServer
}

func ProvideHTTPServerConfigFromAdminWebappConfig(cfg *AdminWebappConfig) http.Config {
	return cfg.HTTPServer
}
