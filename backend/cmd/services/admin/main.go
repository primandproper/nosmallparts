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

// setEnvIfUnset sets the env var only when unset, so K8s-injected values (e.g. from secretKeyRef) are preserved.
func setEnvIfUnset(key, value string) {
	if os.Getenv(key) == "" {
		must(os.Setenv(key, value))
	}
}

func main() {
	ctx := context.Background()

	// Local dev defaults; only set when unset so K8s/CI-injected values (from secretKeyRef) are preserved.
	setEnvIfUnset(envvars.CookiesCookieNameEnvVarKey, "dev_admin_frontend")
	setEnvIfUnset(envvars.CookiesHashKeyEnvVarKey, base64.RawURLEncoding.EncodeToString([]byte("HEREISA32CHARSECRETWHICHISMADEUP")))
	setEnvIfUnset(envvars.CookiesBlockKeyEnvVarKey, base64.RawURLEncoding.EncodeToString([]byte("HEREISA32CHARSECRETWHICHISMADEUP")))
	setEnvIfUnset(envvars.APIServiceHTTPAPIServerURLEnvVarKey, "http://localhost:8000")
	setEnvIfUnset(envvars.APIServiceGrpcAPIServerURLEnvVarKey, ":8001")
	setEnvIfUnset(envvars.APIServiceOauth2APIClientIDEnvVarKey, strings.Repeat("A", oauth.ClientIDSize))
	setEnvIfUnset(envvars.APIServiceOauth2APIClientSecretEnvVarKey, strings.Repeat("A", oauth.ClientSecretSize))
	setEnvIfUnset(envvars.ServerPortEnvVarKey, "8888")
	setEnvIfUnset(envvars.ServerStartupDeadlineEnvVarKey, time.Minute.String())
	setEnvIfUnset(envvars.CookiesDomainEnvVarKey, "localhost")
	setEnvIfUnset(envvars.CookiesLifetimeEnvVarKey, (7 * 24 * time.Hour).String())

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
