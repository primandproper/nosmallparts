package stages

import (
	"context"
	"database/sql"
	"testing"

	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	mgr := NewManager(&fakeStagesRepo{})
	require.NotNil(t, mgr)
}

func Test_manager_CreateStage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	expected := &stagesdomain.Stage{ID: "stage-1", Code: "ABC123"}
	repo := &fakeStagesRepo{
		CreateStageFunc: func(ctx context.Context, input *stagesdomain.StageCreationInput) (*stagesdomain.Stage, error) {
			require.NotNil(t, input)
			assert.NotEmpty(t, input.ID)
			assert.NotEmpty(t, input.Code)
			return expected, nil
		},
	}
	mgr := NewManager(repo)

	got, err := mgr.CreateStage(ctx, &stagesdomain.StageCreationInput{ID: "stage-1", Code: "ABC123"})
	require.NoError(t, err)
	assert.Same(t, expected, got)
}

func Test_manager_GetStageByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	expected := &stagesdomain.Stage{ID: "stage-1", Code: "ABC123"}
	repo := &fakeStagesRepo{
		GetStageByIDFunc: func(ctx context.Context, stageID string) (*stagesdomain.Stage, error) {
			assert.Equal(t, "stage-1", stageID)
			return expected, nil
		},
	}
	mgr := NewManager(repo)

	got, err := mgr.GetStageByID(ctx, "stage-1")
	require.NoError(t, err)
	assert.Same(t, expected, got)
}

func Test_manager_GetStageByCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	expected := &stagesdomain.Stage{ID: "stage-1", Code: "XYZ789"}
	repo := &fakeStagesRepo{
		GetStageByCodeFunc: func(ctx context.Context, code string) (*stagesdomain.Stage, error) {
			assert.Equal(t, "XYZ789", code)
			return expected, nil
		},
	}
	mgr := NewManager(repo)

	got, err := mgr.GetStageByCode(ctx, "XYZ789")
	require.NoError(t, err)
	assert.Same(t, expected, got)
}

func Test_manager_GetStageByCode_not_found(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &fakeStagesRepo{
		GetStageByCodeFunc: func(ctx context.Context, code string) (*stagesdomain.Stage, error) {
			return nil, sql.ErrNoRows
		},
	}
	mgr := NewManager(repo)

	got, err := mgr.GetStageByCode(ctx, "NONEXIST")
	assert.Nil(t, got)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func Test_manager_UpdateStage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &fakeStagesRepo{
		UpdateStageFunc: func(ctx context.Context, stageID string, input *stagesdomain.StageUpdateInput) error {
			assert.Equal(t, "stage-1", stageID)
			require.NotNil(t, input)
			return nil
		},
	}
	mgr := NewManager(repo)

	err := mgr.UpdateStage(ctx, "stage-1", &stagesdomain.StageUpdateInput{})
	require.NoError(t, err)
}

func Test_manager_ArchiveStage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &fakeStagesRepo{
		ArchiveStageFunc: func(ctx context.Context, stageID string) error {
			assert.Equal(t, "stage-1", stageID)
			return nil
		},
	}
	mgr := NewManager(repo)

	err := mgr.ArchiveStage(ctx, "stage-1")
	require.NoError(t, err)
}

func Test_manager_AddParticipant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	expected := &stagesdomain.StageParticipant{ID: "p1", StageID: "stage-1", DisplayName: "Alice"}
	repo := &fakeStagesRepo{
		AddParticipantFunc: func(ctx context.Context, input *stagesdomain.ParticipantCreationInput) (*stagesdomain.StageParticipant, error) {
			require.NotNil(t, input)
			assert.Equal(t, "stage-1", input.StageID)
			assert.Equal(t, "Alice", input.DisplayName)
			return expected, nil
		},
	}
	mgr := NewManager(repo)

	got, err := mgr.AddParticipant(ctx, &stagesdomain.ParticipantCreationInput{StageID: "stage-1", DisplayName: "Alice"})
	require.NoError(t, err)
	assert.Same(t, expected, got)
}
