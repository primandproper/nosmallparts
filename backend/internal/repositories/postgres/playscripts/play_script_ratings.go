package playscripts

import (
	"context"
	"database/sql"
	"errors"

	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	generated "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/playscripts/generated"
)

var (
	_ types.PlayScriptRatingDataManager = (*repository)(nil)
)

// ratingFromInterface converts sqlc's interface{} for average_rating to float64.
func ratingFromInterface(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	default:
		return 0
	}
}

// RatePlayScript archives any existing rating for the user+play and inserts the new one.
func (r *repository) RatePlayScript(ctx context.Context, userID, playScriptID string, rating int) (*types.PlayScriptRating, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()
	if userID == "" || playScriptID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.PlayScriptIDKey, playScriptID).WithValue(keys.UserIDKey, userID)
	tracing.AttachToSpan(span, keys.PlayScriptIDKey, playScriptID)
	tracing.AttachToSpan(span, keys.UserIDKey, userID)

	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	_, err = r.generatedQuerier.ArchiveRatingForUserAndPlay(ctx, tx, &generated.ArchiveRatingForUserAndPlayParams{
		PlayScriptID:  playScriptID,
		BelongsToUser: userID,
	})
	if err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "archiving play script rating")
	}

	newID := identifiers.New()
	now := r.CurrentTime()
	if err = r.generatedQuerier.CreatePlayScriptRating(ctx, tx, &generated.CreatePlayScriptRatingParams{
		ID:            newID,
		PlayScriptID:  playScriptID,
		BelongsToUser: userID,
		Rating:        int32(rating),
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "creating play script rating")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	return &types.PlayScriptRating{
		ID:            newID,
		PlayScriptID:  playScriptID,
		BelongsToUser: userID,
		Rating:        rating,
		CreatedAt:     now,
		LastUpdatedAt: nil,
		ArchivedAt:    nil,
	}, nil
}

// GetUserRatingForPlayScript returns the current user's rating for a play, or nil if none.
func (r *repository) GetUserRatingForPlayScript(ctx context.Context, userID, playScriptID string) (*types.PlayScriptRating, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || playScriptID == "" {
		return nil, database.ErrInvalidIDProvided
	}

	row, err := r.generatedQuerier.GetActiveRatingForUserAndPlay(ctx, r.db, &generated.GetActiveRatingForUserAndPlayParams{
		PlayScriptID:  playScriptID,
		BelongsToUser: userID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger, span, "fetching user rating for play script")
	}

	return &types.PlayScriptRating{
		ID:            row.ID,
		PlayScriptID:  row.PlayScriptID,
		BelongsToUser: row.BelongsToUser,
		Rating:        int(row.Rating),
		CreatedAt:     row.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(row.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(row.ArchivedAt),
	}, nil
}

// GetRatingStatsForPlayScript returns aggregate rating stats for a play script.
func (r *repository) GetRatingStatsForPlayScript(ctx context.Context, playScriptID string) (*types.PlayScriptRatingStats, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if playScriptID == "" {
		return nil, database.ErrInvalidIDProvided
	}

	row, err := r.generatedQuerier.GetRatingStatsForPlayScript(ctx, r.db, playScriptID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &types.PlayScriptRatingStats{AverageRating: 0, RatingCount: 0}, nil
		}
		return nil, observability.PrepareAndLogError(err, r.logger.Clone(), span, "fetching rating stats for play script")
	}

	return &types.PlayScriptRatingStats{
		AverageRating: ratingFromInterface(row.AverageRating),
		RatingCount:   int(row.RatingCount),
	}, nil
}

// GetRatingStatsForPlayScripts returns aggregate rating stats for multiple play scripts.
func (r *repository) GetRatingStatsForPlayScripts(ctx context.Context, playScriptIDs []string) (map[string]*types.PlayScriptRatingStats, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	result := make(map[string]*types.PlayScriptRatingStats, len(playScriptIDs))
	for _, id := range playScriptIDs {
		result[id] = &types.PlayScriptRatingStats{AverageRating: 0, RatingCount: 0}
	}

	if len(playScriptIDs) == 0 {
		return result, nil
	}

	rows, err := r.generatedQuerier.GetRatingStatsForPlayScripts(ctx, r.db, playScriptIDs)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger.Clone(), span, "fetching rating stats for play scripts")
	}

	for _, row := range rows {
		result[row.PlayScriptID] = &types.PlayScriptRatingStats{
			AverageRating: ratingFromInterface(row.AverageRating),
			RatingCount:   int(row.RatingCount),
		}
	}

	return result, nil
}

// GetUserRatingsForPlayScripts returns the user's ratings for the given play scripts.
func (r *repository) GetUserRatingsForPlayScripts(ctx context.Context, userID string, playScriptIDs []string) (map[string]int, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	result := make(map[string]int)
	if userID == "" || len(playScriptIDs) == 0 {
		return result, nil
	}

	rows, err := r.generatedQuerier.GetUserRatingsForPlayScripts(ctx, r.db, &generated.GetUserRatingsForPlayScriptsParams{
		BelongsToUser: userID,
		PlayScriptIds: playScriptIDs,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger.Clone(), span, "fetching user ratings for play scripts")
	}

	for _, row := range rows {
		result[row.PlayScriptID] = int(row.Rating)
	}

	return result, nil
}
