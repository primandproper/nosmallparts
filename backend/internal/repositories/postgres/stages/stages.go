package stages

import (
	"context"
	"database/sql"

	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	generated "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/stages/generated"
)

var (
	_ types.StageDataManager = (*repository)(nil)
)

// CreateStage creates a stage in the database.
func (r *repository) CreateStage(ctx context.Context, input *types.StageCreationInput) (*types.Stage, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, database.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, keys.StageIDKey, input.ID)
	logger = logger.WithValue(keys.StageIDKey, input.ID)

	if err := r.generatedQuerier.CreateStage(ctx, r.db, &generated.CreateStageParams{
		ID:   input.ID,
		Code: input.Code,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating stage")
	}

	stage := &types.Stage{
		ID:               input.ID,
		Code:             input.Code,
		Phase:            types.StagePhaseLobby,
		CurrentLineIndex: 0,
		CreatedAt:        r.CurrentTime(),
		Participants:     []*types.StageParticipant{},
	}

	return stage, nil
}

// GetStageByID fetches a stage by ID from the database.
func (r *repository) GetStageByID(ctx context.Context, stageID string) (*types.Stage, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if stageID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.StageIDKey, stageID)
	tracing.AttachToSpan(span, keys.StageIDKey, stageID)

	result, err := r.generatedQuerier.GetStageByID(ctx, r.db, stageID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching stage by ID")
	}

	stage := r.convertStage(result)

	// Fetch participants
	participants, err := r.GetActiveParticipantsForStage(ctx, stageID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching stage participants")
	}
	stage.Participants = participants

	return stage, nil
}

// GetStageByCode fetches a stage by join code from the database.
func (r *repository) GetStageByCode(ctx context.Context, code string) (*types.Stage, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if code == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue("code", code)
	tracing.AttachToSpan(span, "code", code)

	result, err := r.generatedQuerier.GetStageByCode(ctx, r.db, code)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching stage by code")
	}

	stage := r.convertStage(result)

	// Fetch participants
	participants, err := r.GetActiveParticipantsForStage(ctx, stage.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching stage participants")
	}
	stage.Participants = participants

	return stage, nil
}

