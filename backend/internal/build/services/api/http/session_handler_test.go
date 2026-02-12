package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	identitymock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeSessionTokenLoader struct {
	err    error
	userID string
}

func (f *fakeSessionTokenLoader) UserIDFromAccessToken(ctx context.Context, access string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.userID, nil
}

var _ sessionTokenLoader = (*fakeSessionTokenLoader)(nil)

type fakeCookieManager struct {
	encodeErr error
	encoded   string
}

func (f *fakeCookieManager) Encode(ctx context.Context, name string, value any) (string, error) {
	if f.encodeErr != nil {
		return "", f.encodeErr
	}
	return f.encoded, nil
}

func (f *fakeCookieManager) Decode(ctx context.Context, name, value string, dst any) error {
	return nil
}

var _ cookies.Manager = (*fakeCookieManager)(nil)

func TestSessionCookieHandler_ServeHTTP(T *testing.T) {
	T.Parallel()

	logger := logging.NewNoopLogger()
	tracerProvider := tracing.NewNoopTracerProvider()
	sessionName := "session"

	T.Run("missing_authorization_returns_401", func(t *testing.T) {
		t.Parallel()

		identityRepo := &identitymock.RepositoryMock{}
		h := NewSessionCookieHandler(
			&fakeSessionTokenLoader{userID: "user-1"},
			identityRepo,
			&fakeCookieManager{},
			sessionName,
			logger,
			tracerProvider,
		)
		req := httptest.NewRequest(http.MethodGet, "/api/session", http.NoBody)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	T.Run("empty_bearer_token_returns_401", func(t *testing.T) {
		t.Parallel()

		identityRepo := &identitymock.RepositoryMock{}
		h := NewSessionCookieHandler(
			&fakeSessionTokenLoader{userID: "user-1"},
			identityRepo,
			&fakeCookieManager{},
			sessionName,
			logger,
			tracerProvider,
		)
		req := httptest.NewRequest(http.MethodGet, "/api/session", http.NoBody)
		req.Header.Set("Authorization", "Bearer ")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	T.Run("wrong_method_returns_405", func(t *testing.T) {
		t.Parallel()

		identityRepo := &identitymock.RepositoryMock{}
		h := NewSessionCookieHandler(
			&fakeSessionTokenLoader{userID: "user-1"},
			identityRepo,
			&fakeCookieManager{},
			sessionName,
			logger,
			tracerProvider,
		)
		req := httptest.NewRequest(http.MethodPut, "/api/session", http.NoBody)
		req.Header.Set("Authorization", "Bearer token")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})

	T.Run("invalid_token_returns_401", func(t *testing.T) {
		t.Parallel()

		identityRepo := &identitymock.RepositoryMock{}
		h := NewSessionCookieHandler(
			&fakeSessionTokenLoader{err: assert.AnError},
			identityRepo,
			&fakeCookieManager{},
			sessionName,
			logger,
			tracerProvider,
		)
		req := httptest.NewRequest(http.MethodGet, "/api/session", http.NoBody)
		req.Header.Set("Authorization", "Bearer bad-token")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	T.Run("valid_token_returns_200_with_session_cookie", func(t *testing.T) {
		t.Parallel()

		encodedCookie := "encoded-session-cookie-value"
		identityRepo := &identitymock.RepositoryMock{}
		identityRepo.On("BuildSessionContextDataForUser", mock.Anything, "user-1").Return(&sessions.ContextData{}, nil).Once()
		h := NewSessionCookieHandler(
			&fakeSessionTokenLoader{userID: "user-1"},
			identityRepo,
			&fakeCookieManager{encoded: encodedCookie},
			sessionName,
			logger,
			tracerProvider,
		)
		req := httptest.NewRequest(http.MethodGet, "/api/session", http.NoBody)
		req.Header.Set("Authorization", "Bearer valid-token")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var body sessionCookieResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.Equal(t, encodedCookie, body.SessionCookie)
		mock.AssertExpectationsForObjects(t, identityRepo)
	})
}
