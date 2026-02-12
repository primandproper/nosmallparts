package cloudtrace

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	o11yutils "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/utils"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider is a trace provider that supports ForceFlush. Returned by SetupCloudTrace
// and satisfied by *sdktrace.TracerProvider. Callers may use it as tracing.TracerProvider.
type TracerProvider interface {
	trace.TracerProvider
	ForceFlush(context.Context) error
}

type errorHandler struct {
	logger logging.Logger
}

func (h errorHandler) Handle(err error) {
	h.logger.Error("tracer reported issue", err)
}

func init() {
	otel.SetErrorHandler(errorHandler{logger: logging.NewNoopLogger().WithName("otel_errors")})
}

// SetupCloudTrace creates a new trace provider instance and registers it as global trace provider.
func SetupCloudTrace(ctx context.Context, serviceName string, spanCollectionProbability float64, cfg *Config) (TracerProvider, error) {
	exporter, err := texporter.New(texporter.WithProjectID(cfg.ProjectID))
	if err != nil {
		return nil, fmt.Errorf("setting up trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(o11yutils.MustOtelResource(ctx, serviceName)),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(spanCollectionProbability)),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}
