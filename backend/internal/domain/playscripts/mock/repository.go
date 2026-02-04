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
func (m *Repository) GetPlayScript(ctx context.Context, playScriptID string) (*playscripts.PlayScript, error) {
	args := m.Called(ctx, playScriptID)
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
