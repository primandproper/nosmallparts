package stages

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/domain/stages/generated"

	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"

	"github.com/oklog/ulid/v2"
)

const storeName = "stages_store"

// ErrNilExecutor is a store method handed no executor.
var ErrNilExecutor = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil executor for the stages store")

// SQLStore is the SQLite implementation of Store. It holds a database.Client
// for CurrentTime, which stamps the rows.
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
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client for the stages store")
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

func (s *SQLStore) now() time.Time {
	return s.client.CurrentTime().UTC()
}

// write runs one write, refusing a nil Tx and logging a failure with what was
// being done.
func (s *SQLStore) write(ctx context.Context, tx database.Tx, description string, fn func(context.Context) error) error {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	if tx == nil {
		return op.Error(ErrNilExecutor, "%s", description)
	}

	if err := fn(ctx); err != nil {
		return op.Error(err, "%s", description)
	}

	return nil
}

// GetStage implements Store.
func (s *SQLStore) GetStage(ctx context.Context, q database.SQLQueryExecutor, id string) (*Stage, error) {
	if q == nil {
		return nil, ErrNilExecutor
	}

	row, err := s.querier.GetStageByID(ctx, q, id)

	return s.stageFromRow(ctx, row, err, "reading a stage")
}

// GetStageByCode implements Store. Codes are stored upper case; the caller
// folds what a person typed.
func (s *SQLStore) GetStageByCode(ctx context.Context, q database.SQLQueryExecutor, code string) (*Stage, error) {
	if q == nil {
		return nil, ErrNilExecutor
	}

	row, err := s.querier.GetStageByCode(ctx, q, code)

	return s.stageFromRow(ctx, row, err, "reading a stage by code")
}

func (s *SQLStore) stageFromRow(ctx context.Context, row *generated.Stages, err error, description string) (*Stage, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStageNotFound
		}

		_, op := s.o11y.Begin(ctx)
		defer op.End()

		return nil, op.Error(err, "%s", description)
	}

	return &Stage{
		ID:           row.ID,
		Code:         row.Code,
		PlayScriptID: deref(row.PlayScriptID),
		LineIndex:    int32(row.CurrentLineIndex),
		Phase:        Phase(row.Phase),
		Host:         Identity{UserID: deref(row.HostUserID), GuestID: deref(row.HostGuestID)},
		PausedByHost: row.PausedByHost,
		DirectorMode: row.DirectorMode,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

// CodeTaken implements Store.
func (s *SQLStore) CodeTaken(ctx context.Context, q database.SQLQueryExecutor, code string) (bool, error) {
	if q == nil {
		return false, ErrNilExecutor
	}

	taken, err := s.querier.StageCodeTaken(ctx, q, code)
	if err != nil {
		return false, platformerrors.Wrap(err, "checking a stage code")
	}

	return taken != 0, nil
}

// ActiveParticipants implements Store: the participants who have not left, in
// the order they joined, with the roles each holds.
func (s *SQLStore) ActiveParticipants(ctx context.Context, q database.SQLQueryExecutor, stageID string) ([]*Participant, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("stage.id", stageID))
	defer op.End()

	if q == nil {
		return nil, op.Error(ErrNilExecutor, "listing participants")
	}

	rows, err := s.querier.ListActiveParticipants(ctx, q, stageID)
	if err != nil {
		return nil, op.Error(err, "listing participants")
	}

	roles, err := s.querier.ListStageRoles(ctx, q, stageID)
	if err != nil {
		return nil, op.Error(err, "listing roles")
	}

	held := map[string][]string{}
	for _, role := range roles {
		held[role.StageParticipantID] = append(held[role.StageParticipantID], role.CastMemberID)
	}

	participants := make([]*Participant, 0, len(rows))
	for _, row := range rows {
		participant := participantFromRow(row)
		participant.CastMemberIDs = held[row.ID]
		participants = append(participants, participant)
	}

	return participants, nil
}

