package playscripts

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	generated "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/playscripts/generated"
)

const (
	resourceTypePlayScripts = "play_scripts"
)

var (
	_ types.PlayScriptDataManager = (*repository)(nil)
)

// GetPlayScript fetches a play script from the database with all its cast members and lines.
func (r *repository) GetPlayScript(ctx context.Context, playScriptID string) (*types.PlayScript, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if playScriptID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.PlayScriptIDKey, playScriptID)
	tracing.AttachToSpan(span, keys.PlayScriptIDKey, playScriptID)

	result, err := r.generatedQuerier.GetPlayScript(ctx, r.db, playScriptID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching play script")
	}

	playScript := &types.PlayScript{
		ID:               result.ID,
		Title:            result.Title,
		Description:      result.Description,
		Author:           result.Author,
		IsPublic:         result.IsPublic,
		CreatedAt:        result.CreatedAt,
		LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
		CreatedByUser:    result.CreatedByUser,
		BelongsToAccount: result.BelongsToAccount,
	}

	// Fetch cast members
	castMembers, err := r.generatedQuerier.GetPlayScriptCastMembersForPlayScript(ctx, r.db, playScriptID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching play script cast members")
	}

	playScript.CastMembers = make([]*types.PlayScriptCastMember, len(castMembers))
	for i, cm := range castMembers {
		playScript.CastMembers[i] = &types.PlayScriptCastMember{
			ID:           cm.ID,
			PlayScriptID: cm.PlayScriptID,
			Name:         cm.Name,
			Description:  cm.Description,
			SortOrder:    int(cm.SortOrder),
			CreatedAt:    cm.CreatedAt,
		}
	}

	// Fetch lines
	lines, err := r.generatedQuerier.GetPlayScriptLinesForPlayScript(ctx, r.db, playScriptID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching play script lines")
	}

	playScript.Lines = make([]*types.PlayScriptLine, len(lines))
	for i, l := range lines {
		autoScrollWpm := 0
		if l.AutoScrollWpm.Valid {
			autoScrollWpm = int(l.AutoScrollWpm.Int32)
		}
		playScript.Lines[i] = &types.PlayScriptLine{
			ID:            l.ID,
			PlayScriptID:  l.PlayScriptID,
			CastMemberID:  l.CastMemberID,
			LineText:      l.LineText,
			LineIndex:     int(l.LineIndex),
			Direction:     database.StringFromNullString(l.Direction),
			ScrollMode:    types.ScrollMode(l.ScrollMode),
			AutoScrollWpm: autoScrollWpm,
			CreatedAt:     l.CreatedAt,
		}
	}

	// Fetch tags
	tags, err := r.generatedQuerier.GetTagsForPlayScript(ctx, r.db, playScriptID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching tags for play script")
	}
	playScript.Tags = make([]*types.ScriptTag, len(tags))
	for i, t := range tags {
		playScript.Tags[i] = &types.ScriptTag{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			CreatedAt:   t.CreatedAt,
		}
	}

	return playScript, nil
}

