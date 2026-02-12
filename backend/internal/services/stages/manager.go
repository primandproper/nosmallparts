package stages

import (
	"context"

	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
)

var _ stagesdomain.StagesManager = (*manager)(nil)

type manager struct {
	repo stagesdomain.Repository
}

// NewManager returns a StagesManager that delegates to the given repository.
func NewManager(repo stagesdomain.Repository) stagesdomain.StagesManager {
	return &manager{repo: repo}
}

func (m *manager) CreateStage(ctx context.Context, input *stagesdomain.StageCreationInput) (*stagesdomain.Stage, error) {
	return m.repo.CreateStage(ctx, input)
}

func (m *manager) GetStageByID(ctx context.Context, stageID string) (*stagesdomain.Stage, error) {
	return m.repo.GetStageByID(ctx, stageID)
}

func (m *manager) GetStageByCode(ctx context.Context, code string) (*stagesdomain.Stage, error) {
	return m.repo.GetStageByCode(ctx, code)
}

func (m *manager) UpdateStage(ctx context.Context, stageID string, input *stagesdomain.StageUpdateInput) error {
	return m.repo.UpdateStage(ctx, stageID, input)
}

func (m *manager) ArchiveStage(ctx context.Context, stageID string) error {
	return m.repo.ArchiveStage(ctx, stageID)
}

func (m *manager) AddParticipant(ctx context.Context, input *stagesdomain.ParticipantCreationInput) (*stagesdomain.StageParticipant, error) {
	return m.repo.AddParticipant(ctx, input)
}

func (m *manager) GetParticipant(ctx context.Context, participantID string) (*stagesdomain.StageParticipant, error) {
	return m.repo.GetParticipant(ctx, participantID)
}

func (m *manager) GetActiveParticipantsForStage(ctx context.Context, stageID string) ([]*stagesdomain.StageParticipant, error) {
	return m.repo.GetActiveParticipantsForStage(ctx, stageID)
}

func (m *manager) UpdateParticipant(ctx context.Context, participantID string, input *stagesdomain.ParticipantUpdateInput) error {
	return m.repo.UpdateParticipant(ctx, participantID, input)
}

func (m *manager) SetParticipantLeft(ctx context.Context, participantID string) error {
	return m.repo.SetParticipantLeft(ctx, participantID)
}

func (m *manager) AddParticipantRole(ctx context.Context, participantID, castMemberID string) error {
	return m.repo.AddParticipantRole(ctx, participantID, castMemberID)
}

func (m *manager) RemoveParticipantRole(ctx context.Context, participantID, castMemberID string) error {
	return m.repo.RemoveParticipantRole(ctx, participantID, castMemberID)
}

func (m *manager) GetParticipantRoles(ctx context.Context, participantID string) ([]string, error) {
	return m.repo.GetParticipantRoles(ctx, participantID)
}
