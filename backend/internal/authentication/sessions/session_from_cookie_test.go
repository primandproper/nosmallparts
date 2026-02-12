package sessions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCookieManager struct {
	encodeErr error
	decodeErr error
	decoded   *ContextData
}

func (f *fakeCookieManager) Encode(ctx context.Context, name string, value any) (string, error) {
	if f.encodeErr != nil {
		return "", f.encodeErr
	}
	return "encoded-value", nil
}

func (f *fakeCookieManager) Decode(ctx context.Context, name, value string, dst any) error {
	if f.decodeErr != nil {
		return f.decodeErr
	}
	if f.decoded != nil {
		*(dst.(*ContextData)) = *f.decoded
	}
	return nil
}

var _ cookies.Manager = (*fakeCookieManager)(nil)

func TestSessionFromCookieMiddleware(T *testing.T) {
	T.Parallel()

	logger := logging.NewNoopLogger()
	tracerProvider := tracing.NewNoopTracerProvider()
	sessionName := "session"

	T.Run("missing_value_calls_next_without_session", func(t *testing.T) {
		t.Parallel()

		fake := &fakeCookieManager{}
		mw := SessionFromCookieMiddleware(fake, sessionName, SessionFromCookieConfig{HeaderName: DefaultSessionHeaderName}, logger, tracerProvider)
		var captured *ContextData
		next := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			captured, _ = FetchContextDataFromRequest(req)
		})
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		mw(next).ServeHTTP(httptest.NewRecorder(), req)
		assert.Nil(t, captured)
	})

	T.Run("valid_header_attaches_session", func(t *testing.T) {
		t.Parallel()

		want := &ContextData{ActiveAccountID: "account-1"}
		want.Requester.UserID = "user-1"
		fake := &fakeCookieManager{decoded: want}
		mw := SessionFromCookieMiddleware(fake, sessionName, SessionFromCookieConfig{HeaderName: DefaultSessionHeaderName}, logger, tracerProvider)
		var captured *ContextData
		next := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			captured, _ = FetchContextDataFromRequest(req)
		})
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set(DefaultSessionHeaderName, "encoded-session-value")
		mw(next).ServeHTTP(httptest.NewRecorder(), req)
		require.NotNil(t, captured)
		assert.Equal(t, "user-1", captured.GetUserID())
		assert.Equal(t, "account-1", captured.GetActiveAccountID())
	})

	T.Run("decode_error_calls_next_without_session", func(t *testing.T) {
		t.Parallel()

		fake := &fakeCookieManager{decodeErr: assert.AnError}
		mw := SessionFromCookieMiddleware(fake, sessionName, SessionFromCookieConfig{HeaderName: DefaultSessionHeaderName}, logger, tracerProvider)
		var captured *ContextData
		next := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			captured, _ = FetchContextDataFromRequest(req)
		})
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set(DefaultSessionHeaderName, "bad-value")
		mw(next).ServeHTTP(httptest.NewRecorder(), req)
		assert.Nil(t, captured)
	})
}
