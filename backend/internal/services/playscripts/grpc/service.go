package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	playscripts "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

const (
	o11yName = "play_scripts_service"
)

var _ playscriptssvc.PlayScriptsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		playscriptssvc.UnimplementedPlayScriptsServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
		playScriptRepository      playscripts.Repository
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	playScriptRepository playscripts.Repository,
) playscriptssvc.PlayScriptsServiceServer {
	return &serviceImpl{
		logger:                    logging.EnsureLogger(logger).WithName(o11yName),
		tracer:                    tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		sessionContextDataFetcher: sessions.FetchContextDataFromContext,
		playScriptRepository:      playScriptRepository,
	}
}
