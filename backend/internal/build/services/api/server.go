package api

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapi "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/build/services/api/http"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/http"
)

type Server struct {
	logger     logging.Logger
	httpServer http.Server
}

func NewServer(ctx context.Context, logger logging.Logger, cfg *config.APIServiceConfig) (*Server, error) {
	httpServer, err := httpapi.Build(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create http server: %w", err)
	}

	return &Server{
		logger:     logging.EnsureLogger(logger),
		httpServer: httpServer,
	}, nil
}

func (s *Server) Run() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	go s.httpServer.Serve()

	sig := <-signalChan
	s.logger.WithValue("signal", sig.String()).Info("received shutdown signal")

	cancelCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	s.logger.Info("shutting down")

	if err := s.httpServer.Shutdown(cancelCtx); err != nil {
		s.logger.Error("shutting down HTTP server", err)
	}
}