// GetParticipant implements Store, roles included.
func (s *SQLStore) GetParticipant(ctx context.Context, q database.SQLQueryExecutor, id string) (*Participant, error) {
	if q == nil {
		return nil, ErrNilExecutor
	}

	row, err := s.querier.GetParticipant(ctx, q, id)

	return s.participantWithRoles(ctx, q, row, err)
}

// LatestParticipantFor implements Store: the most recent row this identity has
// on the stage, whether or not they are connected, so a reconnect is
// re-admitted with the roles it held.
func (s *SQLStore) LatestParticipantFor(ctx context.Context, q database.SQLQueryExecutor, stageID string, who Identity) (*Participant, error) {
	if q == nil {
		return nil, ErrNilExecutor
	}

	if who.UserID != "" {
		row, err := s.querier.FindLatestParticipantByUser(ctx, q, &generated.FindLatestParticipantByUserParams{StageID: stageID, UserID: &who.UserID})
		if !errors.Is(err, sql.ErrNoRows) {
			return s.participantWithRoles(ctx, q, row, err)
		}
	}

	if who.GuestID != "" {
		row, err := s.querier.FindLatestParticipantByGuest(ctx, q, &generated.FindLatestParticipantByGuestParams{StageID: stageID, GuestID: &who.GuestID})

		return s.participantWithRoles(ctx, q, row, err)
	}

	return nil, ErrParticipantNotFound
}

func (s *SQLStore) participantWithRoles(ctx context.Context, q database.SQLQueryExecutor, row *generated.StageParticipants, err error) (*Participant, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrParticipantNotFound
		}

		return nil, platformerrors.Wrap(err, "reading a participant")
	}

	participant := participantFromRow(row)

	if participant.CastMemberIDs, err = s.querier.ListParticipantRoles(ctx, q, row.ID); err != nil {
		return nil, platformerrors.Wrap(err, "reading a participant's roles")
	}

	return participant, nil
}

func participantFromRow(row *generated.StageParticipants) *Participant {
	return &Participant{
		ID:          row.ID,
		StageID:     row.StageID,
		Identity:    Identity{UserID: deref(row.UserID), GuestID: deref(row.GuestID)},
		DisplayName: row.DisplayName,
		IsReady:     row.IsReady,
		IsAudience:  row.IsAudience,
		ForceMuted:  row.ForceMutedByHost,
		CreatedAt:   row.CreatedAt,
		LeftAt:      row.LeftAt,
	}
}

// IsBanned implements Store: banned by identity, or by the address they are
// connecting from.
func (s *SQLStore) IsBanned(ctx context.Context, q database.SQLQueryExecutor, stageID string, who Identity, ip string) (bool, error) {
	if q == nil {
		return false, ErrNilExecutor
	}

	if who.UserID != "" {
		n, err := s.querier.CountUserBans(ctx, q, &generated.CountUserBansParams{StageID: stageID, UserID: &who.UserID})
		if err != nil || n > 0 {
			return n > 0, platformerrors.Wrap(err, "checking user bans")
		}
	}

	if who.GuestID != "" {
		n, err := s.querier.CountGuestBans(ctx, q, &generated.CountGuestBansParams{StageID: stageID, GuestID: &who.GuestID})
		if err != nil || n > 0 {
			return n > 0, platformerrors.Wrap(err, "checking guest bans")
		}
	}

	if ip != "" {
		n, err := s.querier.CountIPBans(ctx, q, &generated.CountIPBansParams{StageID: stageID, Ip: ip})
		if err != nil || n > 0 {
			return n > 0, platformerrors.Wrap(err, "checking address bans")
		}
	}

	return false, nil
}

