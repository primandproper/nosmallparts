package authentication

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens"
	tokenscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	name = "authentication_manager"
)

type (
	Manager interface {
		ProcessLogin(ctx context.Context, adminOnly bool, loginData *auth.UserLoginInput) (*auth.TokenResponse, error)
		ExchangeTokenForUser(ctx context.Context, refreshToken string) (*auth.TokenResponse, error)
	}

	manager struct {
		tokenIssuer             tokens.Issuer
		authenticator           Authenticator
		tracer                  tracing.Tracer
		logger                  logging.Logger
		dataChangesPublisher    messagequeue.Publisher
		userAuthDataManager     identity.Repository
		maxAccessTokenLifetime  time.Duration
		maxRefreshTokenLifetime time.Duration
	}
)

func NewManager(
	ctx context.Context,
	queuesConfig *msgconfig.QueuesConfig,
	tokenIssuer tokens.Issuer,
	authenticator Authenticator,
	tracingProvider tracing.TracerProvider,
	logger logging.Logger,
	publisherProvider messagequeue.PublisherProvider,
	userAuthDataManager identity.Repository,
	cfg *tokenscfg.Config,
) (Manager, error) {
	dataChangesPublisher, err := publisherProvider.ProvidePublisher(ctx, queuesConfig.DataChangesTopicName)
	if err != nil {
		return nil, observability.PrepareError(err, nil, "creating data changes publisher")
	}

	m := &manager{
		maxRefreshTokenLifetime: cfg.MaxRefreshTokenLifetime,
		maxAccessTokenLifetime:  cfg.MaxAccessTokenLifetime,
		tracer:                  tracing.NewTracer(tracing.EnsureTracerProvider(tracingProvider).Tracer(name)),
		logger:                  logging.EnsureLogger(logger).WithName(name),
		tokenIssuer:             tokenIssuer,
		authenticator:           authenticator,
		dataChangesPublisher:    dataChangesPublisher,
		userAuthDataManager:     userAuthDataManager,
	}

	return m, nil
}

// validateLogin takes login information and returns whether the login is valid.
// In the event that there's an error, this function will return false and the error.
func (m *manager) validateLogin(ctx context.Context, user *identity.User, loginInput *auth.UserLoginInput) (bool, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	loginInput.TOTPToken = strings.TrimSpace(loginInput.TOTPToken)
	loginInput.Password = strings.TrimSpace(loginInput.Password)
	loginInput.Username = strings.TrimSpace(loginInput.Username)

	// alias the relevant data.
	logger := m.logger.WithValue(keys.UsernameKey, user.Username)

	// check for login validity.
	loginValid, err := m.authenticator.CredentialsAreValid(
		ctx,
		user.HashedPassword,
		loginInput.Password,
		user.TwoFactorSecret,
		loginInput.TOTPToken,
	)

	if errors.Is(err, ErrInvalidTOTPToken) || errors.Is(err, ErrPasswordDoesNotMatch) {
		return false, err
	}

	if err != nil {
		return false, observability.PrepareError(err, span, "validating login")
	}

	logger.Debug("login validated")

	return loginValid, nil
}

