package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/settings"
	settingssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/settings"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

const (
	o11yName = "configuration_service"
)

var _ settingssvc.SettingsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		settingssvc.UnimplementedSettingsServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
		serviceSettingsRepository settings.Repository
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	settingsRepository settings.Repository,
) settingssvc.SettingsServiceServer {
	return &serviceImpl{
		logger:                    logging.EnsureLogger(logger).WithName(o11yName),
		tracer:                    tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		serviceSettingsRepository: settingsRepository,
		sessionContextDataFetcher: sessions.FetchContextDataFromContext,
	}
}