// GetPlayScripts fetches a list of play scripts from the database that meet a particular filter.
func (r *repository) GetPlayScripts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.PlayScript], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetPlayScripts(ctx, r.db, &generated.GetPlayScriptsParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching play scripts from database")
	}

	var (
		data                      []*types.PlayScript
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		ps := &types.PlayScript{
			ID:               result.ID,
			Title:            result.Title,
			Description:      result.Description,
			Author:           result.Author,
			IsPublic:         result.IsPublic,
			CreatedAt:        result.CreatedAt,
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			CreatedByUser:    result.CreatedByUser,
			BelongsToAccount: result.BelongsToAccount,
		}
		tags, tagErr := r.generatedQuerier.GetTagsForPlayScript(ctx, r.db, result.ID)
		if tagErr == nil {
			ps.Tags = make([]*types.ScriptTag, len(tags))
			for i, t := range tags {
				ps.Tags[i] = &types.ScriptTag{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt}
			}
		}
		data = append(data, ps)

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.PlayScript) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetPlayScriptsForAccount fetches a list of play scripts for a specific account from the database that meet a particular filter.
func (r *repository) GetPlayScriptsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.PlayScript], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetPlayScriptsForAccount(ctx, r.db, &generated.GetPlayScriptsForAccountParams{
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
		BelongsToAccount: accountID,
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching play scripts from database")
	}

	var (
		data                      []*types.PlayScript
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		ps := &types.PlayScript{
			ID:               result.ID,
			Title:            result.Title,
			Description:      result.Description,
			Author:           result.Author,
			IsPublic:         result.IsPublic,
			CreatedAt:        result.CreatedAt,
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			CreatedByUser:    result.CreatedByUser,
			BelongsToAccount: result.BelongsToAccount,
		}
		tags, tagErr := r.generatedQuerier.GetTagsForPlayScript(ctx, r.db, result.ID)
		if tagErr == nil {
			ps.Tags = make([]*types.ScriptTag, len(tags))
			for i, t := range tags {
				ps.Tags[i] = &types.ScriptTag{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt}
			}
		}
		data = append(data, ps)

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.PlayScript) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetPublicPlayScripts fetches a list of public play scripts from the database that meet a particular filter.
func (r *repository) GetPublicPlayScripts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.PlayScript], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetPublicPlayScripts(ctx, r.db, &generated.GetPublicPlayScriptsParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching public play scripts from database")
	}

	var (
		data                      []*types.PlayScript
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		ps := &types.PlayScript{
			ID:               result.ID,
			Title:            result.Title,
			Description:      result.Description,
			Author:           result.Author,
			IsPublic:         result.IsPublic,
			CreatedAt:        result.CreatedAt,
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			CreatedByUser:    result.CreatedByUser,
			BelongsToAccount: result.BelongsToAccount,
		}
		tags, tagErr := r.generatedQuerier.GetTagsForPlayScript(ctx, r.db, result.ID)
		if tagErr == nil {
			ps.Tags = make([]*types.ScriptTag, len(tags))
			for i, t := range tags {
				ps.Tags[i] = &types.ScriptTag{ID: t.ID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt}
			}
		}
		data = append(data, ps)

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.PlayScript) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreatePlayScript creates a play script in the database with all its cast members and lines.
func (r *repository) CreatePlayScript(ctx context.Context, input *types.PlayScriptDatabaseCreationInput, castMembers []*types.PlayScriptCastMemberDatabaseCreationInput, lines []*types.PlayScriptLineDatabaseCreationInput) (*types.PlayScript, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, database.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, keys.AccountIDKey, input.BelongsToAccount)
	logger = logger.WithValue(keys.AccountIDKey, input.BelongsToAccount)

	logger.Debug("CreatePlayScript invoked")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	// Create play script
	if err = r.generatedQuerier.CreatePlayScript(ctx, tx, &generated.CreatePlayScriptParams{
		ID:               input.ID,
		Title:            input.Title,
		Description:      input.Description,
		Author:           input.Author,
		IsPublic:         input.IsPublic,
		CreatedByUser:    input.CreatedByUser,
		BelongsToAccount: input.BelongsToAccount,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "performing play script creation query")
	}

	// Create cast members
	createdCastMembers := make([]*types.PlayScriptCastMember, len(castMembers))
	for i, cm := range castMembers {
		if err = r.generatedQuerier.CreatePlayScriptCastMember(ctx, tx, &generated.CreatePlayScriptCastMemberParams{
			ID:           cm.ID,
			PlayScriptID: cm.PlayScriptID,
			Name:         cm.Name,
			Description:  cm.Description,
			SortOrder:    int32(cm.SortOrder),
		}); err != nil {
			r.RollbackTransaction(ctx, tx)
			return nil, observability.PrepareAndLogError(err, logger, span, "performing cast member creation query")
		}
		createdCastMembers[i] = &types.PlayScriptCastMember{
			ID:           cm.ID,
			PlayScriptID: cm.PlayScriptID,
			Name:         cm.Name,
			Description:  cm.Description,
			SortOrder:    cm.SortOrder,
			CreatedAt:    r.CurrentTime(),
		}
	}

	// Create lines
	createdLines := make([]*types.PlayScriptLine, len(lines))
	for i, l := range lines {
		autoScrollWpm := sql.NullInt32{}
		if l.AutoScrollWpm > 0 {
			autoScrollWpm = sql.NullInt32{Int32: int32(l.AutoScrollWpm), Valid: true}
		}
		if err = r.generatedQuerier.CreatePlayScriptLine(ctx, tx, &generated.CreatePlayScriptLineParams{
			ID:            l.ID,
			PlayScriptID:  l.PlayScriptID,
			CastMemberID:  l.CastMemberID,
			LineText:      l.LineText,
			LineIndex:     int32(l.LineIndex),
			Direction:     database.NullStringFromString(l.Direction),
			ScrollMode:    generated.ScrollModeEnum(l.ScrollMode),
			AutoScrollWpm: autoScrollWpm,
		}); err != nil {
			r.RollbackTransaction(ctx, tx)
			return nil, observability.PrepareAndLogError(err, logger, span, "performing line creation query")
		}
		createdLines[i] = &types.PlayScriptLine{
			ID:            l.ID,
			PlayScriptID:  l.PlayScriptID,
			CastMemberID:  l.CastMemberID,
			LineText:      l.LineText,
			LineIndex:     l.LineIndex,
			Direction:     l.Direction,
			ScrollMode:    l.ScrollMode,
			AutoScrollWpm: l.AutoScrollWpm,
			CreatedAt:     r.CurrentTime(),
		}
	}

	x := &types.PlayScript{
		ID:               input.ID,
		Title:            input.Title,
		Description:      input.Description,
		Author:           input.Author,
		IsPublic:         input.IsPublic,
		CastMembers:      createdCastMembers,
		Lines:            createdLines,
		CreatedByUser:    input.CreatedByUser,
		BelongsToAccount: input.BelongsToAccount,
		CreatedAt:        r.CurrentTime(),
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &x.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypePlayScripts,
		RelevantID:       x.ID,
		EventType:        audit.AuditLogEventTypeCreated,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "committing database transaction")
	}

	tracing.AttachToSpan(span, keys.PlayScriptIDKey, x.ID)

	return x, nil
}

