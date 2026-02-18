package chi

import (
	"net/http"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// buildLoggingMiddleware builds a logging middleware.
func buildLoggingMiddleware(logger logging.Logger, tracer tracing.Tracer, silenceRouteLogging bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx, span := tracer.StartSpan(req.Context())
			defer span.End()

			ww := chimiddleware.NewWrapResponseWriter(res, req.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, req.WithContext(ctx))

			if !silenceRouteLogging {
				logger.WithRequest(req).WithSpan(span).WithValues(map[string]any{
					"status":  ww.Status(),
					"elapsed": time.Since(start).Milliseconds(),
					"written": ww.BytesWritten(),
				}).Info("response served")
			}
		})
	}
}
