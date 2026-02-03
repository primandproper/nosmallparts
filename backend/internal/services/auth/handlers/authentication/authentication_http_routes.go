package authentication

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/types"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	servertiming "github.com/mitchellh/go-server-timing"
)

func (s *service) postLogin(ctx context.Context, user *identity.User, defaultAccountID string) (int, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	dcm := &audit.DataChangeMessage{
		EventType: identity.UserLoggedInServiceEventType,
		AccountID: defaultAccountID,
		UserID:    user.ID,
	}

	if err := s.dataChangesPublisher.Publish(ctx, dcm); err != nil {
		observability.AcknowledgeError(err, s.logger, span, "publishing data change message")
		return http.StatusAccepted, nil
	}

	if err := s.analyticsReporter.AddUser(ctx, user.ID, map[string]any{
		"username":        user.Username,
		"default_account": defaultAccountID,
		"first_name":      user.FirstName,
		"last_name":       user.LastName,
	}); err != nil {
		return http.StatusAccepted, observability.PrepareError(err, span, "identifying user for analytics")
	}

	if err := s.featureFlagManager.Identify(ctx, user); err != nil {
		return http.StatusAccepted, observability.PrepareError(err, span, "identifying user in feature flag manager")
	}

	return http.StatusAccepted, nil
}

func (s *service) SSOLoginHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req).WithSpan(span)
	tracing.AttachRequestToSpan(span, req)

	responseDetails := types.ResponseDetails{
		TraceID: span.SpanContext().TraceID().String(),
	}

	providerName := s.authProviderFetcher(req)
	if providerName == "" {
		errRes := types.NewAPIErrorResponse("provider name is missing", types.ErrValidatingRequestInput, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusBadRequest)
		return
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "getting provider")

		errorResponse := &types.APIResponse[any]{
			Details: types.ResponseDetails{
				TraceID: span.SpanContext().TraceID().String(),
			},
			Error: &types.APIError{
				Message: "provider is not supported",
			},
		}
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errorResponse, http.StatusBadRequest)
		return
	}

	sess, err := provider.BeginAuth(gothic.SetState(req))
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "beginning auth")

		errorResponse := &types.APIResponse[any]{
			Details: types.ResponseDetails{
				TraceID: span.SpanContext().TraceID().String(),
			},
			Error: &types.APIError{
				Message: "failed to begin auth",
			},
		}
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errorResponse, http.StatusInternalServerError)
		return
	}

	u, err := sess.GetAuthURL()
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "getting auth url")

		errorResponse := &types.APIResponse[any]{
			Details: types.ResponseDetails{
				TraceID: span.SpanContext().TraceID().String(),
			},
			Error: &types.APIError{
				Message: "failed to get auth url",
			},
		}
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errorResponse, http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-type", "application/json")
	http.Redirect(res, req, u, http.StatusTemporaryRedirect)
}