// UpdateStage updates a stage in the database.
func (r *repository) UpdateStage(ctx context.Context, stageID string, input *types.StageUpdateInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if stageID == "" {
		return database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.StageIDKey, stageID)
	tracing.AttachToSpan(span, keys.StageIDKey, stageID)

	params := &generated.UpdateStageParams{
		ID: stageID,
	}

	if input.PlayScriptID != nil {
		params.PlayScriptID = database.NullStringFromStringPointer(input.PlayScriptID)
	}
	if input.CurrentLineIndex != nil {
		params.CurrentLineIndex = sql.NullInt32{Int32: int32(*input.CurrentLineIndex), Valid: true}
	}
	if input.Phase != nil {
		params.Phase = sql.NullString{String: string(*input.Phase), Valid: true}
	}

	rowsAffected, err := r.generatedQuerier.UpdateStage(ctx, r.db, params)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating stage")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ArchiveStage archives a stage in the database.
func (r *repository) ArchiveStage(ctx context.Context, stageID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if stageID == "" {
		return database.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, keys.StageIDKey, stageID)

	logger := r.logger.WithValue(keys.StageIDKey, stageID)

	rowsAffected, err := r.generatedQuerier.ArchiveStage(ctx, r.db, stageID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving stage")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// AddParticipant adds a participant to a stage.
func (r *repository) AddParticipant(ctx context.Context, input *types.ParticipantCreationInput) (*types.StageParticipant, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, database.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, keys.StageIDKey, input.StageID)
	logger = logger.WithValue(keys.StageIDKey, input.StageID)

	if err := r.generatedQuerier.CreateParticipant(ctx, r.db, &generated.CreateParticipantParams{
		ID:          input.ID,
		StageID:     input.StageID,
		UserID:      database.NullStringFromStringPointer(input.UserID),
		DisplayName: input.DisplayName,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating participant")
	}

	participant := &types.StageParticipant{
		ID:            input.ID,
		StageID:       input.StageID,
		UserID:        input.UserID,
		DisplayName:   input.DisplayName,
		CastMemberIDs: []string{},
		IsDirector:    false,
		IsReady:       false,
		IsAudience:    false,
		CreatedAt:     r.CurrentTime(),
	}

	return participant, nil
}

// GetParticipant fetches a participant from the database.
func (r *repository) GetParticipant(ctx context.Context, participantID string) (*types.StageParticipant, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if participantID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue("participantID", participantID)
	tracing.AttachToSpan(span, "participantID", participantID)

	result, err := r.generatedQuerier.GetParticipant(ctx, r.db, participantID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching participant")
	}

	participant := r.convertParticipant(result)

	// Fetch roles
	roles, err := r.GetParticipantRoles(ctx, participantID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching participant roles")
	}
	participant.CastMemberIDs = roles

	return participant, nil
}

// GetActiveParticipantsForStage fetches all active participants for a stage.
func (r *repository) GetActiveParticipantsForStage(ctx context.Context, stageID string) ([]*types.StageParticipant, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if stageID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.StageIDKey, stageID)
	tracing.AttachToSpan(span, keys.StageIDKey, stageID)

	results, err := r.generatedQuerier.GetActiveParticipantsForStage(ctx, r.db, stageID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching active participants")
	}

	participants := make([]*types.StageParticipant, len(results))
	for i, result := range results {
		participant := r.convertParticipant(result)

		// Fetch roles for each participant
		roles, err := r.GetParticipantRoles(ctx, participant.ID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching participant roles")
		}
		participant.CastMemberIDs = roles
		participants[i] = participant
	}

	return participants, nil
}

// UpdateParticipant updates a participant in the database.
func (r *repository) UpdateParticipant(ctx context.Context, participantID string, input *types.ParticipantUpdateInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if participantID == "" {
		return database.ErrInvalidIDProvided
	}
	logger = logger.WithValue("participantID", participantID)
	tracing.AttachToSpan(span, "participantID", participantID)

	params := &generated.UpdateParticipantParams{
		ID: participantID,
	}

	if input.IsDirector != nil {
		params.IsDirector = sql.NullBool{Bool: *input.IsDirector, Valid: true}
	}
	if input.IsReady != nil {
		params.IsReady = sql.NullBool{Bool: *input.IsReady, Valid: true}
	}
	if input.IsAudience != nil {
		params.IsAudience = sql.NullBool{Bool: *input.IsAudience, Valid: true}
	}

	rowsAffected, err := r.generatedQuerier.UpdateParticipant(ctx, r.db, params)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating participant")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// SetParticipantLeft marks a participant as having left the stage.
func (r *repository) SetParticipantLeft(ctx context.Context, participantID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if participantID == "" {
		return database.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, "participantID", participantID)

	logger := r.logger.WithValue("participantID", participantID)

	rowsAffected, err := r.generatedQuerier.SetParticipantLeft(ctx, r.db, participantID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "setting participant left")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// AddParticipantRole adds a cast member role to a participant.
func (r *repository) AddParticipantRole(ctx context.Context, participantID, castMemberID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if participantID == "" || castMemberID == "" {
		return database.ErrInvalidIDProvided
	}
	logger = logger.WithValue("participantID", participantID).WithValue("castMemberID", castMemberID)
	tracing.AttachToSpan(span, "participantID", participantID)
	tracing.AttachToSpan(span, "castMemberID", castMemberID)

	if err := r.generatedQuerier.AddParticipantRole(ctx, r.db, &generated.AddParticipantRoleParams{
		StageParticipantID: participantID,
		CastMemberID:       castMemberID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "adding participant role")
	}

	return nil
}

// RemoveParticipantRole removes a cast member role from a participant.
func (r *repository) RemoveParticipantRole(ctx context.Context, participantID, castMemberID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if participantID == "" || castMemberID == "" {
		return database.ErrInvalidIDProvided
	}
	logger = logger.WithValue("participantID", participantID).WithValue("castMemberID", castMemberID)
	tracing.AttachToSpan(span, "participantID", participantID)
	tracing.AttachToSpan(span, "castMemberID", castMemberID)

	_, err := r.generatedQuerier.RemoveParticipantRole(ctx, r.db, &generated.RemoveParticipantRoleParams{
		StageParticipantID: participantID,
		CastMemberID:       castMemberID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "removing participant role")
	}

	return nil
}

// GetParticipantRoles fetches the cast member roles for a participant.
func (r *repository) GetParticipantRoles(ctx context.Context, participantID string) ([]string, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if participantID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue("participantID", participantID)
	tracing.AttachToSpan(span, "participantID", participantID)

	roles, err := r.generatedQuerier.GetParticipantRoles(ctx, r.db, participantID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching participant roles")
	}

	return roles, nil
}

// Helper functions

func (r *repository) convertStage(result *generated.Stages) *types.Stage {
	return &types.Stage{
		ID:               result.ID,
		Code:             result.Code,
		PlayScriptID:     database.StringPointerFromNullString(result.PlayScriptID),
		CurrentLineIndex: int(result.CurrentLineIndex),
		Phase:            types.StagePhase(result.Phase),
		CreatedAt:        result.CreatedAt,
		LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
		Participants:     []*types.StageParticipant{},
	}
}

func (r *repository) convertParticipant(result *generated.StageParticipants) *types.StageParticipant {
	return &types.StageParticipant{
		ID:            result.ID,
		StageID:       result.StageID,
		UserID:        database.StringPointerFromNullString(result.UserID),
		DisplayName:   result.DisplayName,
		IsDirector:    result.IsDirector,
		IsReady:       result.IsReady,
		IsAudience:    result.IsAudience,
		CreatedAt:     result.CreatedAt,
		LeftAt:        database.TimePointerFromNullTime(result.LeftAt),
		CastMemberIDs: []string{},
	}
}
