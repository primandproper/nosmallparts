package playscripts

import (
	"context"
	"errors"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
)

// ErrPermissionDenied is returned when the caller does not have access to the resource.
var ErrPermissionDenied = errors.New("permission denied")

var _ playscripts.PlayScriptsManager = (*manager)(nil)

type manager struct {
	repo playscripts.Repository
}

// NewManager returns a PlayScriptsManager that delegates to the given repository and applies ownership checks.
func NewManager(repo playscripts.Repository) playscripts.PlayScriptsManager {
	return &manager{repo: repo}
}

func (m *manager) CreatePlayScript(ctx context.Context, input *playscripts.PlayScriptCreationRequestInput, accountID, userID string) (*playscripts.PlayScript, error) {
	if input == nil {
		return nil, errors.New("input is required")
	}
	dbInput := converters.ConvertPlayScriptCreationRequestInputToPlayScriptDatabaseCreationInput(input, userID, accountID)
	if err := dbInput.ValidateWithContext(ctx); err != nil {
		return nil, err
	}

	castMembers := input.CastMembers
	if castMembers == nil {
		castMembers = []*playscripts.PlayScriptCastMemberCreationInput{}
	}
	lines := input.Lines
	if lines == nil {
		lines = []*playscripts.PlayScriptLineCreationRequestInput{}
	}

	castMemberDbInputs, idMapping := converters.ConvertCastMemberCreationInputsToDatabaseInputs(dbInput.ID, castMembers)
	lineDbInputs := converters.ConvertLineCreationInputsToDatabaseInputs(dbInput.ID, lines, idMapping)

	return m.repo.CreatePlayScript(ctx, dbInput, castMemberDbInputs, lineDbInputs)
}

func (m *manager) GetPlayScript(ctx context.Context, playScriptID, accountID string) (*playscripts.PlayScript, error) {
	play, err := m.repo.GetPlayScript(ctx, playScriptID)
	if err != nil {
		return nil, err
	}
	if play.IsPublic {
		return play, nil
	}
	if accountID != "" && play.BelongsToAccount == accountID {
		return play, nil
	}
	return nil, ErrPermissionDenied
}

func (m *manager) GetPlayScriptsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error) {
	return m.repo.GetPlayScriptsForAccount(ctx, accountID, filter)
}

func (m *manager) GetPublicPlayScripts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error) {
	return m.repo.GetPublicPlayScripts(ctx, filter)
}

func (m *manager) UpdatePlayScript(ctx context.Context, playScriptID string, input *playscripts.PlayScriptUpdateRequestInput, accountID string) (*playscripts.PlayScript, error) {
	play, err := m.repo.GetPlayScript(ctx, playScriptID)
	if err != nil {
		return nil, err
	}
	if play.BelongsToAccount != accountID {
		return nil, ErrPermissionDenied
	}
	if input != nil {
		play.Update(input)
	}
	if updateErr := m.repo.UpdatePlayScript(ctx, play); updateErr != nil {
		return nil, updateErr
	}
	return play, nil
}

func (m *manager) ArchivePlayScript(ctx context.Context, playScriptID, accountID string) error {
	play, err := m.repo.GetPlayScript(ctx, playScriptID)
	if err != nil {
		return err
	}
	if play.BelongsToAccount != accountID {
		return ErrPermissionDenied
	}
	return m.repo.ArchivePlayScript(ctx, playScriptID)
}
