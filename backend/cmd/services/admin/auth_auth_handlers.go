package main

import (
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/pkg/client"

	g "maragu.dev/gomponents"
)

func (s *AdminFrontendServer) LoginPage(_ http.ResponseWriter, _ *http.Request) (g.Node, error) {
	return page("Login",
		s.componentRenderer.LoginForm(&components.LoginFormProps{}),
	), nil
}

func (s *AdminFrontendServer) LoginSubmission(res http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	var loginInput *auth.UserLoginInput
	if err := s.encoder.DecodeRequest(ctx, req, &loginInput); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding request")
		return s.componentRenderer.LoginForm(&components.LoginFormProps{GeneralError: err.Error()}), nil
	}

	var (
		usernameError, passwordError, totpError string
	)

	if err := loginInput.ValidateAdminLoginWithContext(ctx); err != nil {
		usernameError = fetchErrorString(err, "username")
		passwordError = fetchErrorString(err, "password")
		totpError = fetchErrorString(err, "totpToken")
	}

	if usernameError != "" || passwordError != "" || totpError != "" {
		return s.componentRenderer.LoginForm(&components.LoginFormProps{
			UsernameError: usernameError,
			PasswordError: passwordError,
			TOTPError:     totpError,
		}), nil
	}

	unauthedClient, err := client.BuildUnauthenticatedGRPCClient(s.config.APIServiceConnection.GRPCAPIServerURL)
	if err != nil {
		return s.componentRenderer.LoginForm(&components.LoginFormProps{
			GeneralError: err.Error(),
		}), nil
	}

	tokenRes, err := unauthedClient.AdminLoginForToken(ctx, &authsvc.AdminLoginForTokenRequest{
		Input: &authsvc.UserLoginInput{
			Username:  loginInput.Username,
			Password:  loginInput.Password,
			TotpToken: loginInput.TOTPToken,
		},
	})
	if err != nil {
		return s.componentRenderer.LoginForm(&components.LoginFormProps{
			GeneralError: err.Error(),
		}), nil
	}

	// Exchange JWT for OAuth2 token so subsequent requests use the OAuth2 token (stored in oauth2_client_tokens).
	oauth2Token, err := client.ExchangeJWTForOAuth2Token(
		ctx,
		s.config.APIServiceConnection.HTTPAPIServerURL,
		s.config.APIServiceConnection.OAuth2APIClientID,
		s.config.APIServiceConnection.OAuth2APIClientSecret,
		tokenRes.Result.AccessToken,
	)
	if err != nil {
		return s.componentRenderer.LoginForm(&components.LoginFormProps{
			GeneralError: err.Error(),
		}), nil
	}

	encodedCookie, err := s.cookieManager.Encode(ctx, s.config.Cookies.CookieName, &authPayload{AccessToken: oauth2Token.AccessToken})
	if err != nil {
		return s.componentRenderer.LoginForm(&components.LoginFormProps{
			GeneralError: err.Error(),
		}), nil
	}

	http.SetCookie(res, s.buildCookie(ctx, encodedCookie))

	// Redirect to home page after successful login using HTMX redirect
	// This causes HTMX to do a full page redirect instead of swapping content
	res.Header().Set("HX-Redirect", "/")
	res.WriteHeader(http.StatusOK)

	return g.El("div"), nil
}
