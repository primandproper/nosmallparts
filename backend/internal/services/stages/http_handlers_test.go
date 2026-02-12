package stages

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	playscriptsmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts/mock"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_CreateStageHTTP(t *testing.T) {
	t.Parallel()

	t.Run("manager_error_returns_500", func(t *testing.T) {
		t.Parallel()

		mgr := NewManager(&fakeStagesRepo{
			CreateStageFunc: func(ctx context.Context, input *stagesdomain.StageCreationInput) (*stagesdomain.Stage, error) {
				return nil, errors.New("db error")
			},
		})
		h := buildTestHandler(t, mgr, &playscriptsmock.Repository{})

		req := httptest.NewRequest(http.MethodPost, "/api/stages", http.NoBody)
		rec := httptest.NewRecorder()

		h.CreateStageHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "error")
	})

	t.Run("success_returns_201_with_stage_id_and_code", func(t *testing.T) {
		t.Parallel()

		stageID := "stage-123"
		mgr := NewManager(&fakeStagesRepo{
			CreateStageFunc: func(ctx context.Context, input *stagesdomain.StageCreationInput) (*stagesdomain.Stage, error) {
				return &stagesdomain.Stage{ID: stageID, Code: input.Code}, nil
			},
		})
		h := buildTestHandler(t, mgr, &playscriptsmock.Repository{})

		req := httptest.NewRequest(http.MethodPost, "/api/stages", http.NoBody)
		rec := httptest.NewRecorder()

		h.CreateStageHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		var body CreateStageResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.Equal(t, stageID, body.StageID)
		assert.Len(t, body.Code, 6)
	})
}

func TestHandler_CheckStageExists(t *testing.T) {
	t.Parallel()

	reqWithCode := func(code string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/api/stages/"+code+"/exists", http.NoBody)
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("code", code)
		return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	}

	t.Run("missing_code_returns_400", func(t *testing.T) {
		t.Parallel()

		mgr := NewManager(&fakeStagesRepo{})
		h := buildTestHandler(t, mgr, &playscriptsmock.Repository{})

		req := httptest.NewRequest(http.MethodGet, "/api/stages//exists", http.NoBody)
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("code", "")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
		rec := httptest.NewRecorder()

		h.CheckStageExists(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "error")
	})

	t.Run("stage_not_found_returns_200_exists_false", func(t *testing.T) {
		t.Parallel()

		mgr := NewManager(&fakeStagesRepo{
			GetStageByCodeFunc: func(ctx context.Context, code string) (*stagesdomain.Stage, error) {
				return nil, sql.ErrNoRows
			},
		})
		h := buildTestHandler(t, mgr, &playscriptsmock.Repository{})

		req := reqWithCode("NONEXIST")
		rec := httptest.NewRecorder()

		h.CheckStageExists(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var body StageExistsResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.False(t, body.Exists)
	})

	t.Run("stage_found_returns_200_exists_true", func(t *testing.T) {
		t.Parallel()

		mgr := NewManager(&fakeStagesRepo{
			GetStageByCodeFunc: func(ctx context.Context, code string) (*stagesdomain.Stage, error) {
				assert.Equal(t, "ABC123", code)
				return &stagesdomain.Stage{ID: "s1", Code: code}, nil
			},
		})
		h := buildTestHandler(t, mgr, &playscriptsmock.Repository{})

		req := reqWithCode("ABC123")
		rec := httptest.NewRecorder()

		h.CheckStageExists(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var body StageExistsResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.True(t, body.Exists)
	})

	t.Run("manager_error_returns_500", func(t *testing.T) {
		t.Parallel()

		mgr := NewManager(&fakeStagesRepo{
			GetStageByCodeFunc: func(ctx context.Context, code string) (*stagesdomain.Stage, error) {
				return nil, errors.New("db error")
			},
		})
		h := buildTestHandler(t, mgr, &playscriptsmock.Repository{})

		req := reqWithCode("ABC123")
		rec := httptest.NewRecorder()

		h.CheckStageExists(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
