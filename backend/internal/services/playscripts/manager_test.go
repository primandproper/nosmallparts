package playscripts

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	playscriptsmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	repo := &playscriptsmock.Repository{}
	mgr := NewManager(repo)
	require.NotNil(t, mgr)
}

func Test_manager_CreatePlayScript(t *testing.T) {
	t.Parallel()

	t.Run("nil_input_returns_error", func(t *testing.T) {
		t.Parallel()

		repo := &playscriptsmock.Repository{}
		mgr := NewManager(repo)
		got, err := mgr.CreatePlayScript(context.Background(), nil, "acc-1", "user-1")
		assert.Nil(t, got)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "input is required")
	})

	t.Run("validation_error_returns_error", func(t *testing.T) {
		t.Parallel()

		repo := &playscriptsmock.Repository{}
		mgr := NewManager(repo)
		input := &playscripts.PlayScriptCreationRequestInput{} // Title required
		got, err := mgr.CreatePlayScript(context.Background(), input, "acc-1", "user-1")
		assert.Nil(t, got)
		require.Error(t, err)
	})

	t.Run("success_delegates_to_repo", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &playscriptsmock.Repository{}
		expected := &playscripts.PlayScript{ID: "ps-1", Title: "Test Play", BelongsToAccount: "acc-1"}
		repo.On("CreatePlayScript", ctx, mock.AnythingOfType("*playscripts.PlayScriptDatabaseCreationInput"),
			mock.AnythingOfType("[]*playscripts.PlayScriptCastMemberDatabaseCreationInput"),
			mock.AnythingOfType("[]*playscripts.PlayScriptLineDatabaseCreationInput")).
			Return(expected, nil).Once()

		mgr := NewManager(repo)
		input := &playscripts.PlayScriptCreationRequestInput{Title: "Test Play"}
		got, err := mgr.CreatePlayScript(ctx, input, "acc-1", "user-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "ps-1", got.ID)
		assert.Equal(t, "Test Play", got.Title)
		repo.AssertExpectations(t)
	})
}

func Test_manager_GetPlayScript(t *testing.T) {
	t.Parallel()

	t.Run("public_play_returns_without_account", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &playscriptsmock.Repository{}
		expected := &playscripts.PlayScript{ID: "ps-1", IsPublic: true}
		repo.On("GetPlayScript", ctx, "ps-1").Return(expected, nil).Once()

		mgr := NewManager(repo)
		got, err := mgr.GetPlayScript(ctx, "ps-1", "")
		require.NoError(t, err)
		assert.Same(t, expected, got)
		repo.AssertExpectations(t)
	})

	t.Run("same_account_returns_private_play", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &playscriptsmock.Repository{}
		expected := &playscripts.PlayScript{ID: "ps-1", BelongsToAccount: "acc-1", IsPublic: false}
		repo.On("GetPlayScript", ctx, "ps-1").Return(expected, nil).Once()

		mgr := NewManager(repo)
		got, err := mgr.GetPlayScript(ctx, "ps-1", "acc-1")
		require.NoError(t, err)
		assert.Same(t, expected, got)
		repo.AssertExpectations(t)
	})

	t.Run("wrong_account_returns_permission_denied", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &playscriptsmock.Repository{}
		play := &playscripts.PlayScript{ID: "ps-1", BelongsToAccount: "acc-1", IsPublic: false}
		repo.On("GetPlayScript", ctx, "ps-1").Return(play, nil).Once()

		mgr := NewManager(repo)
		got, err := mgr.GetPlayScript(ctx, "ps-1", "other-account")
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrPermissionDenied))
		repo.AssertExpectations(t)
	})

	t.Run("repo_error_propagates", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &playscriptsmock.Repository{}
		repoErr := errors.New("db error")
		repo.On("GetPlayScript", ctx, "ps-1").Return(nil, repoErr).Once()

		mgr := NewManager(repo)
		got, err := mgr.GetPlayScript(ctx, "ps-1", "")
		assert.Nil(t, got)
		assert.Same(t, repoErr, err)
		repo.AssertExpectations(t)
	})
}

