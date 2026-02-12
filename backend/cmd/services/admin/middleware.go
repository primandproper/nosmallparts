package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/localdev"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/pkg/client"
)

func (s *AdminFrontendServer) authMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		cookieName := s.config.Cookies.CookieName
		cookie, err := req.Cookie(cookieName)
		if err != nil {
			logger.Error("no cookie found", err)
			http.Redirect(res, req, "/login", http.StatusFound)
			return
		} else if cookie == nil {
			logger.Debug("no cookie found")
			http.Redirect(res, req, "/login", http.StatusFound)
			return
		}

		var payload authPayload
		if err = s.cookieManager.Decode(ctx, cookieName, cookie.Value, &payload); err != nil {
			logger.Error("decoding cookie", err)
			http.Redirect(res, req, "/login", http.StatusFound)
			return
		}

		c, err := localdev.BuildInsecureOAuthedGRPCClient(
			ctx,
			s.config.APIServiceConnection.OAuth2APIClientID,
			s.config.APIServiceConnection.OAuth2APIClientSecret,
			s.config.APIServiceConnection.HTTPAPIServerURL,
			s.config.APIServiceConnection.GRPCAPIServerURL,
			payload.AccessToken,
		)
		if err != nil {
			logger.Error("building client", err)
			http.Redirect(res, req, "/login", http.StatusFound)
			return
		}

		handler.ServeHTTP(res, req.WithContext(context.WithValue(ctx, apiClientContextKey, c)))
	})
}

type authPayload struct {
	AccessToken string
}

func fetchClientFromContext(ctx context.Context) (client.Client, error) {
	c, ok := ctx.Value(apiClientContextKey).(client.Client)
	if !ok {
		return nil, errors.New("no api client found in context")
	}

	return c, nil
}
