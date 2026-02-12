package stages

import "context"

// StagesManager describes operations for managing stages (service-level interface).
// Implementations may wrap a Repository and add orchestration or validation.
type StagesManager interface {
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

	AddParticipantRole(ctx context.Context, participantID, castMemberID string) error
	RemoveParticipantRole(ctx context.Context, participantID, castMemberID string) error
	GetParticipantRoles(ctx context.Context, participantID string) ([]string, error)
}
