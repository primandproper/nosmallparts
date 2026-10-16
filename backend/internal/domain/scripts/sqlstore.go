package scripts

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"

	"github.com/primandproper/nosmallparts/backend/internal/domain/scripts/generated"

	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"

	"github.com/oklog/ulid/v2"
)

const storeName = "scripts_store"

// ErrNilExecutor is a store method handed no executor.
var ErrNilExecutor = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil executor for the scripts store")

// SQLStore is the SQLite implementation of Store.
//
// It takes a database.Client only for CurrentTime, which stamps the rows; its
// writes are handed a Tx and its reads an executor.
//
// A play is read with one query per child table, and a listing reads each play
// that way. A library holds dozens of scripts, not thousands, so the queries
// stay plain rather than batched.
type SQLStore struct {
	o11y    observability.Observer
	client  database.Client
	querier generated.Querier
}

var _ Store = (*SQLStore)(nil)

// Option configures a SQLStore.
type Option func(*SQLStore)

// WithPillars attaches the observability pillars.
func WithPillars(p *observability.Pillars) Option {
	return func(s *SQLStore) {
		if p != nil {
			s.o11y = observability.NewObserver(storeName, p.Logger, p.TracerProvider)
		}
	}
}

// NewSQLStore builds the store.
func NewSQLStore(client database.Client, opts ...Option) (*SQLStore, error) {
	if client == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client for the scripts store")
	}

	s := &SQLStore{
		o11y:    observability.NewObserver(storeName, nil, nil),
		client:  client,
		querier: generated.New(),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}

	return s, nil
}

// GetPlay implements Store.
func (s *SQLStore) GetPlay(ctx context.Context, q database.SQLQueryExecutor, id string) (*Play, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", id))
	defer op.End()

	if q == nil {
		return nil, op.Error(ErrNilExecutor, "reading a script")
	}

	play, err := s.hydrate(ctx, q, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, err
		}

		return nil, op.Error(err, "reading a script")
	}

	return play, nil
}

// ListPlays implements Store.
func (s *SQLStore) ListPlays(ctx context.Context, q database.SQLQueryExecutor, filter *ListFilter) ([]*Play, error) {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	if q == nil {
		return nil, op.Error(ErrNilExecutor, "listing scripts")
	}

	if filter == nil {
		filter = &ListFilter{}
	}

	var (
		ids []string
		err error
	)

	if filter.All {
		ids, err = s.querier.ListAllPlayScriptIDs(ctx, q)
	} else {
		ids, err = s.querier.ListVisiblePlayScriptIDs(ctx, q, filter.ViewerID)
	}

	if err != nil {
		return nil, op.Error(err, "listing script ids")
	}

	plays := make([]*Play, 0, len(ids))

	for _, id := range ids {
		play, hydrateErr := s.hydrate(ctx, q, id)
		if hydrateErr != nil {
			return nil, op.Error(hydrateErr, "reading listed script %s", id)
		}

		if matches(play, filter.Search) {
			plays = append(plays, play)
		}
	}

	return plays, nil
}

// matches reports whether a play answers a search, by the fields today's
// search covers.
func matches(play *Play, search string) bool {
	needle := strings.ToLower(strings.TrimSpace(search))
	if needle == "" {
		return true
	}

	haystack := []string{play.Title, play.Description}
	if play.Author != nil {
		haystack = append(haystack, *play.Author)
	}

	for _, member := range play.Cast {
		haystack = append(haystack, member.Name)
	}

	for _, tag := range play.Tags {
		haystack = append(haystack, tag.Name)
	}

	return slices.ContainsFunc(haystack, func(field string) bool {
		return strings.Contains(strings.ToLower(field), needle)
	})
}

// PlayOwner implements Store.
func (s *SQLStore) PlayOwner(ctx context.Context, q database.SQLQueryExecutor, id string) (string, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", id))
	defer op.End()

	if q == nil {
		return "", op.Error(ErrNilExecutor, "reading a script's owner")
	}

	row, err := s.querier.GetPlayScript(ctx, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}

		return "", op.Error(err, "reading a script's owner")
	}

	return row.CreatedByUser, nil
}

