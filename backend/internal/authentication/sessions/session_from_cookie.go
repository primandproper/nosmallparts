package sessions

import (
	"context"
	"net/http"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing"
)

const (
	// DefaultSessionHeaderName is the default header the frontend sends the backend cookie value in.
	DefaultSessionHeaderName = "X-Session-Cookie"
	tokenPrefix              = "Bearer "
)

// SessionFromCookieConfig configures where the session value is read from.
type SessionFromCookieConfig struct {
	// HeaderName is the request header that may contain the backend cookie value (e.g. X-Session-Cookie). If empty, DefaultSessionHeaderName is used.
	HeaderName string
	// CookieName is the name used when encoding/decoding the session (must match cookies.Config.CookieName). If set, the middleware also checks the Cookie header for this name.
	CookieName string
}

// SessionFromCookieMiddleware returns middleware that reads the session from the request (header or cookie), decodes it with the cookie manager, and attaches ContextData to the request context.
// On missing value or decode error, the next handler is called without session (no 401).
func SessionFromCookieMiddleware(
	cookieManager cookies.Manager,
	sessionCookieName string,
	cfg SessionFromCookieConfig,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) routing.Middleware {
	headerName := cfg.HeaderName
	if headerName == "" {
		headerName = DefaultSessionHeaderName
	}
	tracer := tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer("session_from_cookie"))
	log := logging.EnsureLogger(logger).WithName("session_from_cookie")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ctx, span := tracer.StartSpan(req.Context())
			defer span.End()

			value := readSessionValue(req, headerName, cfg.CookieName)
			if value == "" {
				next.ServeHTTP(res, req.WithContext(ctx))
				return
			}

			var sessionData ContextData
			if err := cookieManager.Decode(ctx, sessionCookieName, value, &sessionData); err != nil {
				log.WithSpan(span).Debug("session cookie decode failed, continuing without session")
				next.ServeHTTP(res, req.WithContext(ctx))
				return
			}

			ctx = context.WithValue(ctx, SessionContextDataKey, &sessionData)
			next.ServeHTTP(res, req.WithContext(ctx))
		})
	}
}

// readSessionValue returns the session value from the header first, then from the cookie if configured.
func readSessionValue(req *http.Request, headerName, cookieName string) string {
	if v := req.Header.Get(headerName); v != "" {
		return strings.TrimPrefix(v, tokenPrefix)
	}
	if cookieName != "" {
		if c, err := req.Cookie(cookieName); err == nil && c.Value != "" {
			return c.Value
		}
	}
	return ""
}
