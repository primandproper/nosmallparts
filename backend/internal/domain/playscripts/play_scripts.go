package playscripts

import (
	"context"
	"encoding/gob"
	"sort"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// PlayScriptCreatedServiceEventType indicates a play script was created.
	PlayScriptCreatedServiceEventType = "play_script_created"
	// PlayScriptUpdatedServiceEventType indicates a play script was updated.
	PlayScriptUpdatedServiceEventType = "play_script_updated"
	// PlayScriptArchivedServiceEventType indicates a play script was archived.
	PlayScriptArchivedServiceEventType = "play_script_archived"
)

// ScrollMode enum for line display behavior (maps to PostgreSQL scroll_mode_enum).
type ScrollMode string

const (
	ScrollModeAuto   ScrollMode = "auto"   // Default value - auto-scroll at specified WPM
	ScrollModeManual ScrollMode = "manual" // Manual scroll control
)

// ValidScrollModes returns all valid scroll mode values.
func ValidScrollModes() []ScrollMode {
	return []ScrollMode{ScrollModeAuto, ScrollModeManual}
}

// IsValid checks if the scroll mode is a valid value.
func (s ScrollMode) IsValid() bool {
	switch s {
	case ScrollModeAuto, ScrollModeManual:
		return true
	}
	return false
}

// DefaultScrollMode returns the default scroll mode.
func DefaultScrollMode() ScrollMode {
	return ScrollModeAuto
}

func init() {
	gob.Register(new(PlayScript))
	gob.Register(new(PlayScriptCastMember))
	gob.Register(new(PlayScriptLine))
	gob.Register(new(PlayScriptCreationRequestInput))
	gob.Register(new(PlayScriptDatabaseCreationInput))
	gob.Register(new(PlayScriptUpdateRequestInput))
	gob.Register(new(PlayScriptRating))
	gob.Register(new(PlayScriptRatingStats))
}

type (
	// PlayScript represents a theatrical script with cast members and lines.
	PlayScript struct {
		_                struct{}                `json:"-"`
		CreatedAt        time.Time               `json:"createdAt"`
		LastUpdatedAt    *time.Time              `json:"lastUpdatedAt"`
		ArchivedAt       *time.Time              `json:"archivedAt"`
		Precedes         *PlayScript             `json:"precedes,omitempty"`
		PrecededBy       *PlayScript             `json:"precededBy,omitempty"`
		PrecedesID       *string                 `json:"precedesId,omitempty"`
		PrecededByID     *string                 `json:"precededById,omitempty"`
		RatingStats      *PlayScriptRatingStats  `json:"ratingStats,omitempty"`
		Author           string                  `json:"author,omitempty"`
		BelongsToAccount string                  `json:"belongsToAccount"`
		ID               string                  `json:"id"`
		Title            string                  `json:"title"`
		Description      string                  `json:"description"`
		CreatedByUser    string                  `json:"createdByUser"`
		CastMembers      []*PlayScriptCastMember `json:"castMembers"`
		Lines            []*PlayScriptLine       `json:"lines"`
		Tags             []*ScriptTag            `json:"tags"`
		UserRating       int                     `json:"userRating,omitempty"`
		IsPublic         bool                    `json:"isPublic"`
	}

	// PlayScriptCastMember represents a cast member (character) in a play script.
	PlayScriptCastMember struct {
		_            struct{}  `json:"-"`
		CreatedAt    time.Time `json:"createdAt"`
		ID           string    `json:"id"`
		PlayScriptID string    `json:"playScriptId"`
		Name         string    `json:"name"`
		Description  string    `json:"description,omitempty"`
		SortOrder    int       `json:"sortOrder"`
	}

	// PlayScriptLine represents a single line in a play script.
	PlayScriptLine struct {
		_             struct{}   `json:"-"`
		CreatedAt     time.Time  `json:"createdAt"`
		ID            string     `json:"id"`
		PlayScriptID  string     `json:"playScriptId"`
		CastMemberID  string     `json:"castMemberId"`
		LineText      string     `json:"lineText"`
		Direction     string     `json:"direction,omitempty"`
		ScrollMode    ScrollMode `json:"scrollMode"`
		LineIndex     int        `json:"lineIndex"`
		AutoScrollWpm int        `json:"autoScrollWpm,omitempty"`
	}

	// PlayScriptCreationRequestInput represents input for creating a play script.
	PlayScriptCreationRequestInput struct {
		_            struct{}                              `json:"-"`
		PrecededByID *string                               `json:"precededById,omitempty"`
		PrecedesID   *string                               `json:"precedesId,omitempty"`
		ID           string                                `json:"id,omitempty"`
		Title        string                                `json:"title"`
		Description  string                                `json:"description"`
		Author       string                                `json:"author,omitempty"`
		CastMembers  []*PlayScriptCastMemberCreationInput  `json:"castMembers"`
		Lines        []*PlayScriptLineCreationRequestInput `json:"lines"`
		IsPublic     bool                                  `json:"isPublic"`
	}

	// PlayScriptCastMemberCreationInput represents input for creating a cast member.
	PlayScriptCastMemberCreationInput struct {
		_           struct{} `json:"-"`
		ID          string   `json:"id"` // Client-provided ID for referencing in lines
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		SortOrder   int      `json:"sortOrder"`
	}

	// PlayScriptLineCreationRequestInput represents input for creating a line.
	PlayScriptLineCreationRequestInput struct {
		_             struct{}   `json:"-"`
		CastMemberID  string     `json:"castMemberId"`
		LineText      string     `json:"lineText"`
		Direction     string     `json:"direction,omitempty"`
		ScrollMode    ScrollMode `json:"scrollMode,omitempty"`
		LineIndex     int        `json:"lineIndex"`
		AutoScrollWpm int        `json:"autoScrollWpm,omitempty"`
	}

	// PlayScriptDatabaseCreationInput is used for creating a play script in persistence.
	PlayScriptDatabaseCreationInput struct {
		_                struct{} `json:"-"`
		PrecededByID     *string  `json:"-"`
		PrecedesID       *string  `json:"-"`
		ID               string   `json:"-"`
		Title            string   `json:"-"`
		Description      string   `json:"-"`
		Author           string   `json:"-"`
		CreatedByUser    string   `json:"-"`
		BelongsToAccount string   `json:"-"`
		IsPublic         bool     `json:"-"`
	}

	// PlayScriptCastMemberDatabaseCreationInput is used for creating a cast member in persistence.
	PlayScriptCastMemberDatabaseCreationInput struct {
		_            struct{} `json:"-"`
		ID           string   `json:"-"`
		PlayScriptID string   `json:"-"`
		Name         string   `json:"-"`
		Description  string   `json:"-"`
		SortOrder    int      `json:"-"`
	}

	// PlayScriptLineDatabaseCreationInput is used for creating a line in persistence.
	PlayScriptLineDatabaseCreationInput struct {
		_             struct{}   `json:"-"`
		ID            string     `json:"-"`
		PlayScriptID  string     `json:"-"`
		CastMemberID  string     `json:"-"`
		LineText      string     `json:"-"`
		Direction     string     `json:"-"`
		ScrollMode    ScrollMode `json:"-"`
		LineIndex     int        `json:"-"`
		AutoScrollWpm int        `json:"-"`
	}

	// PlayScriptUpdateRequestInput represents input for updating a play script.
	PlayScriptUpdateRequestInput struct {
		_            struct{} `json:"-"`
		Title        *string  `json:"title,omitempty"`
		Description  *string  `json:"description,omitempty"`
		Author       *string  `json:"author,omitempty"`
		IsPublic     *bool    `json:"isPublic,omitempty"`
		PrecededByID *string  `json:"precededById,omitempty"`
		PrecedesID   *string  `json:"precedesId,omitempty"`
	}

	// PlayScriptDataManager describes a structure capable of storing play scripts.
	PlayScriptDataManager interface {
		GetPlayScript(ctx context.Context, playScriptID string, includeLinkedScripts bool) (*PlayScript, error)
		GetPlayScripts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[PlayScript], error)
		GetPlayScriptsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[PlayScript], error)
		GetPublicPlayScripts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[PlayScript], error)
		CreatePlayScript(ctx context.Context, input *PlayScriptDatabaseCreationInput, castMembers []*PlayScriptCastMemberDatabaseCreationInput, lines []*PlayScriptLineDatabaseCreationInput) (*PlayScript, error)
		UpdatePlayScript(ctx context.Context, playScript *PlayScript) error
		UpdatePlayScriptLine(ctx context.Context, line *PlayScriptLine) error
		ArchivePlayScript(ctx context.Context, playScriptID string) error
	}
)