// CreatePlay implements Store.
func (s *SQLStore) CreatePlay(ctx context.Context, tx database.Tx, id, ownerID string, input *PlayInput) error {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("identity.user_id", ownerID))
	defer op.End()

	if tx == nil {
		return op.Error(ErrNilExecutor, "creating a script")
	}

	op.Set("play_script.id", id)

	now := s.client.CurrentTime().UTC()

	if err := s.querier.InsertPlayScript(ctx, tx, &generated.InsertPlayScriptParams{
		ID:            id,
		Title:         input.Title,
		Description:   input.Description,
		Author:        input.Author,
		IsPublic:      input.IsPublic,
		Enabled:       true,
		CreatedByUser: ownerID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}); err != nil {
		return op.Error(err, "inserting a script")
	}

	if err := s.writeChildren(ctx, tx, id, input); err != nil {
		return op.Error(err, "writing a new script's cast, lines and tags")
	}

	return nil
}

// ReplaceOwnedPlay implements Store. A script the owner does not own is
// ErrNotFound, the same as one that does not exist.
func (s *SQLStore) ReplaceOwnedPlay(ctx context.Context, tx database.Tx, ownerID, id string, input *PlayInput) error {
	ctx, op := s.o11y.Begin(ctx,
		observability.WithValue("identity.user_id", ownerID),
		observability.WithValue("play_script.id", id),
	)
	defer op.End()

	if tx == nil {
		return op.Error(ErrNilExecutor, "replacing a script")
	}

	updated, err := s.querier.UpdateOwnedPlayScript(ctx, tx, &generated.UpdateOwnedPlayScriptParams{
		Title:         input.Title,
		Description:   input.Description,
		Author:        input.Author,
		IsPublic:      input.IsPublic,
		UpdatedAt:     s.client.CurrentTime().UTC(),
		ID:            id,
		CreatedByUser: ownerID,
	})
	if err != nil {
		return op.Error(err, "updating a script")
	}

	if updated == 0 {
		return ErrNotFound
	}

	// Lines first: their speakers reference the cast members.
	if err = s.querier.DeleteLines(ctx, tx, id); err != nil {
		return op.Error(err, "clearing a script's lines")
	}

	if err = s.querier.DeleteCastMembers(ctx, tx, id); err != nil {
		return op.Error(err, "clearing a script's cast")
	}

	if err = s.querier.DeletePlayScriptTags(ctx, tx, id); err != nil {
		return op.Error(err, "clearing a script's tags")
	}

	if err = s.writeChildren(ctx, tx, id, input); err != nil {
		return op.Error(err, "writing a script's cast, lines and tags")
	}

	return nil
}

// writeChildren writes a script's cast, lines and tags. Every cast member is
// given a fresh ID, and a line's speakers are mapped from the IDs the input's
// cast carried to the ones just minted. A speaker naming no role in the input
// is refused rather than stored dangling.
func (s *SQLStore) writeChildren(ctx context.Context, tx database.Tx, playID string, input *PlayInput) error {
	castIDs := make(map[string]string, len(input.Cast))

	for position, member := range input.Cast {
		memberID := ulid.Make().String()

		handle := member.ID
		if handle == "" {
			handle = memberID
		}

		castIDs[handle] = memberID

		if err := s.querier.InsertCastMember(ctx, tx, &generated.InsertCastMemberParams{
			ID:           memberID,
			PlayScriptID: playID,
			Name:         member.Name,
			Description:  member.Description,
			Position:     int64(position),
		}); err != nil {
			return platformerrors.Wrapf(err, "inserting cast member %q", member.Name)
		}
	}

	for position, line := range input.Lines {
		lineID := ulid.Make().String()

		scrollMode := line.ScrollMode
		if scrollMode == "" {
			scrollMode = ScrollModeAuto
		}

		var wpm *int64
		if line.AutoScrollWPM != nil {
			wpm = new(int64(*line.AutoScrollWPM))
		}

		if err := s.querier.InsertLine(ctx, tx, &generated.InsertLineParams{
			ID:            lineID,
			PlayScriptID:  playID,
			LineText:      line.Text,
			LineIndex:     int64(line.Index),
			Position:      int64(position),
			Direction:     line.Direction,
			ScrollMode:    string(scrollMode),
			AutoScrollWpm: wpm,
		}); err != nil {
			return platformerrors.Wrapf(err, "inserting line %d", position)
		}

		seen := map[string]bool{}

		for speakerPosition, handle := range line.CastMemberIDs {
			memberID, ok := castIDs[handle]
			if !ok {
				return platformerrors.Join(ErrInvalidInput,
					platformerrors.Newf("line %d names cast member %q, which the cast does not include", position, handle))
			}

			if seen[memberID] {
				continue
			}

			seen[memberID] = true

			if err := s.querier.InsertLineSpeaker(ctx, tx, &generated.InsertLineSpeakerParams{
				LineID:       lineID,
				CastMemberID: memberID,
				Position:     int64(speakerPosition),
			}); err != nil {
				return platformerrors.Wrapf(err, "inserting a speaker for line %d", position)
			}
		}
	}

	for _, tagID := range input.TagIDs {
		if err := s.querier.InsertPlayScriptTag(ctx, tx, &generated.InsertPlayScriptTagParams{
			PlayScriptID: playID,
			ScriptTagID:  tagID,
		}); err != nil {
			return platformerrors.Wrapf(err, "tagging the script with %s", tagID)
		}
	}

	return nil
}

