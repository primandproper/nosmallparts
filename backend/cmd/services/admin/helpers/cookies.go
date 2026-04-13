package helpers

import (
	"context"
	"net/http"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
)

func BuildCookie(ctx context.Context, tracer tracing.Tracer, cfg *cookies.Config, value string) *http.Cookie {
	_, span := tracer.StartSpan(ctx)
	defer span.End()

	expiry := time.Now().Add(cfg.Lifetime)

	// https://www.calhoun.io/securing-cookies-in-go/
	cookie := &http.Cookie{
		Name:     cfg.CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.SecureOnly,
		// Domain:   cfg.Domain,
		Expires:  expiry,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Until(expiry).Seconds()),
	}

	return cookie
}

func BuildCookieManager(cfg *cookies.Config, tracerProvider tracing.TracerProvider) (cookies.Manager, error) {
	return cookies.NewCookieManager(cfg, tracerProvider)
}
