package main

import (
	"context"
	"encoding/base64"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config/envvars"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing/chi"
)

const (
	defaultAdminServerConfigurationFilepath = "deploy/environments/localdev/config_files/admin_webapp_config.json"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx := context.Background()

	// We don't yet have a way to write these values into the AdminWebappConfig (because they're not present in the root APIConfig struct).
	// This approach is an atrocious hack that I have to employ because I wasn't smart enough to design a better config generation system.
	must(os.Setenv(envvars.CookiesCookieNameEnvVarKey, "dev_admin_frontend"))
	must(os.Setenv(envvars.CookiesHashKeyEnvVarKey, base64.RawURLEncoding.EncodeToString([]byte("HEREISA32CHARSECRETWHICHISMADEUP"))))
	must(os.Setenv(envvars.CookiesBlockKeyEnvVarKey, base64.RawURLEncoding.EncodeToString([]byte("HEREISA32CHARSECRETWHICHISMADEUP"))))
	must(os.Setenv(envvars.APIServiceHTTPAPIServerURLEnvVarKey, "http://localhost:8000"))
	must(os.Setenv(envvars.APIServiceOauth2APIClientIDEnvVarKey, strings.Repeat("A", oauth.ClientIDSize)))
	must(os.Setenv(envvars.APIServiceOauth2APIClientSecretEnvVarKey, strings.Repeat("A", oauth.ClientSecretSize)))
	must(os.Setenv(envvars.ServerPortEnvVarKey, "8888"))
	must(os.Setenv(envvars.ServerStartupDeadlineEnvVarKey, time.Minute.String()))
	must(os.Setenv(envvars.CookiesDomainEnvVarKey, "localhost"))
	must(os.Setenv(envvars.CookiesLifetimeEnvVarKey, (7 * 24 * time.Hour).String()))

	configFilepath := os.Getenv("CONFIGURATION_FILEPATH")
	if configFilepath == "" {
		configFilepath = defaultAdminServerConfigurationFilepath
	}

	cfg, err := config.LoadConfigFromPath[config.AdminWebappConfig](ctx, configFilepath)
	if err != nil {
		log.Fatal(err)
	}

	logger, tracerProvider, _, err := cfg.Observability.ProvideThreePillars(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err = cfg.ValidateWithContext(ctx); err != nil {
		log.Fatal(err)
	}

	fs, err := NewAdminFrontendServer(
		ctx,
		logger,
		tracerProvider,
		encoding.ProvideServerEncoderDecoder(logger, tracerProvider, encoding.ContentTypeJSON),
		chi.NewRouteParamManager(),
		cfg,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("serving now")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	// Run server
	go fs.server.Serve()

	// os.Interrupt
	<-signalChan

	go func() {
		// os.Kill
		<-signalChan
	}()

	cancelCtx, cancelShutdown := context.WithTimeout(ctx, 10*time.Second)
	defer cancelShutdown()

	// Gracefully shutdown the server by waiting on existing requests (except websockets).
	if err = fs.server.Shutdown(cancelCtx); err != nil {
		log.Println("shutting down server", err)
	}
}