// DeletePlay implements Store.
func (s *SQLStore) DeletePlay(ctx context.Context, tx database.Tx, id string) error {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", id))
	defer op.End()

	if tx == nil {
		return op.Error(ErrNilExecutor, "deleting a script")
	}

	deleted, err := s.querier.DeletePlayScript(ctx, tx, id)
	if err != nil {
		return op.Error(err, "deleting a script")
	}

	if deleted == 0 {
		return ErrNotFound
	}

	return nil
}

// SetPlayEnabled implements Store.
func (s *SQLStore) SetPlayEnabled(ctx context.Context, tx database.Tx, id string, enabled bool) error {
	ctx, op := s.o11y.Begin(ctx,
		observability.WithValue("play_script.id", id),
		observability.WithValue("play_script.enabled", enabled),
	)
	defer op.End()

	if tx == nil {
		return op.Error(ErrNilExecutor, "switching a script")
	}

	updated, err := s.querier.SetPlayScriptEnabled(ctx, tx, &generated.SetPlayScriptEnabledParams{
		Enabled:   enabled,
		UpdatedAt: s.client.CurrentTime().UTC(),
		ID:        id,
	})
	if err != nil {
		return op.Error(err, "switching a script")
	}

	if updated == 0 {
		return ErrNotFound
	}

	return nil
}

// RatePlay implements Store. A second rating by the same user replaces the
// first.
func (s *SQLStore) RatePlay(ctx context.Context, tx database.Tx, userID, id string, rating int32) (*RatingStats, error) {
	ctx, op := s.o11y.Begin(ctx,
		observability.WithValue("identity.user_id", userID),
		observability.WithValue("play_script.id", id),
	)
	defer op.End()

	if tx == nil {
		return nil, op.Error(ErrNilExecutor, "rating a script")
	}

	if err := s.querier.UpsertRating(ctx, tx, &generated.UpsertRatingParams{
		PlayScriptID: id,
		UserID:       userID,
		Rating:       int64(rating),
		RatedAt:      s.client.CurrentTime().UTC(),
	}); err != nil {
		return nil, op.Error(err, "recording a rating")
	}

	stats, err := s.ratingStats(ctx, tx, id)
	if err != nil {
		return nil, op.Error(err, "reading a script's ratings")
	}

	return stats, nil
}

// UserRating implements Store. rated is false for a user who has not rated
// the script.
func (s *SQLStore) UserRating(ctx context.Context, q database.SQLQueryExecutor, userID, id string) (rating int32, rated bool, err error) {
	ctx, op := s.o11y.Begin(ctx,
		observability.WithValue("identity.user_id", userID),
		observability.WithValue("play_script.id", id),
	)
	defer op.End()

	if q == nil {
		return 0, false, op.Error(ErrNilExecutor, "reading a user's rating")
	}

	stored, err := s.querier.GetUserRating(ctx, q, &generated.GetUserRatingParams{PlayScriptID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}

		return 0, false, op.Error(err, "reading a user's rating")
	}

	return int32(stored), true, nil
}

