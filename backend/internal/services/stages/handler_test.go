package stages

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	playscriptsmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts/mock"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	platformws "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/websocket"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fakeStagesRepo is a stages repository for testing that delegates to optional funcs.
type fakeStagesRepo struct {
	CreateStageFunc                   func(ctx context.Context, input *stagesdomain.StageCreationInput) (*stagesdomain.Stage, error)
	GetStageByIDFunc                  func(ctx context.Context, stageID string) (*stagesdomain.Stage, error)
	GetStageByCodeFunc                func(ctx context.Context, code string) (*stagesdomain.Stage, error)
	UpdateStageFunc                   func(ctx context.Context, stageID string, input *stagesdomain.StageUpdateInput) error
	ArchiveStageFunc                  func(ctx context.Context, stageID string) error
	AddParticipantFunc                func(ctx context.Context, input *stagesdomain.ParticipantCreationInput) (*stagesdomain.StageParticipant, error)
	GetParticipantFunc                func(ctx context.Context, participantID string) (*stagesdomain.StageParticipant, error)
	GetActiveParticipantsForStageFunc func(ctx context.Context, stageID string) ([]*stagesdomain.StageParticipant, error)
	UpdateParticipantFunc             func(ctx context.Context, participantID string, input *stagesdomain.ParticipantUpdateInput) error
	SetParticipantLeftFunc            func(ctx context.Context, participantID string) error
	AddParticipantRoleFunc            func(ctx context.Context, participantID, castMemberID string) error
	RemoveParticipantRoleFunc         func(ctx context.Context, participantID, castMemberID string) error
	GetParticipantRolesFunc           func(ctx context.Context, participantID string) ([]string, error)
}

func (f *fakeStagesRepo) CreateStage(ctx context.Context, input *stagesdomain.StageCreationInput) (*stagesdomain.Stage, error) {
	if f.CreateStageFunc != nil {
		return f.CreateStageFunc(ctx, input)
	}
	return nil, sql.ErrNoRows
}

func (f *fakeStagesRepo) GetStageByID(ctx context.Context, stageID string) (*stagesdomain.Stage, error) {
	if f.GetStageByIDFunc != nil {
		return f.GetStageByIDFunc(ctx, stageID)
	}
	return nil, sql.ErrNoRows
}

func (f *fakeStagesRepo) GetStageByCode(ctx context.Context, code string) (*stagesdomain.Stage, error) {
	if f.GetStageByCodeFunc != nil {
		return f.GetStageByCodeFunc(ctx, code)
	}
	return nil, sql.ErrNoRows
}

func (f *fakeStagesRepo) UpdateStage(ctx context.Context, stageID string, input *stagesdomain.StageUpdateInput) error {
	if f.UpdateStageFunc != nil {
		return f.UpdateStageFunc(ctx, stageID, input)
	}
	return nil
}

func (f *fakeStagesRepo) ArchiveStage(ctx context.Context, stageID string) error {
	if f.ArchiveStageFunc != nil {
		return f.ArchiveStageFunc(ctx, stageID)
	}
	return nil
}

func (f *fakeStagesRepo) AddParticipant(ctx context.Context, input *stagesdomain.ParticipantCreationInput) (*stagesdomain.StageParticipant, error) {
	if f.AddParticipantFunc != nil {
		return f.AddParticipantFunc(ctx, input)
	}
	return nil, sql.ErrNoRows
}

func (f *fakeStagesRepo) GetParticipant(ctx context.Context, participantID string) (*stagesdomain.StageParticipant, error) {
	if f.GetParticipantFunc != nil {
		return f.GetParticipantFunc(ctx, participantID)
	}
	return nil, sql.ErrNoRows
}

func (f *fakeStagesRepo) GetActiveParticipantsForStage(ctx context.Context, stageID string) ([]*stagesdomain.StageParticipant, error) {
	if f.GetActiveParticipantsForStageFunc != nil {
		return f.GetActiveParticipantsForStageFunc(ctx, stageID)
	}
	return nil, nil
}

func (f *fakeStagesRepo) UpdateParticipant(ctx context.Context, participantID string, input *stagesdomain.ParticipantUpdateInput) error {
	if f.UpdateParticipantFunc != nil {
		return f.UpdateParticipantFunc(ctx, participantID, input)
	}
	return nil
}

func (f *fakeStagesRepo) SetParticipantLeft(ctx context.Context, participantID string) error {
	if f.SetParticipantLeftFunc != nil {
		return f.SetParticipantLeftFunc(ctx, participantID)
	}
	return nil
}

func (f *fakeStagesRepo) AddParticipantRole(ctx context.Context, participantID, castMemberID string) error {
	if f.AddParticipantRoleFunc != nil {
		return f.AddParticipantRoleFunc(ctx, participantID, castMemberID)
	}
	return nil
}

