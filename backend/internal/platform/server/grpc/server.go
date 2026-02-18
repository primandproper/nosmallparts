package grpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/internalerrors"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName = "grpc_server"
)

type (
	Config struct {
		HTTPSCertificateFile  string `env:"TLS_CERTIFICATE_FILEPATH"     json:"tlsCertificate,omitempty"`
		TLSCertificateKeyFile string `env:"TLS_CERTIFICATE_KEY_FILEPATH" json:"tlsCertificateKey,omitempty"`
		Port                  uint16 `env:"PORT"                         json:"port"`
	}

	Server struct {
		logger     logging.Logger
		config     *Config
		grpcServer *grpc.Server
	}

	// RegistrationFunc is i.e. protobuf.RegisterSomeExampleServiceServer(grpcServer, &exampleServiceServerImpl{}).
	RegistrationFunc func(*grpc.Server)
)

func NewGRPCServer(
	cfg *Config,
	logger logging.Logger,
	unaryServerInterceptors []grpc.UnaryServerInterceptor,
	streamServerInterceptors []grpc.StreamServerInterceptor,
	registrationFunctions ...RegistrationFunc,
) (*Server, error) {
	if cfg == nil {
		return nil, internalerrors.NilConfigError("grpc server")
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(append([]grpc.UnaryServerInterceptor{LoggingInterceptor(logger)}, unaryServerInterceptors...)...),
		grpc.ChainStreamInterceptor(streamServerInterceptors...),
	}

	if cfg.TLSCertificateKeyFile != "" && cfg.HTTPSCertificateFile != "" {
		serverCert, err := tls.LoadX509KeyPair(cfg.HTTPSCertificateFile, cfg.TLSCertificateKeyFile)
		if err != nil {
			return nil, err
		}

		config := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientAuth:   tls.NoClientCert,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}

		opts = append(opts, grpc.Creds(credentials.NewTLS(config)))
	}

	grpcServer := grpc.NewServer(opts...)
	for _, rf := range registrationFunctions {
		rf(grpcServer)
	}

	reflection.Register(grpcServer)

	return &Server{
		logger:     logging.EnsureLogger(logger).WithName(serviceName),
		config:     cfg,
		grpcServer: grpcServer,
	}, nil
}

// Shutdown shuts down the server.
func (s *Server) Shutdown() {
	s.grpcServer.Stop()
}

// Serve serves GRPC traffic.
func (s *Server) Serve() {
	var lc net.ListenConfig
	lis, err := lc.Listen(context.Background(), "tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s.logger.WithValue("port", s.config.Port).Info("Listening for GRPC requests")
	if err = s.grpcServer.Serve(lis); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			// NOTE: there is a chance that next line won't have tim  e to run,
			// as main() doesn't wait for this goroutine to stop.
			os.Exit(0)
		}
	}
}

func LoggingInterceptor(logger logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		logger.WithValue("rpc.method", info.FullMethod).Info("rpc invoked")
		return handler(ctx, req)
	}
}