// IdleStageIDs implements Store.
func (s *SQLStore) IdleStageIDs(ctx context.Context, q database.SQLQueryExecutor, cutoff time.Time) ([]string, error) {
	if q == nil {
		return nil, ErrNilExecutor
	}

	ids, err := s.querier.ListIdleStageIDs(ctx, q, cutoff.UTC())
	if err != nil {
		return nil, platformerrors.Wrap(err, "listing idle stages")
	}

	return ids, nil
}

// CreateStage implements Store. The stage starts in the lobby with nobody
// hosting; the caller supplies the ID and code.
func (s *SQLStore) CreateStage(ctx context.Context, tx database.Tx, stage *Stage) error {
	return s.write(ctx, tx, "creating a stage", func(ctx context.Context) error {
		now := s.now()

		return s.querier.InsertStage(ctx, tx, &generated.InsertStageParams{
			ID:           stage.ID,
			Code:         stage.Code,
			PlayScriptID: nonEmpty(stage.PlayScriptID),
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	})
}

// DeleteStage implements Store. Participants, roles and bans go with it.
func (s *SQLStore) DeleteStage(ctx context.Context, tx database.Tx, id string) error {
	return s.write(ctx, tx, "deleting a stage", func(ctx context.Context) error {
		return s.querier.DeleteStage(ctx, tx, id)
	})
}

// SetPlay implements Store. Selecting a play clears every role on the stage,
// since the roles belonged to the play that was selected before.
func (s *SQLStore) SetPlay(ctx context.Context, tx database.Tx, stageID, playScriptID string) error {
	return s.write(ctx, tx, "selecting a play", func(ctx context.Context) error {
		if err := s.querier.SetStagePlay(ctx, tx, &generated.SetStagePlayParams{
			PlayScriptID: nonEmpty(playScriptID),
			UpdatedAt:    s.now(),
			ID:           stageID,
		}); err != nil {
			return err
		}

		return s.querier.DeleteStageRoles(ctx, tx, stageID)
	})
}

// SetPhase implements Store.
func (s *SQLStore) SetPhase(ctx context.Context, tx database.Tx, stageID string, phase Phase) error {
	return s.write(ctx, tx, "setting a stage's phase", func(ctx context.Context) error {
		return s.querier.SetStagePhase(ctx, tx, &generated.SetStagePhaseParams{Phase: string(phase), UpdatedAt: s.now(), ID: stageID})
	})
}

// SetPhaseAndLine implements Store.
func (s *SQLStore) SetPhaseAndLine(ctx context.Context, tx database.Tx, stageID string, phase Phase, line int32) error {
	return s.write(ctx, tx, "setting a stage's phase and line", func(ctx context.Context) error {
		return s.querier.SetStagePhaseAndLine(ctx, tx, &generated.SetStagePhaseAndLineParams{
			Phase:            string(phase),
			CurrentLineIndex: int64(line),
			UpdatedAt:        s.now(),
			ID:               stageID,
		})
	})
}

// SetLine implements Store.
func (s *SQLStore) SetLine(ctx context.Context, tx database.Tx, stageID string, line int32) error {
	return s.write(ctx, tx, "moving a stage's playhead", func(ctx context.Context) error {
		return s.querier.SetStageLine(ctx, tx, &generated.SetStageLineParams{CurrentLineIndex: int64(line), UpdatedAt: s.now(), ID: stageID})
	})
}

// SetPaused implements Store.
func (s *SQLStore) SetPaused(ctx context.Context, tx database.Tx, stageID string, paused bool) error {
	return s.write(ctx, tx, "pausing a stage", func(ctx context.Context) error {
		return s.querier.SetStagePaused(ctx, tx, &generated.SetStagePausedParams{PausedByHost: paused, UpdatedAt: s.now(), ID: stageID})
	})
}

// SetDirectorMode implements Store.
func (s *SQLStore) SetDirectorMode(ctx context.Context, tx database.Tx, stageID string, enabled bool) error {
	return s.write(ctx, tx, "switching director mode", func(ctx context.Context) error {
		return s.querier.SetStageDirectorMode(ctx, tx, &generated.SetStageDirectorModeParams{DirectorMode: enabled, UpdatedAt: s.now(), ID: stageID})
	})
}

// SetHost implements Store. A zero identity clears the host.
func (s *SQLStore) SetHost(ctx context.Context, tx database.Tx, stageID string, host Identity) error {
	return s.write(ctx, tx, "setting a stage's host", func(ctx context.Context) error {
		return s.querier.SetStageHost(ctx, tx, &generated.SetStageHostParams{
			HostUserID:  nonEmpty(host.UserID),
			HostGuestID: nonEmpty(host.GuestID),
			UpdatedAt:   s.now(),
			ID:          stageID,
		})
	})
}

// Reset implements Store: back to the lobby at the first line, with nobody
// ready. Roles are kept, so the same cast can perform again.
func (s *SQLStore) Reset(ctx context.Context, tx database.Tx, stageID string) error {
	return s.write(ctx, tx, "resetting a stage", func(ctx context.Context) error {
		if err := s.querier.ResetStage(ctx, tx, &generated.ResetStageParams{UpdatedAt: s.now(), ID: stageID}); err != nil {
			return err
		}

		return s.querier.ClearStageReadiness(ctx, tx, stageID)
	})
}

// AddParticipant implements Store. The caller supplies the ID.
func (s *SQLStore) AddParticipant(ctx context.Context, tx database.Tx, participant *Participant) error {
	return s.write(ctx, tx, "adding a participant", func(ctx context.Context) error {
		return s.querier.InsertParticipant(ctx, tx, &generated.InsertParticipantParams{
			ID:          participant.ID,
			StageID:     participant.StageID,
			UserID:      nonEmpty(participant.Identity.UserID),
			GuestID:     nonEmpty(participant.Identity.GuestID),
			DisplayName: participant.DisplayName,
			IsAudience:  participant.IsAudience,
			CreatedAt:   s.now(),
		})
	})
}

// SetLeftAt implements Store. A nil time re-admits a participant who left.
func (s *SQLStore) SetLeftAt(ctx context.Context, tx database.Tx, participantID string, leftAt *time.Time) error {
	return s.write(ctx, tx, "recording a participant leaving or returning", func(ctx context.Context) error {
		if leftAt != nil {
			leftAt = new(leftAt.UTC())
		}

		return s.querier.SetParticipantLeftAt(ctx, tx, &generated.SetParticipantLeftAtParams{LeftAt: leftAt, ID: participantID})
	})
}

// Rename implements Store.
func (s *SQLStore) Rename(ctx context.Context, tx database.Tx, participantID, name string) error {
	return s.write(ctx, tx, "renaming a participant", func(ctx context.Context) error {
		return s.querier.SetParticipantDisplayName(ctx, tx, &generated.SetParticipantDisplayNameParams{DisplayName: name, ID: participantID})
	})
}

// SetReady implements Store.
func (s *SQLStore) SetReady(ctx context.Context, tx database.Tx, participantID string, ready bool) error {
	return s.write(ctx, tx, "setting a participant's readiness", func(ctx context.Context) error {
		return s.querier.SetParticipantReady(ctx, tx, &generated.SetParticipantReadyParams{IsReady: ready, ID: participantID})
	})
}

// SetAudience implements Store. Moving to the audience gives up every role.
func (s *SQLStore) SetAudience(ctx context.Context, tx database.Tx, participantID string, audience bool) error {
	return s.write(ctx, tx, "moving a participant to or from the audience", func(ctx context.Context) error {
		if err := s.querier.SetParticipantAudience(ctx, tx, &generated.SetParticipantAudienceParams{IsAudience: audience, ID: participantID}); err != nil {
			return err
		}

		if !audience {
			return nil
		}

		return s.querier.DeleteParticipantRoles(ctx, tx, participantID)
	})
}

// SetForceMuted implements Store.
func (s *SQLStore) SetForceMuted(ctx context.Context, tx database.Tx, participantID string, muted bool) error {
	return s.write(ctx, tx, "muting a participant", func(ctx context.Context) error {
		return s.querier.SetParticipantForceMuted(ctx, tx, &generated.SetParticipantForceMutedParams{ForceMutedByHost: muted, ID: participantID})
	})
}

// ToggleRole implements Store: a participant holding the role gives it up, and
// one who does not takes it from whoever on the stage held it.
func (s *SQLStore) ToggleRole(ctx context.Context, tx database.Tx, stageID, participantID, castMemberID string) error {
	return s.write(ctx, tx, "toggling a role", func(ctx context.Context) error {
		released, err := s.querier.DeleteParticipantRole(ctx, tx, &generated.DeleteParticipantRoleParams{
			StageParticipantID: participantID,
			CastMemberID:       castMemberID,
		})
		if err != nil || released > 0 {
			return err
		}

		return s.assign(ctx, tx, stageID, participantID, castMemberID)
	})
}

// AssignRole implements Store: the participant takes the role from whoever on
// the stage held it.
func (s *SQLStore) AssignRole(ctx context.Context, tx database.Tx, stageID, participantID, castMemberID string) error {
	return s.write(ctx, tx, "assigning a role", func(ctx context.Context) error {
		return s.assign(ctx, tx, stageID, participantID, castMemberID)
	})
}

func (s *SQLStore) assign(ctx context.Context, tx database.Tx, stageID, participantID, castMemberID string) error {
	if err := s.querier.DeleteRoleOnStage(ctx, tx, &generated.DeleteRoleOnStageParams{CastMemberID: castMemberID, StageID: stageID}); err != nil {
		return err
	}

	return s.querier.InsertParticipantRole(ctx, tx, &generated.InsertParticipantRoleParams{
		StageParticipantID: participantID,
		CastMemberID:       castMemberID,
	})
}

// UnassignRole implements Store.
func (s *SQLStore) UnassignRole(ctx context.Context, tx database.Tx, participantID, castMemberID string) error {
	return s.write(ctx, tx, "unassigning a role", func(ctx context.Context) error {
		_, err := s.querier.DeleteParticipantRole(ctx, tx, &generated.DeleteParticipantRoleParams{
			StageParticipantID: participantID,
			CastMemberID:       castMemberID,
		})

		return err
	})
}

// ClearRoles implements Store.
func (s *SQLStore) ClearRoles(ctx context.Context, tx database.Tx, participantID string) error {
	return s.write(ctx, tx, "clearing a participant's roles", func(ctx context.Context) error {
		return s.querier.DeleteParticipantRoles(ctx, tx, participantID)
	})
}

// AddBan implements Store. A zero identity bans nobody.
func (s *SQLStore) AddBan(ctx context.Context, tx database.Tx, stageID string, who Identity) error {
	if who.IsZero() {
		return nil
	}

	return s.write(ctx, tx, "banning an identity", func(ctx context.Context) error {
		return s.querier.InsertBan(ctx, tx, &generated.InsertBanParams{
			ID:       ulid.Make().String(),
			StageID:  stageID,
			UserID:   nonEmpty(who.UserID),
			GuestID:  nonEmpty(who.GuestID),
			BannedAt: s.now(),
		})
	})
}

// AddIPBan implements Store. An empty address bans nobody.
func (s *SQLStore) AddIPBan(ctx context.Context, tx database.Tx, stageID, ip string) error {
	if ip == "" {
		return nil
	}

	return s.write(ctx, tx, "banning an address", func(ctx context.Context) error {
		return s.querier.InsertIPBan(ctx, tx, &generated.InsertIPBanParams{ID: ulid.Make().String(), StageID: stageID, Ip: ip, BannedAt: s.now()})
	})
}

func deref(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func nonEmpty(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
