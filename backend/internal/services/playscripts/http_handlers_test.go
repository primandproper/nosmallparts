package playscripts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	mockencoding "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// fakePlayScriptsManager implements PlayScriptsManager for tests.
type fakePlayScriptsManager struct {
	CreatePlayScriptFunc         func(ctx context.Context, input *playscripts.PlayScriptCreationRequestInput, accountID, userID string) (*playscripts.PlayScript, error)
	GetPlayScriptFunc            func(ctx context.Context, playScriptID, accountID string) (*playscripts.PlayScript, error)
	GetPlayScriptsForAccountFunc func(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error)
	GetPublicPlayScriptsFunc     func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error)
	UpdatePlayScriptFunc         func(ctx context.Context, playScriptID string, input *playscripts.PlayScriptUpdateRequestInput, accountID string) (*playscripts.PlayScript, error)
	ArchivePlayScriptFunc        func(ctx context.Context, playScriptID, accountID string) error
}

func (f *fakePlayScriptsManager) CreatePlayScript(ctx context.Context, input *playscripts.PlayScriptCreationRequestInput, accountID, userID string) (*playscripts.PlayScript, error) {
	if f.CreatePlayScriptFunc != nil {
		return f.CreatePlayScriptFunc(ctx, input, accountID, userID)
	}
	return nil, errors.New("not implemented")
}

func (f *fakePlayScriptsManager) GetPlayScript(ctx context.Context, playScriptID, accountID string) (*playscripts.PlayScript, error) {
	if f.GetPlayScriptFunc != nil {
		return f.GetPlayScriptFunc(ctx, playScriptID, accountID)
	}
	return nil, errors.New("not implemented")
}

func (f *fakePlayScriptsManager) GetPlayScriptsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error) {
	if f.GetPlayScriptsForAccountFunc != nil {
		return f.GetPlayScriptsForAccountFunc(ctx, accountID, filter)
	}
	return nil, errors.New("not implemented")
}

func (f *fakePlayScriptsManager) GetPublicPlayScripts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error) {
	if f.GetPublicPlayScriptsFunc != nil {
		return f.GetPublicPlayScriptsFunc(ctx, filter)
	}
	return nil, errors.New("not implemented")
}

func (f *fakePlayScriptsManager) UpdatePlayScript(ctx context.Context, playScriptID string, input *playscripts.PlayScriptUpdateRequestInput, accountID string) (*playscripts.PlayScript, error) {
	if f.UpdatePlayScriptFunc != nil {
		return f.UpdatePlayScriptFunc(ctx, playScriptID, input, accountID)
	}
	return nil, errors.New("not implemented")
}

func (f *fakePlayScriptsManager) ArchivePlayScript(ctx context.Context, playScriptID, accountID string) error {
	if f.ArchivePlayScriptFunc != nil {
		return f.ArchivePlayScriptFunc(ctx, playScriptID, accountID)
	}
	return errors.New("not implemented")
}

var _ playscripts.PlayScriptsManager = (*fakePlayScriptsManager)(nil)

func sessionContext(accountID, userID string) *sessions.ContextData {
	return &sessions.ContextData{
		ActiveAccountID: accountID,
		Requester:       sessions.RequesterInfo{UserID: userID},
	}
}

func requestWithSession(req *http.Request, session *sessions.ContextData) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), sessions.SessionContextDataKey, session))
}

func requestWithChiParam(req *http.Request, key, value string) *http.Request {
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
}

