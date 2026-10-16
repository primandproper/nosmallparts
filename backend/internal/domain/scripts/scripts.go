// Package scripts is the script library: plays as their authors write them,
// their tags, and the ratings other people give them.
//
// Store is the data, with the platform's shape: a write takes a database.Tx
// and a read takes an executor. Service is the rules over it: who may change
// what, the range a rating may take, and that nobody rates their own work. The
// transport above decides who is calling and whether they are an admin, and
// hands Service a Viewer saying so.
package scripts

import (
	"context"
	"strings"

	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// The refusals a caller can branch on.
var (
	// ErrNotFound is a script that does not exist, or that the caller may not
	// see or change. The two are one answer on purpose: telling them apart
	// would confirm to a stranger that a private or disabled script exists.
	ErrNotFound = platformerrors.New("script not found")
	// ErrSelfRating is an author rating their own script.
	ErrSelfRating = platformerrors.New("authors can't rate their own work")
	// ErrRatingOutOfRange is a rating outside 1 to 5.
	ErrRatingOutOfRange = platformerrors.New("a rating is a whole number from 1 to 5")
	// ErrNotAdmin is a non-admin asking for an admin-only act.
	ErrNotAdmin = platformerrors.New("only an admin may do that")
	// ErrNotSignedIn is an anonymous caller asking for an act that needs a user.
	ErrNotSignedIn = platformerrors.New("sign in to do that")
	// ErrInvalidInput is a script that fails validation; the wrapped error
	// says which field.
	ErrInvalidInput = platformerrors.New("invalid script")
)

// The lowest and highest rating a person may give.
const (
	MinRating = 1
	MaxRating = 5
)

// ScrollMode is how the teleprompter moves through a line.
type ScrollMode string

// The scroll modes. Auto is the default for a line that names none.
const (
	ScrollModeAuto   ScrollMode = "auto"
	ScrollModeManual ScrollMode = "manual"
)

type (
	// Play is a script as stored, with the viewer's view of its ratings.
	Play struct {
		_ struct{} `json:"-"`

		// RatingStats is nil until somebody has rated the script.
		RatingStats *RatingStats
		// UserRating is the viewer's own rating, when they have given one.
		UserRating      *int32
		Author          *string
		ID              string
		Title           string
		Description     string
		CreatedByUserID string
		Cast            []*CastMember
		Lines           []*Line
		Tags            []*Tag
		IsPublic        bool
		Enabled         bool
	}

	// CastMember is one role in a play.
	CastMember struct {
		_ struct{} `json:"-"`

		Description *string
		ID          string
		Name        string
	}

	// Line is one utterance. More than one speaker is unison; several lines
	// sharing an Index are spoken over one another.
	Line struct {
		_ struct{} `json:"-"`

		Direction     *string
		AutoScrollWPM *int32
		Text          string
		ScrollMode    ScrollMode
		CastMemberIDs []string
		Index         int32
	}

	// Tag labels scripts; names are unique without regard to case.
	Tag struct {
		_ struct{} `json:"-"`

		Description *string
		ID          string
		Name        string
	}

	// RatingStats summarizes a script's ratings.
	RatingStats struct {
		_ struct{} `json:"-"`

		AverageRating float64
		RatingCount   int32
	}

	// Viewer is who is asking. A zero Viewer is anonymous.
	Viewer struct {
		_ struct{} `json:"-"`

		UserID string
		Admin  bool
	}
)

type (
	// PlayInput is a script as an author writes it, for a create or a full
	// replace.
	PlayInput struct {
		_ struct{} `json:"-"`

		Author      *string
		Title       string
		Description string
		Cast        []*CastMemberInput
		Lines       []*LineInput
		TagIDs      []string
		// TagNames are created when no tag has the name, matched without
		// regard to case or surrounding whitespace.
		TagNames []string
		IsPublic bool
	}

	// CastMemberInput is a role as an author writes it. ID is the author's own
	// handle for the role, which lines refer to it by; the stored member is
	// given an ID of its own.
	CastMemberInput struct {
		_ struct{} `json:"-"`

		Description *string
		ID          string
		Name        string
	}

	// LineInput is a line as an author writes it. CastMemberIDs name roles by
	// the IDs the input's cast carries.
	LineInput struct {
		_ struct{} `json:"-"`

		Direction     *string
		AutoScrollWPM *int32
		Text          string
		ScrollMode    ScrollMode
		CastMemberIDs []string
		Index         int32
	}
)

var (
	_ validation.ValidatableWithContext = (*PlayInput)(nil)
	_ validation.ValidatableWithContext = (*CastMemberInput)(nil)
	_ validation.ValidatableWithContext = (*LineInput)(nil)
)

// ValidateWithContext holds an input to the rules today's editor already
// enforces: a title, a name for every role, at least one speaker on every
// line, and a line index that is not negative.
func (p *PlayInput) ValidateWithContext(ctx context.Context) error {
	if err := validation.ValidateStructWithContext(ctx, p,
		validation.Field(&p.Title, validation.By(notBlank)),
		validation.Field(&p.Cast),
		validation.Field(&p.Lines),
	); err != nil {
		return platformerrors.Join(ErrInvalidInput, err)
	}

	return nil
}

// ValidateWithContext implements validation.ValidatableWithContext.
func (c *CastMemberInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.Name, validation.By(notBlank)),
	)
}

