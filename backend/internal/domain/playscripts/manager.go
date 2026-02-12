package playscripts

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
)

// PlayScriptsManager describes operations for managing play scripts (service-level interface).
// Auth is implied via accountID/userID parameters; the HTTP layer extracts session and passes them.
type PlayScriptsManager interface {
	CreatePlayScript(ctx context.Context, input *PlayScriptCreationRequestInput, accountID, userID string) (*PlayScript, error)
	GetPlayScript(ctx context.Context, playScriptID, accountID string) (*PlayScript, error)
	GetPlayScriptsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[PlayScript], error)
	GetPublicPlayScripts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[PlayScript], error)
	UpdatePlayScript(ctx context.Context, playScriptID string, input *PlayScriptUpdateRequestInput, accountID string) (*PlayScript, error)
	ArchivePlayScript(ctx context.Context, playScriptID, accountID string) error
}
