package stages

import (
	"context"
	"encoding/gob"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// StageCreatedServiceEventType indicates a stage was created.
	StageCreatedServiceEventType = "stage_created"
	// StageUpdatedServiceEventType indicates a stage was updated.
	StageUpdatedServiceEventType = "stage_updated"
	// StageArchivedServiceEventType indicates a stage was archived.
	StageArchivedServiceEventType = "stage_archived"
	// ParticipantJoinedServiceEventType indicates a participant joined.
	ParticipantJoinedServiceEventType = "participant_joined"
	// ParticipantLeftServiceEventType indicates a participant left.
	ParticipantLeftServiceEventType = "participant_left"
)

// StagePhase represents the current phase of a stage session.
type StagePhase string

const (
	// StagePhaseLobby is when participants are joining and configuring.
	StagePhaseLobby StagePhase = "lobby"
	// StagePhaseCountdown is the countdown before performance starts.
	StagePhaseCountdown StagePhase = "countdown"
	// StagePhasePerforming is the active performance phase.
	StagePhasePerforming StagePhase = "performing"
	// StagePhaseDiscussion is after the performance for discussion.
	StagePhaseDiscussion StagePhase = "discussion"
)

// ValidStagePhases returns all valid stage phase values.
func ValidStagePhases() []StagePhase {
	return []StagePhase{StagePhaseLobby, StagePhaseCountdown, StagePhasePerforming, StagePhaseDiscussion}
}

// IsValid checks if the stage phase is a valid value.
func (p StagePhase) IsValid() bool {
	switch p {
	case StagePhaseLobby, StagePhaseCountdown, StagePhasePerforming, StagePhaseDiscussion:
		return true
	}
	return false
}

func init() {
	gob.Register(new(Stage))
	gob.Register(new(StageParticipant))
	gob.Register(new(StageCreationInput))
	gob.Register(new(ParticipantCreationInput))
}

type (
	// Stage represents a real-time stage session where participants perform a play.
	Stage struct {
		_                struct{}            `json:"-"`
		CreatedAt        time.Time           `json:"createdAt"`
		LastUpdatedAt    *time.Time          `json:"lastUpdatedAt,omitempty"`
		ArchivedAt       *time.Time          `json:"archivedAt,omitempty"`
		PlayScriptID     *string             `json:"playScriptId,omitempty"`
		ID               string              `json:"id"`
		Code             string              `json:"code"`
		Phase            StagePhase          `json:"phase"`
		Participants     []*StageParticipant `json:"participants,omitempty"`
		CurrentLineIndex int                 `json:"currentLineIndex"`
	}

	// StageParticipant represents a participant in a stage session.
	StageParticipant struct {
		_             struct{}   `json:"-"`
		CreatedAt     time.Time  `json:"createdAt"`
		LeftAt        *time.Time `json:"leftAt,omitempty"`
		ID            string     `json:"id"`
		StageID       string     `json:"stageId"`
		UserID        *string    `json:"userId,omitempty"` // nil for anonymous
		DisplayName   string     `json:"name"`
		CastMemberIDs []string   `json:"castMemberIds"`
		IsDirector    bool       `json:"isDirector"`
		IsReady       bool       `json:"isReady"`
		IsAudience    bool       `json:"isAudience"`
	}

	// StageCreationInput is used for creating a new stage.
	StageCreationInput struct {
		_    struct{} `json:"-"`
		ID   string   `json:"-"`
		Code string   `json:"-"`
	}

	// ParticipantCreationInput is used for adding a participant to a stage.
	ParticipantCreationInput struct {
		_           struct{} `json:"-"`
		ID          string   `json:"-"`
		StageID     string   `json:"-"`
		UserID      *string  `json:"-"`
		DisplayName string   `json:"-"`
	}

	// StageUpdateInput is used for updating a stage.
	StageUpdateInput struct {
		_                struct{}    `json:"-"`
		PlayScriptID     *string     `json:"playScriptId,omitempty"`
		CurrentLineIndex *int        `json:"currentLineIndex,omitempty"`
		Phase            *StagePhase `json:"phase,omitempty"`
	}

	// ParticipantUpdateInput is used for updating a participant.
	ParticipantUpdateInput struct {
		_          struct{} `json:"-"`
		IsDirector *bool    `json:"isDirector,omitempty"`
		IsReady    *bool    `json:"isReady,omitempty"`
		IsAudience *bool    `json:"isAudience,omitempty"`
	}

	// StageDataManager describes a structure capable of storing stages.
	StageDataManager interface {
		CreateStage(ctx context.Context, input *StageCreationInput) (*Stage, error)
		GetStageByID(ctx context.Context, stageID string) (*Stage, error)
		GetStageByCode(ctx context.Context, code string) (*Stage, error)
		UpdateStage(ctx context.Context, stageID string, input *StageUpdateInput) error
		ArchiveStage(ctx context.Context, stageID string) error

		AddParticipant(ctx context.Context, input *ParticipantCreationInput) (*StageParticipant, error)
		GetParticipant(ctx context.Context, participantID string) (*StageParticipant, error)
		GetActiveParticipantsForStage(ctx context.Context, stageID string) ([]*StageParticipant, error)
		UpdateParticipant(ctx context.Context, participantID string, input *ParticipantUpdateInput) error
		SetParticipantLeft(ctx context.Context, participantID string) error

		// RejoinParticipant finds a left participant by stage+displayName, clears left_at, and returns the participant with roles.
		// Returns nil, sql.ErrNoRows if no matching left participant exists.
		RejoinParticipant(ctx context.Context, stageID, displayName string) (*StageParticipant, error)

		AddParticipantRole(ctx context.Context, participantID, castMemberID string) error
		RemoveParticipantRole(ctx context.Context, participantID, castMemberID string) error
		GetParticipantRoles(ctx context.Context, participantID string) ([]string, error)
	}
)

// Update merges a StageUpdateInput into a Stage.
func (s *Stage) Update(input *StageUpdateInput) {
	if input.PlayScriptID != nil {
		s.PlayScriptID = input.PlayScriptID
	}
	if input.CurrentLineIndex != nil {
		s.CurrentLineIndex = *input.CurrentLineIndex
	}
	if input.Phase != nil {
		s.Phase = *input.Phase
	}
}

// Update merges a ParticipantUpdateInput into a StageParticipant.
func (p *StageParticipant) Update(input *ParticipantUpdateInput) {
	if input.IsDirector != nil {
		p.IsDirector = *input.IsDirector
	}
	if input.IsReady != nil {
		p.IsReady = *input.IsReady
	}
	if input.IsAudience != nil {
		p.IsAudience = *input.IsAudience
	}
}

var _ validation.ValidatableWithContext = (*StageCreationInput)(nil)

// ValidateWithContext validates a StageCreationInput.
func (s *StageCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		s,
		validation.Field(&s.ID, validation.Required),
		validation.Field(&s.Code, validation.Required, validation.Length(6, 6)),
	)
}

var _ validation.ValidatableWithContext = (*ParticipantCreationInput)(nil)

// ValidateWithContext validates a ParticipantCreationInput.
func (p *ParticipantCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		p,
		validation.Field(&p.ID, validation.Required),
		validation.Field(&p.StageID, validation.Required),
		validation.Field(&p.DisplayName, validation.Required),
	)
}
