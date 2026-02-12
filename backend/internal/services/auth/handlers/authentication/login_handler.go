package authentication

import (
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/types"
)

// loginResponse is the JSON body returned on successful login.
type loginResponse struct {
	UserID    string `json:"userID"`
	AccountID string `json:"accountID"`
}

// LoginHandler handles POST /api/login: accepts UserLoginInput, validates credentials, and sets a session cookie.
func (s *service) LoginHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req).WithSpan(span)
	tracing.AttachRequestToSpan(span, req)

	responseDetails := types.ResponseDetails{
		TraceID: span.SpanContext().TraceID().String(),
	}

	if req.Method != http.MethodPost {
		errRes := types.NewAPIErrorResponse("method not allowed", types.ErrValidatingRequestInput, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusMethodNotAllowed)
		return
	}

	var input auth.UserLoginInput
	if err := s.encoderDecoder.DecodeRequest(ctx, req, &input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding login request")
		errRes := types.NewAPIErrorResponse("invalid request body", types.ErrDecodingRequestInput, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusBadRequest)
		return
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		errRes := types.NewAPIErrorResponse(err.Error(), types.ErrValidatingRequestInput, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusBadRequest)
		return
	}

	tokenResponse, err := s.authManager.ProcessLogin(ctx, false, &input)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "processing login")
		errRes := types.NewAPIErrorResponse("invalid credentials", types.ErrUserIsNotAuthorized, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusUnauthorized)
		return
	}

	sessionCtxData, err := s.accountMembershipManager.BuildSessionContextDataForUser(ctx, tokenResponse.UserID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "building session context")
		errRes := types.NewAPIErrorResponse(staticError, types.ErrFetchingSessionContextData, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusInternalServerError)
		return
	}

	encoded, err := s.cookieManager.Encode(ctx, s.sessionCookieName, sessionCtxData)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding session cookie")
		errRes := types.NewAPIErrorResponse(staticError, types.ErrEncryptionIssue, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     s.sessionCookieName,
		Value:    encoded,
		Path:     "/",
		MaxAge:   int(s.cookieConfig.Lifetime.Seconds()),
		HttpOnly: true,
		Secure:   s.cookieConfig.SecureOnly,
		SameSite: http.SameSiteLaxMode,
	}
	if s.cookieConfig.Domain != "" {
		cookie.Domain = s.cookieConfig.Domain
	}
	http.SetCookie(res, cookie)

	responseValue := &types.APIResponse[loginResponse]{
		Details: responseDetails,
		Data: loginResponse{
			UserID:    tokenResponse.UserID,
			AccountID: tokenResponse.AccountID,
		},
	}
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, responseValue, http.StatusOK)
}