// ValidateWithContext implements validation.ValidatableWithContext.
func (l *LineInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, l,
		validation.Field(&l.CastMemberIDs, validation.Required, validation.Each(validation.Required)),
		validation.Field(&l.Index, validation.Min(int32(0))),
		validation.Field(&l.ScrollMode, validation.In(ScrollModeAuto, ScrollModeManual, ScrollMode(""))),
		validation.Field(&l.AutoScrollWPM, validation.NilOrNotEmpty, validation.Min(int32(1))),
	)
}

func notBlank(value any) error {
	if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
		return platformerrors.New("cannot be blank")
	}

	return nil
}

// NormalizeTagName is how a tag name is stored and matched: trimmed, with runs
// of whitespace collapsed to one space. Case is left alone for display and
// ignored for matching.
func NormalizeTagName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

// ListFilter narrows a listing.
type ListFilter struct {
	_ struct{} `json:"-"`

	// ViewerID sees their own scripts beside the public ones; empty is anonymous.
	ViewerID string
	// Search matches title, description, author, cast names and tag names,
	// without regard to case.
	Search string
	// All lists every script, private and disabled included, which is an
	// admin's view.
	All bool
}

// Store is the script library's data.
type Store interface {
	GetPlay(ctx context.Context, q database.SQLQueryExecutor, id string) (*Play, error)
	ListPlays(ctx context.Context, q database.SQLQueryExecutor, filter *ListFilter) ([]*Play, error)
	PlayOwner(ctx context.Context, q database.SQLQueryExecutor, id string) (string, error)
	CreatePlay(ctx context.Context, tx database.Tx, id, ownerID string, input *PlayInput) error
	ReplaceOwnedPlay(ctx context.Context, tx database.Tx, ownerID, id string, input *PlayInput) error
	DeletePlay(ctx context.Context, tx database.Tx, id string) error
	SetPlayEnabled(ctx context.Context, tx database.Tx, id string, enabled bool) error
	RatePlay(ctx context.Context, tx database.Tx, userID, id string, rating int32) (*RatingStats, error)
	UserRating(ctx context.Context, q database.SQLQueryExecutor, userID, id string) (rating int32, rated bool, err error)
	ListTags(ctx context.Context, q database.SQLQueryExecutor) ([]*Tag, error)
	EnsureTag(ctx context.Context, tx database.Tx, name string) (string, error)
}
