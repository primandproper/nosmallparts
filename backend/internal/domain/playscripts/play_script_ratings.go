package playscripts

import (
	"context"
	"time"
)

// PlayScriptRating represents a user's rating of a play script (1-5).
type PlayScriptRating struct {
	_             struct{}   `json:"-"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastUpdatedAt *time.Time `json:"lastUpdatedAt,omitempty"`
	ArchivedAt    *time.Time `json:"archivedAt,omitempty"`
	ID            string     `json:"id"`
	PlayScriptID  string     `json:"playScriptId"`
	BelongsToUser string     `json:"belongsToUser"`
	Rating        int        `json:"rating"`
}

// PlayScriptRatingStats holds aggregate rating data for a play script.
type PlayScriptRatingStats struct {
	AverageRating float64 `json:"averageRating"`
	RatingCount   int     `json:"ratingCount"`
}

// PlayScriptRatingDataManager describes persistence for play script ratings.
type PlayScriptRatingDataManager interface {
	RatePlayScript(ctx context.Context, userID, playScriptID string, rating int) (*PlayScriptRating, error)
	GetUserRatingForPlayScript(ctx context.Context, userID, playScriptID string) (*PlayScriptRating, error)
	GetRatingStatsForPlayScript(ctx context.Context, playScriptID string) (*PlayScriptRatingStats, error)
	GetRatingStatsForPlayScripts(ctx context.Context, playScriptIDs []string) (map[string]*PlayScriptRatingStats, error)
	GetUserRatingsForPlayScripts(ctx context.Context, userID string, playScriptIDs []string) (map[string]int, error)
}
