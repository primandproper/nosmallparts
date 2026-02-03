package grpc

import (
	"context"
	"errors"

	authentication2 "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth/managers"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

const (
	o11yName = "auth_service"
)

var _ authsvc.AuthServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		authsvc.UnimplementedAuthServiceServer
		tracer                tracing.Tracer
		logger                logging.Logger
		identityRepository    identity.Repository
		authenticationManager authentication2.Manager
		authManager           managers.AuthManagerInterface
	}
)

func NewAuthService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	identityRepository identity.Repository,
	// bruh what the actual fuck are we even doing here
	authManager managers.AuthManagerInterface,
	authenticationManager authentication2.Manager,
) authsvc.AuthServiceServer {
	return &serviceImpl{
		logger:                logging.EnsureLogger(logger).WithName(o11yName),
		tracer:                tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		identityRepository:    identityRepository,
		authManager:           authManager,
		authenticationManager: authenticationManager,
	}
}

func (s *serviceImpl) fetchSessionContext(ctx context.Context) (*sessions.ContextData, error) {
	sessionContext, ok := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
	if !ok {
		return nil, errors.New("session context not found")
	}

	return sessionContext, nil
}