// UpdatePlayScript updates a play script in the database.
func (r *repository) UpdatePlayScript(ctx context.Context, playScript *types.PlayScript) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if playScript == nil {
		return database.ErrNilInputProvided
	}
	logger = logger.WithValue(keys.PlayScriptIDKey, playScript.ID)
	tracing.AttachToSpan(span, keys.PlayScriptIDKey, playScript.ID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	rowsAffected, err := r.generatedQuerier.UpdatePlayScript(ctx, tx, &generated.UpdatePlayScriptParams{
		ID:          playScript.ID,
		Title:       playScript.Title,
		Description: playScript.Description,
		Author:      playScript.Author,
		IsPublic:    playScript.IsPublic,
	})
	if err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "updating play script")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &playScript.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypePlayScripts,
		RelevantID:       playScript.ID,
		EventType:        audit.AuditLogEventTypeUpdated,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing database transaction")
	}

	return nil
}

// ArchivePlayScript archives a play script from the database.
func (r *repository) ArchivePlayScript(ctx context.Context, playScriptID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if playScriptID == "" {
		return database.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, keys.PlayScriptIDKey, playScriptID)

	logger := r.logger.WithValue(keys.PlayScriptIDKey, playScriptID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	rowsAffected, err := r.generatedQuerier.ArchivePlayScript(ctx, tx, playScriptID)
	if err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "archiving play script")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	// Note: We don't have account ID here, so we'll create audit log without it
	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		ID:           identifiers.New(),
		ResourceType: resourceTypePlayScripts,
		RelevantID:   playScriptID,
		EventType:    audit.AuditLogEventTypeArchived,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing database transaction")
	}

	return nil
}
