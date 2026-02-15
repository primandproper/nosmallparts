package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"

	"github.com/stretchr/testify/mock"
)

var _ playscripts.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

// GetPlayScript is a mock function.
func (m *Repository) GetPlayScript(ctx context.Context, playScriptID string, includeLinkedScripts bool) (*playscripts.PlayScript, error) {
	args := m.Called(ctx, playScriptID, includeLinkedScripts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*playscripts.PlayScript), args.Error(1)
}

// GetPlayScripts is a mock function.
func (m *Repository) GetPlayScripts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[playscripts.PlayScript]), args.Error(1)
}

// GetPlayScriptsForAccount is a mock function.
func (m *Repository) GetPlayScriptsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error) {
	args := m.Called(ctx, accountID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[playscripts.PlayScript]), args.Error(1)
}

// GetPublicPlayScripts is a mock function.
func (m *Repository) GetPublicPlayScripts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[playscripts.PlayScript]), args.Error(1)
}

// CreatePlayScript is a mock function.
func (m *Repository) CreatePlayScript(ctx context.Context, input *playscripts.PlayScriptDatabaseCreationInput, castMembers []*playscripts.PlayScriptCastMemberDatabaseCreationInput, lines []*playscripts.PlayScriptLineDatabaseCreationInput) (*playscripts.PlayScript, error) {
	args := m.Called(ctx, input, castMembers, lines)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*playscripts.PlayScript), args.Error(1)
}

// UpdatePlayScript is a mock function.
func (m *Repository) UpdatePlayScript(ctx context.Context, playScript *playscripts.PlayScript) error {
	args := m.Called(ctx, playScript)
	return args.Error(0)
}

// ArchivePlayScript is a mock function.
func (m *Repository) ArchivePlayScript(ctx context.Context, playScriptID string) error {
	args := m.Called(ctx, playScriptID)
	return args.Error(0)
}

// CreateScriptTag is a mock function.
func (m *Repository) CreateScriptTag(ctx context.Context, id, name, description string) (*playscripts.ScriptTag, error) {
	args := m.Called(ctx, id, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*playscripts.ScriptTag), args.Error(1)
}

// GetScriptTag is a mock function.
func (m *Repository) GetScriptTag(ctx context.Context, id string) (*playscripts.ScriptTag, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*playscripts.ScriptTag), args.Error(1)
}

// GetScriptTagByName is a mock function.
func (m *Repository) GetScriptTagByName(ctx context.Context, name string) (*playscripts.ScriptTag, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*playscripts.ScriptTag), args.Error(1)
}

// GetTagsForPlayScript is a mock function.
func (m *Repository) GetTagsForPlayScript(ctx context.Context, playScriptID string) ([]*playscripts.ScriptTag, error) {
	args := m.Called(ctx, playScriptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*playscripts.ScriptTag), args.Error(1)
}

// AddTagToPlayScript is a mock function.
func (m *Repository) AddTagToPlayScript(ctx context.Context, playScriptID, scriptTagID string) error {
	args := m.Called(ctx, playScriptID, scriptTagID)
	return args.Error(0)
}

// RemoveTagFromPlayScript is a mock function.
func (m *Repository) RemoveTagFromPlayScript(ctx context.Context, playScriptID, scriptTagID string) error {
	args := m.Called(ctx, playScriptID, scriptTagID)
	return args.Error(0)
}

// GetPlayScriptsWithTag is a mock function.
func (m *Repository) GetPlayScriptsWithTag(ctx context.Context, tagName string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[playscripts.PlayScript], error) {
	args := m.Called(ctx, tagName, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[playscripts.PlayScript]), args.Error(1)
}
