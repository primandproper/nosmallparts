package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/go-oauth2/oauth2/v4/manage"
)

const (
	authHeaderName = "Authorization"
	tokenPrefix    = "Bearer "
)

// sessionTokenLoader resolves an access token to a user ID (for testing we can fake this).
type sessionTokenLoader interface {
	UserIDFromAccessToken(ctx context.Context, access string) (string, error)
}

// oauth2ManagerForSession is a distinct type so wire can provide a separate OAuth2 manager for the session-issuing endpoint without conflicting with the auth service's manager.
type oauth2ManagerForSession struct{ *manage.Manager }

// UserIDFromAccessToken implements sessionTokenLoader using the OAuth2 manager.
func (o *oauth2ManagerForSession) UserIDFromAccessToken(ctx context.Context, access string) (string, error) {
	token, err := o.LoadAccessToken(ctx, access)
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", err
	}
	if userID := token.GetUserID(); userID != "" {
		return userID, nil
	}
	return "", err
}

// sessionCookieResponse is the JSON response for GET/POST /api/session.
type sessionCookieResponse struct {
	SessionCookie string `json:"sessionCookie"`
}

// SessionCookieHandler issues the backend session cookie value given a valid Bearer token.
// The frontend calls this once after OAuth2 token exchange, then stores and sends the cookie on each request.
type SessionCookieHandler struct {
	tokenLoader   sessionTokenLoader
	identityRepo  identity.Repository
	cookieManager cookies.Manager
	logger        logging.Logger
	tracer        tracing.Tracer
	sessionName   string
}

// NewSessionCookieHandler builds a SessionCookieHandler.
func NewSessionCookieHandler(
	tokenLoader sessionTokenLoader,
	identityRepo identity.Repository,
	cookieManager cookies.Manager,
	sessionCookieName string,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) *SessionCookieHandler {
	return &SessionCookieHandler{
		tokenLoader:   tokenLoader,
		identityRepo:  identityRepo,
		cookieManager: cookieManager,
		sessionName:   sessionCookieName,
		logger:        logging.EnsureLogger(logger).WithName("session_cookie_handler"),
		tracer:        tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer("session_cookie_handler")),
	}
}

// ServeHTTP handles GET and POST /api/session. Expects Authorization: Bearer <accessToken>.
// Returns 200 JSON {"sessionCookie":"<encoded value>"} or 401 if token is missing/invalid.
func (h *SessionCookieHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx, span := h.tracer.StartSpan(req.Context())
	defer span.End()

	if req.Method != http.MethodGet && req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	authHeader := req.Header.Get(authHeaderName)
	if authHeader == "" {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}
	accessToken := strings.TrimPrefix(authHeader, tokenPrefix)
	if accessToken == "" {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := h.tokenLoader.UserIDFromAccessToken(ctx, accessToken)
	if err != nil || userID == "" {
		h.logger.WithSpan(span).Debug("load access token failed")
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	sessionCtxData, err := h.identityRepo.BuildSessionContextDataForUser(ctx, userID)
	if err != nil {
		h.logger.WithSpan(span).Error("build session context data", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoded, err := h.cookieManager.Encode(ctx, h.sessionName, sessionCtxData)
	if err != nil {
		h.logger.WithSpan(span).Error("encode session cookie", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(res).Encode(sessionCookieResponse{SessionCookie: encoded}); err != nil {
		h.logger.WithSpan(span).Error("encode session response", err)
	}
}

var _ http.Handler = (*SessionCookieHandler)(nil)
