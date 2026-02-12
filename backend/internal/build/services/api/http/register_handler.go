package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	identitymanager "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/manager"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

// RegisterRequest is the JSON body for POST /api/register.
type RegisterRequest struct {
	Username              string  `json:"username"`
	Password              string  `json:"password"`
	EmailAddress          string  `json:"emailAddress"`
	FirstName             string  `json:"firstName"`
	LastName              string  `json:"lastName"`
	Birthday              *string `json:"birthday,omitempty"`
	AccountName           string  `json:"accountName"`
	InvitationToken       string  `json:"invitationToken,omitempty"`
	InvitationID          string  `json:"invitationId,omitempty"`
	AcceptedTOS           bool    `json:"acceptedTos"`
	AcceptedPrivacyPolicy bool    `json:"acceptedPrivacyPolicy"`
}

// RegisterResponse is the JSON response for POST /api/register.
type RegisterResponse struct {
	CreatedUserID   string `json:"createdUserId"`
	TwoFactorQRCode string `json:"twoFactorQrCode"`
	TwoFactorSecret string `json:"twoFactorSecret"`
}

// RegisterHandler handles POST /api/register (user registration).
type RegisterHandler struct {
	identityManager identitymanager.IdentityDataManager
	logger          logging.Logger
	tracer          tracing.Tracer
}

// NewRegisterHandler returns a new RegisterHandler.
func NewRegisterHandler(
	identityManager identitymanager.IdentityDataManager,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) *RegisterHandler {
	return &RegisterHandler{
		identityManager: identityManager,
		logger:          logger,
		tracer:          tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer("register_handler")),
	}
}

// ServeHTTP decodes the request body, calls identity manager CreateUser, and returns JSON.
func (h *RegisterHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, span := h.tracer.StartSpan(req.Context())
	defer span.End()

	var body RegisterRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		h.logger.WithValue("error", err).Debug("register request body decode failed")
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte(`{"error":"invalid request body"}`))
		return
	}

	var birthday *time.Time
	if body.Birthday != nil && *body.Birthday != "" {
		t, err := time.Parse("2006-01-02", *body.Birthday)
		if err != nil {
			t, err = time.Parse(time.RFC3339, *body.Birthday)
		}
		if err == nil {
			birthday = &t
		}
	}

	accountName := body.AccountName
	if accountName == "" {
		accountName = body.FirstName + "'s Account"
	}

	input := &identity.UserRegistrationInput{
		Username:              body.Username,
		Password:              body.Password,
		EmailAddress:          body.EmailAddress,
		FirstName:             body.FirstName,
		LastName:              body.LastName,
		Birthday:              birthday,
		AccountName:           accountName,
		InvitationToken:       body.InvitationToken,
		InvitationID:          body.InvitationID,
		AcceptedTOS:           body.AcceptedTOS,
		AcceptedPrivacyPolicy: body.AcceptedPrivacyPolicy,
	}

	created, err := h.identityManager.CreateUser(ctx, input)
	if err != nil {
		h.logger.WithValue("error", err).Debug("create user failed")
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(`{"error":"registration failed"}`))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(res).Encode(RegisterResponse{
		CreatedUserID:   created.CreatedUserID,
		TwoFactorQRCode: created.TwoFactorQRCode,
		TwoFactorSecret: created.TwoFactorSecret,
	})
}
