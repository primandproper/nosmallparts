package playscripts

import (
	"context"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
)

// ScriptTag represents a tag that can be applied to play scripts (e.g. "profanity", "one-act").
type ScriptTag struct {
	CreatedAt   time.Time `json:"createdAt"`
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// ScriptTagDataManager describes persistence for script tags and tag assignments.
type ScriptTagDataManager interface {
	CreateScriptTag(ctx context.Context, id, name, description string) (*ScriptTag, error)
	GetScriptTag(ctx context.Context, id string) (*ScriptTag, error)
	GetScriptTagByName(ctx context.Context, name string) (*ScriptTag, error)
	ListScriptTags(ctx context.Context) ([]*ScriptTag, error)
	GetTagsForPlayScript(ctx context.Context, playScriptID string) ([]*ScriptTag, error)
	AddTagToPlayScript(ctx context.Context, playScriptID, scriptTagID string) error
	RemoveTagFromPlayScript(ctx context.Context, playScriptID, scriptTagID string) error
	GetPlayScriptsWithTag(ctx context.Context, tagName string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[PlayScript], error)
}
