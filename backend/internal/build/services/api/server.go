package api

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcapi "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/build/services/api/grpc"
	httpapi "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/build/services/api/http"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/http"
)

type Server struct {
	logger     logging.Logger
	grpcServer *grpcapi.GRPCService
	httpServer http.Server
}

func NewServer(ctx context.Context, logger logging.Logger, cfg *config.APIServiceConfig) (*Server, error) {
	// build our server struct.
	httpServer, err := httpapi.Build(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create http server: %w", err)
	}

	grpcServer, err := grpcapi.Build(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create grpc server: %w", err)
	}

	return &Server{
		logger:     logging.EnsureLogger(logger),
		grpcServer: grpcServer,
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

	// Run servers
	go s.httpServer.Serve()
	go s.grpcServer.Serve()

	// Wait for shutdown signal
	sig := <-signalChan
	s.logger.WithValue("signal", sig.String()).Info("received shutdown signal")

	cancelCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	s.logger.Info("shutting down")

	if err := s.httpServer.Shutdown(cancelCtx); err != nil {
		s.logger.Error("shutting down HTTP server", err)
	}

	s.grpcServer.Shutdown()
}
