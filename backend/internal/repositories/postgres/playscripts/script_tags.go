package playscripts

import (
	"context"

	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	generated "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/playscripts/generated"
)

var _ types.ScriptTagDataManager = (*repository)(nil)

// CreateScriptTag creates a new script tag.
func (r *repository) CreateScriptTag(ctx context.Context, id, name, description string) (*types.ScriptTag, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	if id == "" {
		id = identifiers.New()
	}
	tracing.AttachToSpan(span, keys.PlayScriptIDKey, id)

	if err := r.generatedQuerier.CreateScriptTag(ctx, r.db, &generated.CreateScriptTagParams{
		ID:          id,
		Name:        name,
		Description: description,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating script tag")
	}

	tag, err := r.generatedQuerier.GetScriptTag(ctx, r.db, id)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching created script tag")
	}

	return &types.ScriptTag{
		ID:          tag.ID,
		Name:        tag.Name,
		Description: tag.Description,
		CreatedAt:   tag.CreatedAt,
	}, nil
}

// GetScriptTag fetches a script tag by ID.
func (r *repository) GetScriptTag(ctx context.Context, id string) (*types.ScriptTag, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tag, err := r.generatedQuerier.GetScriptTag(ctx, r.db, id)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger.Clone(), span, "fetching script tag")
	}

	return &types.ScriptTag{
		ID:          tag.ID,
		Name:        tag.Name,
		Description: tag.Description,
		CreatedAt:   tag.CreatedAt,
	}, nil
}

// GetScriptTagByName fetches a script tag by name.
func (r *repository) GetScriptTagByName(ctx context.Context, name string) (*types.ScriptTag, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tag, err := r.generatedQuerier.GetScriptTagByName(ctx, r.db, name)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger.Clone(), span, "fetching script tag by name")
	}

	return &types.ScriptTag{
		ID:          tag.ID,
		Name:        tag.Name,
		Description: tag.Description,
		CreatedAt:   tag.CreatedAt,
	}, nil
}

// ListScriptTags fetches all script tags ordered by name.
func (r *repository) ListScriptTags(ctx context.Context) ([]*types.ScriptTag, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tags, err := r.generatedQuerier.ListScriptTags(ctx, r.db)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger.Clone(), span, "listing script tags")
	}

	result := make([]*types.ScriptTag, len(tags))
	for i, t := range tags {
		result[i] = &types.ScriptTag{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			CreatedAt:   t.CreatedAt,
		}
	}
	return result, nil
}

// GetTagsForPlayScript fetches all active tags for a play script.
func (r *repository) GetTagsForPlayScript(ctx context.Context, playScriptID string) ([]*types.ScriptTag, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tags, err := r.generatedQuerier.GetTagsForPlayScript(ctx, r.db, playScriptID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger.Clone(), span, "fetching tags for play script")
	}

	result := make([]*types.ScriptTag, len(tags))
	for i, t := range tags {
		result[i] = &types.ScriptTag{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			CreatedAt:   t.CreatedAt,
		}
	}
	return result, nil
}

// AddTagToPlayScript associates a tag with a play script.
func (r *repository) AddTagToPlayScript(ctx context.Context, playScriptID, scriptTagID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	return observability.PrepareAndLogError(
		r.generatedQuerier.AddTagToPlayScript(ctx, r.db, &generated.AddTagToPlayScriptParams{
			PlayScriptID: playScriptID,
			ScriptTagID:  scriptTagID,
		}),
		r.logger.Clone(),
		span,
		"adding tag to play script",
	)
}

// RemoveTagFromPlayScript soft-deletes the tag assignment.
func (r *repository) RemoveTagFromPlayScript(ctx context.Context, playScriptID, scriptTagID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	_, err := r.generatedQuerier.RemoveTagFromPlayScript(ctx, r.db, &generated.RemoveTagFromPlayScriptParams{
		PlayScriptID: playScriptID,
		ScriptTagID:  scriptTagID,
	})
	return observability.PrepareAndLogError(err, r.logger.Clone(), span, "removing tag from play script")
}

// GetPlayScriptsWithTag fetches play scripts that have a tag with the given name.
func (r *repository) GetPlayScriptsWithTag(ctx context.Context, tagName string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.PlayScript], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	results, err := r.generatedQuerier.GetPlayScriptsWithTag(ctx, r.db, &generated.GetPlayScriptsWithTagParams{
		Name:          tagName,
		CreatedAfter:  database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore: database.NullTimeFromTimePointer(filter.CreatedBefore),
		Cursor:        database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:   database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching play scripts with tag")
	}

	data := make([]*types.PlayScript, 0, len(results))
	var filteredCount, totalCount uint64
	for _, row := range results {
		data = append(data, &types.PlayScript{
			ID:               row.ID,
			Title:            row.Title,
			Description:      row.Description,
			Author:           row.Author,
			IsPublic:         row.IsPublic,
			CreatedAt:        row.CreatedAt,
			LastUpdatedAt:    database.TimePointerFromNullTime(row.LastUpdatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(row.ArchivedAt),
			CreatedByUser:    row.CreatedByUser,
			BelongsToAccount: row.BelongsToAccount,
		})
		filteredCount = uint64(row.FilteredCount)
		totalCount = uint64(row.TotalCount)
	}

	return filtering.NewQueryFilteredResult(data, filteredCount, totalCount, func(t *types.PlayScript) string { return t.ID }, filter), nil
}