func buildTestHTTPHandler(manager playscripts.PlayScriptsManager) *HTTPHandler {
	enc := mockencoding.NewMockEncoderDecoder()
	return NewHTTPHandler(manager, enc, logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
}

func TestHTTPHandler_CreatePlayScript(t *testing.T) {
	t.Parallel()

	t.Run("unauthenticated_returns_401", func(t *testing.T) {
		t.Parallel()

		h := buildTestHTTPHandler(&fakePlayScriptsManager{})
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, mock.AnythingOfType("*playscripts.errorResponse"), http.StatusUnauthorized).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/play-scripts", bytes.NewReader([]byte(`{"title":"Test"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.CreatePlayScript(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid_body_returns_400", func(t *testing.T) {
		t.Parallel()

		h := buildTestHTTPHandler(&fakePlayScriptsManager{})
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("DecodeRequest", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("decode error")).Once()
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, mock.AnythingOfType("*playscripts.errorResponse"), http.StatusBadRequest).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/play-scripts", bytes.NewReader([]byte(`{`)))
		req.Header.Set("Content-Type", "application/json")
		req = requestWithSession(req, sessionContext("acc-1", "user-1"))
		rec := httptest.NewRecorder()

		h.CreatePlayScript(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success_returns_201", func(t *testing.T) {
		t.Parallel()

		created := &playscripts.PlayScript{ID: "ps-1", Title: "Test Play"}
		mgr := &fakePlayScriptsManager{
			CreatePlayScriptFunc: func(ctx context.Context, input *playscripts.PlayScriptCreationRequestInput, accountID, userID string) (*playscripts.PlayScript, error) {
				assert.Equal(t, "acc-1", accountID)
				assert.Equal(t, "user-1", userID)
				assert.Equal(t, "Test Play", input.Title)
				return created, nil
			},
		}
		h := buildTestHTTPHandler(mgr)
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("DecodeRequest", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			dest := args.Get(2).(*playscripts.PlayScriptCreationRequestInput)
			*dest = playscripts.PlayScriptCreationRequestInput{Title: "Test Play"}
		}).Return(nil).Once()
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, created, http.StatusCreated).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/play-scripts", bytes.NewReader([]byte(`{"title":"Test Play"}`)))
		req.Header.Set("Content-Type", "application/json")
		req = requestWithSession(req, sessionContext("acc-1", "user-1"))
		rec := httptest.NewRecorder()

		h.CreatePlayScript(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})
}

