// Package stages is the data behind a live stage: the stage itself, who is on
// it, which roles they hold, and who is banned from it.
//
// What happens on a stage, in real time, is the stage socket's
// (internal/services/stages/ws). This package keeps what has to outlive a
// connection, and a restart: a participant who drops and comes back finds
// their row and their roles where they left them, and the idle reaper finds a
// stage a restart orphaned.
package stages

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
)

// The refusals a caller branches on.
var (
	ErrStageNotFound       = platformerrors.New("stage not found")
	ErrParticipantNotFound = platformerrors.New("participant not found")
)

// Phase is where a stage is in a performance.
type Phase string

// The phases, in the order a performance moves through them.
const (
	PhaseLobby      Phase = "lobby"
	PhaseCountdown  Phase = "countdown"
	PhasePerforming Phase = "performing"
	PhaseDiscussion Phase = "discussion"
)

// codeAlphabet is what join codes are made of: upper case and digits, without
// the characters people misread for each other (0 and O, 1, I and L).
const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// CodeLength is how many characters a join code has.
const CodeLength = 6

type (
	// Identity is who somebody is across connections: a signed-in user, or a
	// guest by the ID their cookie carries. A zero Identity is nobody.
	Identity struct {
		_ struct{} `json:"-"`

		UserID  string
		GuestID string
	}

	// Stage is one live room.
	Stage struct {
		_ struct{} `json:"-"`

		CreatedAt    time.Time
		UpdatedAt    time.Time
		Host         Identity
		ID           string
		Code         string
		PlayScriptID string
		Phase        Phase
		LineIndex    int32
		PausedByHost bool
		DirectorMode bool
	}

	// Participant is somebody on a stage, connected or not.
	Participant struct {
		_ struct{} `json:"-"`

		CreatedAt     time.Time
		LeftAt        *time.Time
		Identity      Identity
		ID            string
		StageID       string
		DisplayName   string
		CastMemberIDs []string
		IsReady       bool
		IsAudience    bool
		ForceMuted    bool
	}
)

// IsZero reports whether the identity names nobody.
func (i Identity) IsZero() bool {
	return i.UserID == "" && i.GuestID == ""
}

// SignedIn reports whether the identity is a user rather than a guest.
func (i Identity) SignedIn() bool {
	return i.UserID != ""
}

// Matches reports whether two identities are the same person: the same user,
// or, for two guests, the same guest ID. Nobody matches nobody.
func (i Identity) Matches(other Identity) bool {
	if i.UserID != "" && other.UserID != "" {
		return i.UserID == other.UserID
	}

	if i.GuestID != "" && other.GuestID != "" {
		return i.GuestID == other.GuestID
	}

	return false
}

// HasHost reports whether anybody hosts the stage.
func (s *Stage) HasHost() bool {
	return !s.Host.IsZero()
}

// NewCode mints a join code from codeAlphabet.
func NewCode() string {
	b := make([]byte, CodeLength)
	for idx := range b {
		b[idx] = codeAlphabet[randomIndex(len(codeAlphabet))]
	}

	return string(b)
}

// randomIndex is a uniform index below n, from crypto/rand: a join code is
// what admits a stranger to a stage, so it should not be predictable from the
// one before it.
func randomIndex(n int) int {
	limit := 256 - 256%n

	var b [1]byte
	for {
		// crypto/rand.Read never returns an error; see its documentation.
		_, _ = rand.Read(b[:])

		if int(b[0]) < limit {
			return int(b[0]) % n
		}
	}
}

// Store is the stage data. Reads take an executor and writes a Tx, as the
// platform's stores do, so a caller doing several writes does them atomically.
type Store interface {
	GetStage(ctx context.Context, q database.SQLQueryExecutor, id string) (*Stage, error)
	GetStageByCode(ctx context.Context, q database.SQLQueryExecutor, code string) (*Stage, error)
	CodeTaken(ctx context.Context, q database.SQLQueryExecutor, code string) (bool, error)
	ActiveParticipants(ctx context.Context, q database.SQLQueryExecutor, stageID string) ([]*Participant, error)
	GetParticipant(ctx context.Context, q database.SQLQueryExecutor, id string) (*Participant, error)
	LatestParticipantFor(ctx context.Context, q database.SQLQueryExecutor, stageID string, who Identity) (*Participant, error)
	IsBanned(ctx context.Context, q database.SQLQueryExecutor, stageID string, who Identity, ip string) (bool, error)
	IdleStageIDs(ctx context.Context, q database.SQLQueryExecutor, cutoff time.Time) ([]string, error)

	CreateStage(ctx context.Context, tx database.Tx, stage *Stage) error
	DeleteStage(ctx context.Context, tx database.Tx, id string) error
	SetPlay(ctx context.Context, tx database.Tx, stageID, playScriptID string) error
	SetPhase(ctx context.Context, tx database.Tx, stageID string, phase Phase) error
	SetPhaseAndLine(ctx context.Context, tx database.Tx, stageID string, phase Phase, line int32) error
	SetLine(ctx context.Context, tx database.Tx, stageID string, line int32) error
	SetPaused(ctx context.Context, tx database.Tx, stageID string, paused bool) error
	SetDirectorMode(ctx context.Context, tx database.Tx, stageID string, enabled bool) error
	SetHost(ctx context.Context, tx database.Tx, stageID string, host Identity) error
	Reset(ctx context.Context, tx database.Tx, stageID string) error

	AddParticipant(ctx context.Context, tx database.Tx, participant *Participant) error
	SetLeftAt(ctx context.Context, tx database.Tx, participantID string, leftAt *time.Time) error
	Rename(ctx context.Context, tx database.Tx, participantID, name string) error
	SetReady(ctx context.Context, tx database.Tx, participantID string, ready bool) error
	SetAudience(ctx context.Context, tx database.Tx, participantID string, audience bool) error
	SetForceMuted(ctx context.Context, tx database.Tx, participantID string, muted bool) error
	ToggleRole(ctx context.Context, tx database.Tx, stageID, participantID, castMemberID string) error
	AssignRole(ctx context.Context, tx database.Tx, stageID, participantID, castMemberID string) error
	UnassignRole(ctx context.Context, tx database.Tx, participantID, castMemberID string) error
	ClearRoles(ctx context.Context, tx database.Tx, participantID string) error
	AddBan(ctx context.Context, tx database.Tx, stageID string, who Identity) error
	AddIPBan(ctx context.Context, tx database.Tx, stageID, ip string) error
}