// ListTags implements Store.
func (s *SQLStore) ListTags(ctx context.Context, q database.SQLQueryExecutor) ([]*Tag, error) {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	if q == nil {
		return nil, op.Error(ErrNilExecutor, "listing tags")
	}

	rows, err := s.querier.ListScriptTags(ctx, q)
	if err != nil {
		return nil, op.Error(err, "listing tags")
	}

	tags := make([]*Tag, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, tagFromRow(row))
	}

	return tags, nil
}

// EnsureTag implements Store: the ID of the tag with this name, created if no
// tag has it. The name is normalized first, and matched without regard to
// case.
func (s *SQLStore) EnsureTag(ctx context.Context, tx database.Tx, name string) (string, error) {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	if tx == nil {
		return "", op.Error(ErrNilExecutor, "ensuring a tag")
	}

	name = NormalizeTagName(name)
	op.Set("script_tag.name", name)

	existing, err := s.querier.GetScriptTagByName(ctx, tx, name)
	if err == nil {
		return existing.ID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return "", op.Error(err, "looking up a tag")
	}

	id := ulid.Make().String()

	if err = s.querier.InsertScriptTag(ctx, tx, &generated.InsertScriptTagParams{
		ID:        id,
		Name:      name,
		CreatedAt: s.client.CurrentTime().UTC(),
	}); err != nil {
		return "", op.Error(err, "creating a tag")
	}

	return id, nil
}

// hydrate reads one play whole. A play that does not exist is ErrNotFound.
func (s *SQLStore) hydrate(ctx context.Context, q database.SQLQueryExecutor, id string) (*Play, error) {
	row, err := s.querier.GetPlayScript(ctx, q, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, platformerrors.Wrap(err, "reading the script row")
	}

	play := &Play{
		ID:              row.ID,
		Title:           row.Title,
		Description:     row.Description,
		Author:          row.Author,
		IsPublic:        row.IsPublic,
		Enabled:         row.Enabled,
		CreatedByUserID: row.CreatedByUser,
	}

	cast, err := s.querier.ListCastMembers(ctx, q, id)
	if err != nil {
		return nil, platformerrors.Wrap(err, "reading the cast")
	}

	for _, member := range cast {
		play.Cast = append(play.Cast, &CastMember{ID: member.ID, Name: member.Name, Description: member.Description})
	}

	speakers, err := s.querier.ListLineSpeakers(ctx, q, id)
	if err != nil {
		return nil, platformerrors.Wrap(err, "reading the speakers")
	}

	speakersByLine := map[string][]string{}
	for _, speaker := range speakers {
		speakersByLine[speaker.LineID] = append(speakersByLine[speaker.LineID], speaker.CastMemberID)
	}

	lines, err := s.querier.ListLines(ctx, q, id)
	if err != nil {
		return nil, platformerrors.Wrap(err, "reading the lines")
	}

	for _, line := range lines {
		var wpm *int32
		if line.AutoScrollWpm != nil {
			wpm = new(int32(*line.AutoScrollWpm))
		}

		play.Lines = append(play.Lines, &Line{
			CastMemberIDs: speakersByLine[line.ID],
			Text:          line.LineText,
			Index:         int32(line.LineIndex),
			Direction:     line.Direction,
			ScrollMode:    ScrollMode(line.ScrollMode),
			AutoScrollWPM: wpm,
		})
	}

	tags, err := s.querier.ListPlayScriptTags(ctx, q, id)
	if err != nil {
		return nil, platformerrors.Wrap(err, "reading the tags")
	}

	for _, tag := range tags {
		play.Tags = append(play.Tags, tagFromRow(tag))
	}

	stats, err := s.ratingStats(ctx, q, id)
	if err != nil {
		return nil, err
	}

	if stats.RatingCount > 0 {
		play.RatingStats = stats
	}

	return play, nil
}

// ratingStats counts a script's ratings, which may be none.
func (s *SQLStore) ratingStats(ctx context.Context, q database.SQLQueryExecutor, id string) (*RatingStats, error) {
	stats, err := s.querier.GetRatingStats(ctx, q, id)
	if err != nil {
		return nil, platformerrors.Wrap(err, "reading the ratings")
	}

	return &RatingStats{AverageRating: stats.AverageRating, RatingCount: int32(stats.RatingCount)}, nil
}

func tagFromRow(row *generated.ScriptTags) *Tag {
	return &Tag{ID: row.ID, Name: row.Name, Description: row.Description}
}