func TestHTTPHandler_GetPlayScript(t *testing.T) {
	t.Parallel()

	t.Run("missing_id_returns_400", func(t *testing.T) {
		t.Parallel()

		h := buildTestHTTPHandler(&fakePlayScriptsManager{})
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, mock.AnythingOfType("*playscripts.errorResponse"), http.StatusBadRequest).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/play-scripts/", http.NoBody)
		rec := httptest.NewRecorder()

		h.GetPlayScript(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("permission_denied_returns_403", func(t *testing.T) {
		t.Parallel()

		mgr := &fakePlayScriptsManager{
			GetPlayScriptFunc: func(ctx context.Context, playScriptID, accountID string) (*playscripts.PlayScript, error) {
				return nil, ErrPermissionDenied
			},
		}
		h := buildTestHTTPHandler(mgr)
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, mock.AnythingOfType("*playscripts.errorResponse"), http.StatusForbidden).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/play-scripts/ps-1", http.NoBody)
		req = requestWithChiParam(req, "id", "ps-1")
		rec := httptest.NewRecorder()

		h.GetPlayScript(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("success_returns_200", func(t *testing.T) {
		t.Parallel()

		play := &playscripts.PlayScript{ID: "ps-1", Title: "Test"}
		mgr := &fakePlayScriptsManager{
			GetPlayScriptFunc: func(ctx context.Context, playScriptID, accountID string) (*playscripts.PlayScript, error) {
				assert.Equal(t, "ps-1", playScriptID)
				return play, nil
			},
		}
		h := buildTestHTTPHandler(mgr)
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, play, http.StatusOK).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/play-scripts/ps-1", http.NoBody)
		req = requestWithChiParam(req, "id", "ps-1")
		rec := httptest.NewRecorder()

		h.GetPlayScript(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestHTTPHandler_GetPlayScriptsForAccount(t *testing.T) {
	t.Parallel()

	t.Run("unauthenticated_returns_401", func(t *testing.T) {
		t.Parallel()

		h := buildTestHTTPHandler(&fakePlayScriptsManager{})
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, mock.AnythingOfType("*playscripts.errorResponse"), http.StatusUnauthorized).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/play-scripts/account", http.NoBody)
		rec := httptest.NewRecorder()

		h.GetPlayScriptsForAccount(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("success_returns_200", func(t *testing.T) {
		t.Parallel()

		result := filtering.NewQueryFilteredResult(
			[]*playscripts.PlayScript{},
			0, 0,
			func(p *playscripts.PlayScript) string { return p.ID },
			filtering.DefaultQueryFilter(),
		)
		mgr := &fakePlayScriptsManager{
			GetPlayScriptsForAccountFunc: func(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error) {
				assert.Equal(t, "acc-1", accountID)
				return result, nil
			},
		}
		h := buildTestHTTPHandler(mgr)
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, result, http.StatusOK).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/play-scripts/account", http.NoBody)
		req = requestWithSession(req, sessionContext("acc-1", "user-1"))
		rec := httptest.NewRecorder()

		h.GetPlayScriptsForAccount(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestHTTPHandler_GetPublicPlayScripts(t *testing.T) {
	t.Parallel()

	result := filtering.NewQueryFilteredResult(
		[]*playscripts.PlayScript{},
		0, 0,
		func(p *playscripts.PlayScript) string { return p.ID },
		filtering.DefaultQueryFilter(),
	)
	mgr := &fakePlayScriptsManager{
		GetPublicPlayScriptsFunc: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error) {
			return result, nil
		},
	}
	h := buildTestHTTPHandler(mgr)
	h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, result, http.StatusOK).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/play-scripts/public", http.NoBody)
	rec := httptest.NewRecorder()

	h.GetPublicPlayScripts(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHTTPHandler_UpdatePlayScript(t *testing.T) {
	t.Parallel()

	t.Run("unauthenticated_returns_401", func(t *testing.T) {
		t.Parallel()

		h := buildTestHTTPHandler(&fakePlayScriptsManager{})
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, mock.AnythingOfType("*playscripts.errorResponse"), http.StatusUnauthorized).Once()

		body, _ := json.Marshal(playscripts.PlayScriptUpdateRequestInput{})
		req := httptest.NewRequest(http.MethodPatch, "/api/play-scripts/ps-1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = requestWithChiParam(req, "id", "ps-1")
		rec := httptest.NewRecorder()

		h.UpdatePlayScript(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("permission_denied_returns_403", func(t *testing.T) {
		t.Parallel()

		mgr := &fakePlayScriptsManager{
			UpdatePlayScriptFunc: func(ctx context.Context, playScriptID string, input *playscripts.PlayScriptUpdateRequestInput, accountID string) (*playscripts.PlayScript, error) {
				return nil, ErrPermissionDenied
			},
		}
		h := buildTestHTTPHandler(mgr)
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("DecodeRequest", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, mock.AnythingOfType("*playscripts.errorResponse"), http.StatusForbidden).Once()

		req := httptest.NewRequest(http.MethodPatch, "/api/play-scripts/ps-1", bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		req = requestWithSession(req, sessionContext("acc-1", "user-1"))
		req = requestWithChiParam(req, "id", "ps-1")
		rec := httptest.NewRecorder()

		h.UpdatePlayScript(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestHTTPHandler_ArchivePlayScript(t *testing.T) {
	t.Parallel()

	t.Run("unauthenticated_returns_401", func(t *testing.T) {
		t.Parallel()

		h := buildTestHTTPHandler(&fakePlayScriptsManager{})
		h.encoderDecoder.(*mockencoding.EncoderDecoder).On("EncodeResponseWithStatus", mock.Anything, mock.Anything, mock.AnythingOfType("*playscripts.errorResponse"), http.StatusUnauthorized).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/play-scripts/ps-1", http.NoBody)
		req = requestWithChiParam(req, "id", "ps-1")
		rec := httptest.NewRecorder()

		h.ArchivePlayScript(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("success_returns_204", func(t *testing.T) {
		t.Parallel()

		mgr := &fakePlayScriptsManager{
			ArchivePlayScriptFunc: func(ctx context.Context, playScriptID, accountID string) error {
				assert.Equal(t, "ps-1", playScriptID)
				assert.Equal(t, "acc-1", accountID)
				return nil
			},
		}
		h := buildTestHTTPHandler(mgr)

		req := httptest.NewRequest(http.MethodDelete, "/api/play-scripts/ps-1", http.NoBody)
		req = requestWithSession(req, sessionContext("acc-1", "user-1"))
		req = requestWithChiParam(req, "id", "ps-1")
		rec := httptest.NewRecorder()

		h.ArchivePlayScript(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}
