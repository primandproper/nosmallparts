package main

import (
	"context"
	"log"

	apiserver "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/build/services/api"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"

	_ "go.uber.org/automaxprocs"
)

func main() {
	rootCtx := context.Background()

	cfg, err := config.LoadConfigFromEnvironment[config.APIServiceConfig]()
	if err != nil {
		log.Fatal(err)
	}

	// only allow initialization to take so long.
	buildCtx, cancel := context.WithTimeout(rootCtx, cfg.HTTPServer.StartupDeadline)

	logger, err := cfg.Observability.Logging.ProvideLogger(rootCtx)
	if err != nil {
		log.Fatalf("could not create logger: %v", err)
	}

	// build our server struct.
	server, err := apiserver.NewServer(buildCtx, logger, cfg)
	if err != nil {
		log.Fatal(err)
	}

	cancel()

	// Run server (handles signals internally for graceful shutdown)
	server.Run()
}