// Update merges a PlayScriptUpdateRequestInput into a PlayScript.
func (p *PlayScript) Update(input *PlayScriptUpdateRequestInput) {
	if input.Title != nil && *input.Title != p.Title {
		p.Title = *input.Title
	}
	if input.Description != nil && *input.Description != p.Description {
		p.Description = *input.Description
	}
	if input.Author != nil && *input.Author != p.Author {
		p.Author = *input.Author
	}
	if input.IsPublic != nil && *input.IsPublic != p.IsPublic {
		p.IsPublic = *input.IsPublic
	}
	if input.PrecededByID != nil {
		p.PrecededByID = input.PrecededByID
	}
	if input.PrecedesID != nil {
		p.PrecedesID = input.PrecedesID
	}
}

var _ validation.ValidatableWithContext = (*PlayScriptCreationRequestInput)(nil)

// ValidateWithContext validates a PlayScriptCreationRequestInput.
func (p *PlayScriptCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		p,
		validation.Field(&p.Title, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*PlayScriptDatabaseCreationInput)(nil)

// ValidateWithContext validates a PlayScriptDatabaseCreationInput.
func (p *PlayScriptDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		p,
		validation.Field(&p.ID, validation.Required),
		validation.Field(&p.Title, validation.Required),
		validation.Field(&p.CreatedByUser, validation.Required),
		validation.Field(&p.BelongsToAccount, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*PlayScriptUpdateRequestInput)(nil)

// ValidateWithContext validates a PlayScriptUpdateRequestInput.
func (p *PlayScriptUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		p,
		validation.Field(&p.Title, validation.When(p.Title != nil, validation.Required)),
	)
}

// GetSortedUniqueLineIndices returns the sorted unique LineIndex values from a play's lines.
// This defines "turns" - the frontend uses the same logic; backend must match for correct speaker resolution.
func GetSortedUniqueLineIndices(play *PlayScript) []int {
	seen := make(map[int]struct{})
	for _, l := range play.Lines {
		seen[l.LineIndex] = struct{}{}
	}
	indices := make([]int, 0, len(seen))
	for i := range seen {
		indices = append(indices, i)
	}
	sort.Ints(indices)
	return indices
}
