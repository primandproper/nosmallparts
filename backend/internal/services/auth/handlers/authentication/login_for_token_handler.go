package authentication

import (
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/types"
)

// loginForTokenResponse is the JSON body returned by LoginForTokenHandler.
type loginForTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserID       string `json:"userID"`
	AccountID    string `json:"accountID"`
	Expires      string `json:"expires"`
}

// LoginForTokenHandler handles POST /api/login-for-token: accepts UserLoginInput, returns token JSON (for API clients).
func (s *service) LoginForTokenHandler(res http.ResponseWriter, req *http.Request) {
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
		observability.AcknowledgeError(err, logger, span, "decoding login-for-token request")
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
		observability.AcknowledgeError(err, logger, span, "processing login for token")
		errRes := types.NewAPIErrorResponse("invalid credentials", types.ErrUserIsNotAuthorized, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusUnauthorized)
		return
	}

	responseValue := &types.APIResponse[loginForTokenResponse]{
		Details: responseDetails,
		Data: loginForTokenResponse{
			AccessToken:  tokenResponse.AccessToken,
			RefreshToken: tokenResponse.RefreshToken,
			UserID:       tokenResponse.UserID,
			AccountID:    tokenResponse.AccountID,
			Expires:      tokenResponse.ExpiresUTC.Format("2006-01-02T15:04:05.000Z07:00"),
		},
	}
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, responseValue, http.StatusOK)
}