func (f *fakeStagesRepo) RemoveParticipantRole(ctx context.Context, participantID, castMemberID string) error {
	if f.RemoveParticipantRoleFunc != nil {
		return f.RemoveParticipantRoleFunc(ctx, participantID, castMemberID)
	}
	return nil
}

func (f *fakeStagesRepo) GetParticipantRoles(ctx context.Context, participantID string) ([]string, error) {
	if f.GetParticipantRolesFunc != nil {
		return f.GetParticipantRolesFunc(ctx, participantID)
	}
	return nil, nil
}

var _ stagesdomain.Repository = (*fakeStagesRepo)(nil)

func buildTestHandler(t *testing.T, stagesRepo stagesdomain.Repository, playScriptsRepo playscripts.Repository) *Handler {
	t.Helper()
	logger := logging.NewNoopLogger()
	tracerProvider := tracing.NewNoopTracerProvider()
	connections := platformws.NewConnectionManager(tracerProvider, logger)
	return NewHandler(logger, tracerProvider, stagesRepo, playScriptsRepo, connections)
}

func TestNewHandler(t *testing.T) {
	t.Parallel()

	h := buildTestHandler(t, &fakeStagesRepo{}, &playscriptsmock.Repository{})
	require.NotNil(t, h)
	assert.NotNil(t, h.stagesRepo)
	assert.NotNil(t, h.playScriptsRepo)
	assert.NotNil(t, h.connections)
	assert.NotNil(t, h.logger)
	assert.NotNil(t, h.tracer)
}

func TestHandler_buildStageState(t *testing.T) {
	t.Parallel()

	t.Run("stage_without_play_script", func(t *testing.T) {
		t.Parallel()

		stagesRepo := &fakeStagesRepo{}
		playScriptsRepo := &playscriptsmock.Repository{}
		h := buildTestHandler(t, stagesRepo, playScriptsRepo)

		stage := &stagesdomain.Stage{
			ID:               "stage-1",
			Code:             "ABC123",
			Phase:            stagesdomain.StagePhaseLobby,
			Participants:     nil,
			CurrentLineIndex: 0,
		}

		state := h.buildStageState(context.Background(), stage)
		require.NotNil(t, state)
		assert.Equal(t, "stage-1", state.ID)
		assert.Equal(t, "ABC123", state.Code)
		assert.Equal(t, stagesdomain.StagePhaseLobby, state.Phase)
		assert.Equal(t, 0, state.CurrentLineIndex)
		assert.Nil(t, state.SelectedPlay)
	})

	t.Run("stage_with_play_script_fetches_play", func(t *testing.T) {
		t.Parallel()

		playID := "play-1"
		stagesRepo := &fakeStagesRepo{}
		playScriptsRepo := &playscriptsmock.Repository{}
		expectedPlay := &playscripts.PlayScript{ID: playID, Title: "Test Play"}
		playScriptsRepo.On("GetPlayScript", mock.Anything, playID).Return(expectedPlay, nil)

		h := buildTestHandler(t, stagesRepo, playScriptsRepo)

		stage := &stagesdomain.Stage{
			ID:               "stage-1",
			Code:             "ABC123",
			Phase:            stagesdomain.StagePhaseLobby,
			PlayScriptID:     &playID,
			Participants:     nil,
			CurrentLineIndex: 0,
		}

		state := h.buildStageState(context.Background(), stage)
		require.NotNil(t, state)
		assert.Equal(t, "stage-1", state.ID)
		assert.NotNil(t, state.SelectedPlay)
		assert.Equal(t, playID, state.SelectedPlay.ID)
		assert.Equal(t, "Test Play", state.SelectedPlay.Title)
		playScriptsRepo.AssertExpectations(t)
	})

	t.Run("stage_with_play_script_get_play_error_omits_play", func(t *testing.T) {
		t.Parallel()

		playID := "play-1"
		stagesRepo := &fakeStagesRepo{}
		playScriptsRepo := &playscriptsmock.Repository{}
		playScriptsRepo.On("GetPlayScript", mock.Anything, playID).Return(nil, errors.New("db error"))

		h := buildTestHandler(t, stagesRepo, playScriptsRepo)

		stage := &stagesdomain.Stage{
			ID:               "stage-1",
			Code:             "ABC123",
			Phase:            stagesdomain.StagePhaseLobby,
			PlayScriptID:     &playID,
			Participants:     nil,
			CurrentLineIndex: 0,
		}

		state := h.buildStageState(context.Background(), stage)
		require.NotNil(t, state)
		assert.Equal(t, "stage-1", state.ID)
		assert.Nil(t, state.SelectedPlay)
		playScriptsRepo.AssertExpectations(t)
	})
}