func (m *manager) ProcessLogin(ctx context.Context, adminOnly bool, loginData *auth.UserLoginInput) (*auth.TokenResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.Clone()

	if err := validation.ValidateStructWithContext(ctx, loginData,
		validation.Field(&loginData.Username, validation.Required, validation.Length(4, math.MaxInt8)),
		validation.Field(&loginData.Password, validation.Required, validation.Length(8, math.MaxInt8)),
		// non-admin users can have unverified 2FA secrets
		validation.Field(&loginData.TOTPToken, is.Digit, validation.RuneLength(6, 6), validation.When(adminOnly)),
	); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	logger = logger.WithValue(keys.UsernameKey, loginData.Username)

	userFunc := m.userAuthDataManager.GetUserByUsername
	if adminOnly {
		userFunc = m.userAuthDataManager.GetAdminUserByUsername
	}

	user, err := userFunc(ctx, loginData.Username)
	if err != nil || user == nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, observability.PrepareError(err, span, "user does not exist")
		}

		return nil, observability.PrepareError(err, span, "fetching user")
	}

	logger = logger.WithValue(keys.UserIDKey, user.ID)
	tracing.AttachToSpan(span, keys.UserIDKey, user.ID)

	if user.IsBanned() {
		return nil, observability.PrepareError(err, span, "user is banned")
	}

	loginValid, err := m.validateLogin(ctx, user, loginData)
	logger.WithValue("login_valid", loginValid)

	if err != nil {
		if errors.Is(err, ErrInvalidTOTPToken) {
			return nil, observability.PrepareError(err, span, "invalid TOTP AccessToken")
		}

		if errors.Is(err, ErrPasswordDoesNotMatch) {
			return nil, observability.PrepareError(err, span, "password did not match")
		}

		return nil, observability.PrepareError(err, span, "validating login")
	} else if !loginValid {
		return nil, observability.PrepareError(err, span, "login was invalid")
	}

	if user.TwoFactorSecretVerifiedAt != nil && loginData.TOTPToken == "" {
		return nil, observability.PrepareError(errors.New("TOTP code required but not provided"), span, "processing login")
	}

	defaultAccountID, err := m.userAuthDataManager.GetDefaultAccountIDForUser(ctx, user.ID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	dcm := &audit.DataChangeMessage{
		EventType: identity.UserLoggedInServiceEventType,
		AccountID: defaultAccountID,
		UserID:    user.ID,
	}

	if err = m.dataChangesPublisher.Publish(ctx, dcm); err != nil {
		return nil, observability.PrepareError(err, span, "publishing data change")
	}

	response := &auth.TokenResponse{
		UserID:     user.ID,
		AccountID:  defaultAccountID,
		ExpiresUTC: time.Now().Add(m.maxAccessTokenLifetime).UTC(),
	}

	response.AccessToken, err = m.tokenIssuer.IssueToken(ctx, user, m.maxAccessTokenLifetime)
	if err != nil {
		return nil, observability.PrepareError(err, span, "creating access token")
	}

	response.RefreshToken, err = m.tokenIssuer.IssueToken(ctx, user, m.maxRefreshTokenLifetime)
	if err != nil {
		return nil, observability.PrepareError(err, span, "creating refresh token")
	}

	return response, nil
}

func (m *manager) ExchangeTokenForUser(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.Clone()

	userID, err := m.tokenIssuer.ParseUserIDFromToken(ctx, refreshToken)
	if err != nil {
		return nil, observability.PrepareError(err, span, "parsing userID from token")
	}

	user, err := m.userAuthDataManager.GetUser(ctx, userID)
	if err != nil || user == nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, observability.PrepareError(err, span, "user does not exist")
		}

		return nil, observability.PrepareError(err, span, "fetching user")
	}

	logger = logger.WithValue(keys.UserIDKey, user.ID)
	tracing.AttachToSpan(span, keys.UserIDKey, user.ID)

	if user.IsBanned() {
		return nil, observability.PrepareError(err, span, "user is banned")
	}

	defaultAccountID, err := m.userAuthDataManager.GetDefaultAccountIDForUser(ctx, user.ID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	dcm := &audit.DataChangeMessage{
		EventType: identity.UserLoggedInServiceEventType,
		AccountID: defaultAccountID,
		UserID:    user.ID,
	}

	if err = m.dataChangesPublisher.Publish(ctx, dcm); err != nil {
		return nil, observability.PrepareError(err, span, "publishing data change")
	}

	response := &auth.TokenResponse{
		UserID:     user.ID,
		AccountID:  defaultAccountID,
		ExpiresUTC: time.Now().Add(m.maxAccessTokenLifetime).UTC(),
	}

	response.AccessToken, err = m.tokenIssuer.IssueToken(ctx, user, m.maxAccessTokenLifetime)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating accessToken")
	}

	response.RefreshToken, err = m.tokenIssuer.IssueToken(ctx, user, m.maxRefreshTokenLifetime)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating accessToken")
	}

	return response, nil
}