func (s *service) SSOLoginCallbackHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	timing := servertiming.FromContext(ctx)
	logger := s.logger.WithRequest(req).WithSpan(span)
	tracing.AttachRequestToSpan(span, req)

	responseDetails := types.ResponseDetails{
		TraceID: span.SpanContext().TraceID().String(),
	}

	providerName := s.authProviderFetcher(req)
	if providerName == "" {
		errRes := types.NewAPIErrorResponse("provider name is missing", types.ErrValidatingRequestInput, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusBadRequest)
		return
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "getting provider")

		errorResponse := &types.APIResponse[any]{
			Details: types.ResponseDetails{
				TraceID: span.SpanContext().TraceID().String(),
			},
			Error: &types.APIError{
				Message: "provider is not supported",
			},
		}
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errorResponse, http.StatusBadRequest)
		return
	}

	// NOTE: I know this doesn't work, but I can't be bothered to fix it right now
	value := req.Header.Get("Authorization")

	sess, err := provider.UnmarshalSession(value)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "unmarshaling session")

		errorResponse := &types.APIResponse[any]{
			Details: types.ResponseDetails{
				TraceID: span.SpanContext().TraceID().String(),
			},
			Error: &types.APIError{
				Message: "failed to unmarshal session",
			},
		}
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errorResponse, http.StatusInternalServerError)
		return
	}

	if err = validateState(req, sess); err != nil {
		observability.AcknowledgeError(err, logger, span, "validating state")

		errorResponse := &types.APIResponse[any]{
			Details: types.ResponseDetails{
				TraceID: span.SpanContext().TraceID().String(),
			},
			Error: &types.APIError{
				Message: "failed to validate state",
			},
		}
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errorResponse, http.StatusInternalServerError)
		return
	}

	fetchUserTimer := timing.NewMetric("database").WithDesc("fetch").Start()
	providedUser, err := provider.FetchUser(sess)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching user")

		errorResponse := &types.APIResponse[any]{
			Details: types.ResponseDetails{
				TraceID: span.SpanContext().TraceID().String(),
			},
			Error: &types.APIError{
				Message: "failed to fetch user",
			},
		}
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errorResponse, http.StatusInternalServerError)
		return
	}
	fetchUserTimer.Stop()

	params := req.URL.Query()
	if params.Encode() == "" && req.Method == http.MethodPost {
		if err = req.ParseForm(); err != nil {
			observability.AcknowledgeError(err, logger, span, "parsing form")
		}
		params = req.Form
	}

	// get new token and retry fetch
	if _, err = sess.Authorize(provider, params); err != nil {
		observability.AcknowledgeError(err, logger, span, "authorizing session")

		errorResponse := &types.APIResponse[any]{
			Details: types.ResponseDetails{
				TraceID: span.SpanContext().TraceID().String(),
			},
			Error: &types.APIError{
				Message: "failed to authorize session",
			},
		}
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errorResponse, http.StatusInternalServerError)
		return
	}

	getUserTimer := timing.NewMetric("database").WithDesc("get user by email").Start()
	user, err := s.userDataManager.GetUserByEmail(ctx, providedUser.Email)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "getting user by email")
		errRes := types.NewAPIErrorResponse("database error", types.ErrTalkingToDatabase, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusInternalServerError)
		return
	}
	getUserTimer.Stop()

	defaultAccountTimer := timing.NewMetric("database").WithDesc("get default account for user").Start()
	defaultAccountID, err := s.accountMembershipManager.GetDefaultAccountIDForUser(ctx, user.ID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching user memberships")
		errRes := types.NewAPIErrorResponse("database error", types.ErrTalkingToDatabase, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusInternalServerError)
		return
	}
	defaultAccountTimer.Stop()

	var token string
	token, err = s.tokenIssuer.IssueToken(ctx, user, s.config.TokenLifetime)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "signing token")
		errRes := types.NewAPIErrorResponse(staticError, types.ErrEncryptionIssue, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusInternalServerError)
		return
	}

	responseValue := &types.APIResponse[*identity.TokenResponse]{
		Details: responseDetails,
		Data: &identity.TokenResponse{
			AccountID:   defaultAccountID,
			UserID:      user.ID,
			AccessToken: token,
		},
	}

	responseCode, err := s.postLogin(ctx, user, defaultAccountID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "handling login status")
		errorResponse := &types.APIResponse[any]{
			Details: types.ResponseDetails{
				TraceID: span.SpanContext().TraceID().String(),
			},
			Error: &types.APIError{
				Message: staticError,
			},
		}

		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errorResponse, responseCode)
		return
	}

	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, responseValue, responseCode)
	logger.Debug("user logged in via SSO")
}

// validateState ensures that the state token param from the original.
func validateState(req *http.Request, sess goth.Session) error {
	rawAuthURL, err := sess.GetAuthURL()
	if err != nil {
		return err
	}

	authURL, err := url.Parse(rawAuthURL)
	if err != nil {
		return err
	}

	reqState := gothic.GetState(req)

	originalState := authURL.Query().Get("state")
	if originalState != "" && (originalState != reqState) {
		return errors.New("state token mismatch")
	}
	return nil
}

var _ oauth.OAuth2Service = (*service)(nil)

// AuthorizeHandler is our oauth2 auth route.
func (s *service) AuthorizeHandler(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	if err := s.oauth2Server.HandleAuthorizeRequest(res, req); err != nil {
		observability.AcknowledgeError(err, logger, span, "handling authorization request")
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
}

func (s *service) TokenHandler(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	if err := s.oauth2Server.HandleTokenRequest(res, req); err != nil {
		observability.AcknowledgeError(err, logger, span, "handling token request")
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
