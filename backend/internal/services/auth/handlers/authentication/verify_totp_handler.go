package authentication

import (
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/types"

	"github.com/pquerna/otp/totp"
)

// VerifyTOTPHandler handles POST /api/verify-totp: verifies a TOTP secret for a user.
func (s *service) VerifyTOTPHandler(res http.ResponseWriter, req *http.Request) {
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

	var input auth.TOTPSecretVerificationInput
	if err := s.encoderDecoder.DecodeRequest(ctx, req, &input); err != nil {
		observability.AcknowledgeError(err, logger, span, "decoding verify-totp request")
		errRes := types.NewAPIErrorResponse("invalid request body", types.ErrDecodingRequestInput, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusBadRequest)
		return
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		errRes := types.NewAPIErrorResponse(err.Error(), types.ErrValidatingRequestInput, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusBadRequest)
		return
	}

	user, err := s.userDataManager.GetUserWithUnverifiedTwoFactorSecret(ctx, input.UserID)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "fetching user for TOTP verification")
		errRes := types.NewAPIErrorResponse("user not found or already verified", types.ErrFetchingSessionContextData, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusBadRequest)
		return
	}

	if user.TwoFactorSecretVerifiedAt != nil {
		errRes := types.NewAPIErrorResponse("two factor secret already verified", types.ErrValidatingRequestInput, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusBadRequest)
		return
	}

	if !totp.Validate(input.TOTPToken, user.TwoFactorSecret) {
		errRes := types.NewAPIErrorResponse("invalid TOTP code", types.ErrUserIsNotAuthorized, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusUnauthorized)
		return
	}

	if err := s.userDataManager.MarkUserTwoFactorSecretAsVerified(ctx, user.ID); err != nil {
		observability.AcknowledgeError(err, logger, span, "marking TOTP secret verified")
		errRes := types.NewAPIErrorResponse("verification failed", types.ErrFetchingSessionContextData, responseDetails)
		s.encoderDecoder.EncodeResponseWithStatus(ctx, res, errRes, http.StatusInternalServerError)
		return
	}

	responseValue := &types.APIResponse[struct{ Accepted bool }]{
		Details: responseDetails,
		Data:    struct{ Accepted bool }{Accepted: true},
	}
	s.encoderDecoder.EncodeResponseWithStatus(ctx, res, responseValue, http.StatusOK)
}