func TestHandler_WebSocket_handleJoin(t *testing.T) {
	t.Parallel()

	t.Run("empty_participant_name_returns_error", func(t *testing.T) {
		t.Parallel()

		stagesRepo := &fakeStagesRepo{}
		playScriptsRepo := &playscriptsmock.Repository{}
		h := buildTestHandler(t, stagesRepo, playScriptsRepo)

		server := httptest.NewServer(h)
		defer server.Close()

		wsURL := "ws" + server.URL[4:] + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		err = conn.WriteJSON(&stagesdomain.ClientMessage{
			Type:            stagesdomain.ClientMessageTypeJoin,
			ParticipantName: "",
		})
		require.NoError(t, err)

		var response stagesdomain.ServerMessage
		err = conn.ReadJSON(&response)
		require.NoError(t, err)
		assert.Equal(t, stagesdomain.ServerMessageTypeError, response.Type)
		assert.Contains(t, response.Error, "participant name")
	})

	t.Run("successful_join_adds_connection_and_returns_joined", func(t *testing.T) {
		t.Parallel()

		stageID := "stage-1"
		participantID := "participant-1"
		stage := &stagesdomain.Stage{
			ID:               stageID,
			Code:             "ABC123",
			Phase:            stagesdomain.StagePhaseLobby,
			Participants:     []*stagesdomain.StageParticipant{{ID: participantID, DisplayName: "Alice", StageID: stageID}},
			CurrentLineIndex: 0,
		}

		stagesRepo := &fakeStagesRepo{
			GetStageByIDFunc: func(ctx context.Context, id string) (*stagesdomain.Stage, error) {
				if id == stageID {
					return stage, nil
				}
				return nil, sql.ErrNoRows
			},
			GetStageByCodeFunc: func(ctx context.Context, code string) (*stagesdomain.Stage, error) {
				return nil, sql.ErrNoRows
			},
			CreateStageFunc: func(ctx context.Context, input *stagesdomain.StageCreationInput) (*stagesdomain.Stage, error) {
				return &stagesdomain.Stage{
					ID:               stageID,
					Code:             input.Code,
					Phase:            stagesdomain.StagePhaseLobby,
					Participants:     nil,
					CurrentLineIndex: 0,
				}, nil
			},
			AddParticipantFunc: func(ctx context.Context, input *stagesdomain.ParticipantCreationInput) (*stagesdomain.StageParticipant, error) {
				return &stagesdomain.StageParticipant{
					ID:          participantID,
					StageID:     input.StageID,
					DisplayName: input.DisplayName,
				}, nil
			},
		}

		filterResult := filtering.NewQueryFilteredResult(
			[]*playscripts.PlayScript{},
			0, 0,
			func(p *playscripts.PlayScript) string { return p.ID },
			filtering.DefaultQueryFilter(),
		)
		playScriptsRepo := &playscriptsmock.Repository{}
		playScriptsRepo.On("GetPublicPlayScripts", mock.Anything, mock.Anything).Return(filterResult, nil)

		h := buildTestHandler(t, stagesRepo, playScriptsRepo)

		server := httptest.NewServer(h)
		defer server.Close()

		u, _ := url.Parse(server.URL)
		u.Scheme = "ws"
		wsURL := u.String() + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		err = conn.WriteJSON(&stagesdomain.ClientMessage{
			Type:            stagesdomain.ClientMessageTypeJoin,
			ParticipantName: "Alice",
		})
		require.NoError(t, err)

		var response stagesdomain.ServerMessage
		err = conn.ReadJSON(&response)
		require.NoError(t, err)
		assert.Equal(t, stagesdomain.ServerMessageTypeJoined, response.Type)
		assert.Equal(t, participantID, response.ParticipantID)
		assert.NotNil(t, response.StageState)
		playScriptsRepo.AssertExpectations(t)
	})

	t.Run("unknown_message_type_returns_error", func(t *testing.T) {
		t.Parallel()

		stagesRepo := &fakeStagesRepo{}
		playScriptsRepo := &playscriptsmock.Repository{}
		h := buildTestHandler(t, stagesRepo, playScriptsRepo)

		server := httptest.NewServer(h)
		defer server.Close()

		u, _ := url.Parse(server.URL)
		u.Scheme = "ws"
		wsURL := u.String() + "/"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		err = conn.WriteJSON(&stagesdomain.ClientMessage{
			Type: stagesdomain.ClientMessageType("unknownType"),
		})
		require.NoError(t, err)

		var response stagesdomain.ServerMessage
		err = conn.ReadJSON(&response)
		require.NoError(t, err)
		assert.Equal(t, stagesdomain.ServerMessageTypeError, response.Type)
		assert.Contains(t, response.Error, "unknown message type")
	})
}

func TestGenerateStageCode(t *testing.T) {
	t.Parallel()

	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	allowed := make(map[byte]bool)
	for i := range chars {
		allowed[chars[i]] = true
	}

	for i := 0; i < 20; i++ {
		code := generateStageCode()
		assert.Len(t, code, 6, "stage code should be 6 characters")
		for _, c := range code {
			assert.True(t, allowed[byte(c)], "stage code should only use allowed characters: %q", code)
		}
	}
}