func Test_manager_GetPlayScriptsForAccount(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &playscriptsmock.Repository{}
	qf := filtering.DefaultQueryFilter()
	result := filtering.NewQueryFilteredResult(
		[]*playscripts.PlayScript{},
		0, 0,
		func(p *playscripts.PlayScript) string { return p.ID },
		qf,
	)
	repo.On("GetPlayScriptsForAccount", ctx, "acc-1", qf).Return(result, nil).Once()

	mgr := NewManager(repo)
	got, err := mgr.GetPlayScriptsForAccount(ctx, "acc-1", qf)
	require.NoError(t, err)
	assert.Same(t, result, got)
	repo.AssertExpectations(t)
}

func Test_manager_GetPublicPlayScripts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := &playscriptsmock.Repository{}
	qf := filtering.DefaultQueryFilter()
	result := filtering.NewQueryFilteredResult(
		[]*playscripts.PlayScript{},
		0, 0,
		func(p *playscripts.PlayScript) string { return p.ID },
		qf,
	)
	repo.On("GetPublicPlayScripts", ctx, qf).Return(result, nil).Once()

	mgr := NewManager(repo)
	got, err := mgr.GetPublicPlayScripts(ctx, qf)
	require.NoError(t, err)
	assert.Same(t, result, got)
	repo.AssertExpectations(t)
}

func Test_manager_UpdatePlayScript(t *testing.T) {
	t.Parallel()

	t.Run("wrong_account_returns_permission_denied", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &playscriptsmock.Repository{}
		play := &playscripts.PlayScript{ID: "ps-1", BelongsToAccount: "acc-1"}
		repo.On("GetPlayScript", ctx, "ps-1").Return(play, nil).Once()

		mgr := NewManager(repo)
		title := "Updated"
		input := &playscripts.PlayScriptUpdateRequestInput{Title: &title}
		got, err := mgr.UpdatePlayScript(ctx, "ps-1", input, "other-account")
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrPermissionDenied))
		repo.AssertExpectations(t)
	})

	t.Run("success_updates_and_returns", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &playscriptsmock.Repository{}
		play := &playscripts.PlayScript{ID: "ps-1", Title: "Old", BelongsToAccount: "acc-1"}
		repo.On("GetPlayScript", ctx, "ps-1").Return(play, nil).Once()
		repo.On("UpdatePlayScript", ctx, mock.MatchedBy(func(p *playscripts.PlayScript) bool {
			return p.ID == "ps-1" && p.Title == "New"
		})).Return(nil).Once()

		mgr := NewManager(repo)
		title := "New"
		input := &playscripts.PlayScriptUpdateRequestInput{Title: &title}
		got, err := mgr.UpdatePlayScript(ctx, "ps-1", input, "acc-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "New", got.Title)
		repo.AssertExpectations(t)
	})
}

func Test_manager_ArchivePlayScript(t *testing.T) {
	t.Parallel()

	t.Run("wrong_account_returns_permission_denied", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &playscriptsmock.Repository{}
		play := &playscripts.PlayScript{ID: "ps-1", BelongsToAccount: "acc-1"}
		repo.On("GetPlayScript", ctx, "ps-1").Return(play, nil).Once()

		mgr := NewManager(repo)
		err := mgr.ArchivePlayScript(ctx, "ps-1", "other-account")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrPermissionDenied))
		repo.AssertExpectations(t)
	})

	t.Run("success_archives", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &playscriptsmock.Repository{}
		play := &playscripts.PlayScript{ID: "ps-1", BelongsToAccount: "acc-1"}
		repo.On("GetPlayScript", ctx, "ps-1").Return(play, nil).Once()
		repo.On("ArchivePlayScript", ctx, "ps-1").Return(nil).Once()

		mgr := NewManager(repo)
		err := mgr.ArchivePlayScript(ctx, "ps-1", "acc-1")
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
