package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	api "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/build/services/api/http"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"

	_ "go.uber.org/automaxprocs"
)

func main() {
	rootCtx := context.Background()

	cfg, err := config.LoadConfigFromEnvironment[config.APIServiceConfig]()
	if err != nil {
		log.Fatal(err)
	}

	if strings.TrimSpace(os.Getenv("RUNNING_IN_GCP_CLOUD_RUN")) != "" {
		if portStr := strings.TrimSpace(os.Getenv("PORT")); portStr != "" {
			port, parseErr := strconv.ParseUint(portStr, 10, 16)
			if parseErr != nil {
				log.Printf("RUNNING_IN_GCP_CLOUD_RUN set but invalid PORT %q: %v; using config port %d", portStr, parseErr, cfg.HTTPServer.Port)
			} else {
				cfg.HTTPServer.Port = uint16(port)
			}
		}
	}

	// only allow initialization to take so long.
	buildCtx, cancel := context.WithTimeout(rootCtx, cfg.HTTPServer.StartupDeadline)

	logger, err := cfg.Observability.Logging.ProvideLogger(rootCtx)
	if err != nil {
		log.Fatalf("could not create logger: %v", err)
	}

	// build our server struct.
	srv, err := api.Build(buildCtx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	// Run server
	go srv.Serve()

	// os.Interrupt
	<-signalChan

	go func() {
		// os.Kill
		<-signalChan
	}()

	cancelCtx, cancelShutdown := context.WithTimeout(rootCtx, 10*time.Second)
	defer cancelShutdown()

	logger.Info("shutting down")

	// Gracefully shutdown the server by waiting on existing requests (except websockets).
	if err = srv.Shutdown(cancelCtx); err != nil {
		logger.Error("shutting down server", err)
	}
}
