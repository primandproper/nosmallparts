package stages

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	platformws "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/websocket"
)

// handleJoin handles a client joining a stage.
func (h *Handler) handleJoin(ctx context.Context, connCtx *platformws.ConnectionContext, msg *stagesdomain.ClientMessage) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	stageID := strings.TrimSpace(msg.StageID)
	participantName := msg.ParticipantName

	if participantName == "" {
		h.sendError(ctx, connCtx, "participant name is required")
		return
	}

	var stage *stagesdomain.Stage
	var err error

	// Try to find existing stage. Frontend sends the code (e.g. KNZJBV) in stageId when joining via URL.
	// Codes are 6 alphanumeric chars; IDs are xid (~20 chars). Try code first when input looks like a code.
	if stageID != "" {
		codeUpper := strings.ToUpper(strings.TrimSpace(stageID))
		if looksLikeStageCode(codeUpper) {
			stage, err = h.stagesRepo.GetStageByCode(ctx, codeUpper)
			if errors.Is(err, sql.ErrNoRows) {
				stage, err = h.stagesRepo.GetStageByID(ctx, stageID)
			}
		} else {
			stage, err = h.stagesRepo.GetStageByID(ctx, stageID)
			if errors.Is(err, sql.ErrNoRows) {
				stage, err = h.stagesRepo.GetStageByCode(ctx, codeUpper)
			}
		}
	}

	// Create new stage if not found
	if stage == nil || errors.Is(err, sql.ErrNoRows) {
		stage, err = h.stagesRepo.CreateStage(ctx, &stagesdomain.StageCreationInput{
			ID:   identifiers.New(),
			Code: generateStageCode(),
		})
		if err != nil {
			h.logger.Error("failed to create stage", err)
			h.sendError(ctx, connCtx, "failed to create stage")
			return
		}
		h.logger.WithValue("stageID", stage.ID).WithValue("code", stage.Code).Debug("created new stage")
	} else if err != nil {
		h.logger.Error("failed to get stage", err)
		h.sendError(ctx, connCtx, "failed to get stage")
		return
	}

	// Add participant
	participant, err := h.stagesRepo.AddParticipant(ctx, &stagesdomain.ParticipantCreationInput{
		ID:          identifiers.New(),
		StageID:     stage.ID,
		DisplayName: participantName,
	})
	if err != nil {
		h.logger.Error("failed to add participant", err)
		h.sendError(ctx, connCtx, "failed to join stage")
		return
	}

	// Update connection context
	connCtx.SetGroupAndMember(stage.ID, participant.ID)

	// Add to connection manager
	h.connections.Add(ctx, stage.ID, participant.ID, connCtx)

	// Refresh stage with participants
	stage, err = h.stagesRepo.GetStageByID(ctx, stage.ID)
	if errors.Is(err, sql.ErrNoRows) {
		stage, err = h.stagesRepo.CreateStage(ctx, &stagesdomain.StageCreationInput{})
	}
	if err != nil {
		h.logger.Error("failed to get stage", err)
		h.sendError(ctx, connCtx, "failed to get stage")
		return
	}
	stageState := h.buildStageState(ctx, stage)
	hasSelectedPlay := stageState.SelectedPlay != nil
	h.logger.WithValue("stageID", stage.ID).WithValue("code", stage.Code).WithValue("hasSelectedPlay", hasSelectedPlay).Debug("sending join confirmation")

	// Send join confirmation to the new participant
	if err = connCtx.WriteJSON(&stagesdomain.ServerMessage{
		Type:          stagesdomain.ServerMessageTypeJoined,
		ParticipantID: participant.ID,
		StageState:    stageState,
	}); err != nil {
		observability.AcknowledgeError(err, h.logger, span, "joining stage")
	}

	// Send available plays
	h.sendAvailablePlays(ctx, connCtx)

	// Broadcast updated state to all participants
	h.connections.BroadcastToGroup(ctx, stage.ID, &stagesdomain.ServerMessage{
		Type:  stagesdomain.ServerMessageTypeStageState,
		State: stageState,
	})
}

// handleToggleCastMember handles toggling a cast member assignment.
func (h *Handler) handleToggleCastMember(ctx context.Context, connCtx *platformws.ConnectionContext, msg *stagesdomain.ClientMessage) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	if connCtx.GroupID == "" {
		observability.AcknowledgeError(errors.New("stage ID is required"), h.logger, span, "stage ID is required")
		return
	}

	participantID := msg.ParticipantID
	castMemberID := msg.CastMemberID

	if participantID == "" || castMemberID == "" {
		observability.AcknowledgeError(errors.New("participant ID or cast member ID is required"), h.logger, span, "participant ID or cast member ID is required")
		return
	}

	// Get current roles
	roles, err := h.stagesRepo.GetParticipantRoles(ctx, participantID)
	if err != nil {
		observability.AcknowledgeError(err, h.logger, span, "failed to get participant roles")
		return
	}

	// Check if role exists
	hasRole := false
	for _, role := range roles {
		if role == castMemberID {
			hasRole = true
			break
		}
	}

	if hasRole {
		// Remove role
		if err = h.stagesRepo.RemoveParticipantRole(ctx, participantID, castMemberID); err != nil {
			h.logger.Error("failed to remove participant role", err)
			return
		}
	} else {
		// Add role
		if err = h.stagesRepo.AddParticipantRole(ctx, participantID, castMemberID); err != nil {
			h.logger.Error("failed to add participant role", err)
			return
		}
	}

	// Broadcast updated state
	h.broadcastCastMembersUpdated(ctx, connCtx.GroupID)
	h.broadcastStageState(ctx, connCtx.GroupID)
}

// handleSetDirector handles setting a participant as director.
func (h *Handler) handleSetDirector(ctx context.Context, connCtx *platformws.ConnectionContext, msg *stagesdomain.ClientMessage) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	if connCtx.GroupID == "" {
		return
	}

	if msg.IsDirector == nil {
		return
	}

	if err := h.stagesRepo.UpdateParticipant(ctx, msg.ParticipantID, &stagesdomain.ParticipantUpdateInput{
		IsDirector: msg.IsDirector,
	}); err != nil {
		h.logger.Error("failed to update participant", err)
		return
	}

	h.broadcastStageState(ctx, connCtx.GroupID)
}

// handleSetAudience handles setting a participant as audience.
func (h *Handler) handleSetAudience(ctx context.Context, connCtx *platformws.ConnectionContext, msg *stagesdomain.ClientMessage) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	if connCtx.GroupID == "" {
		return
	}

	if msg.IsAudience == nil {
		return
	}

	if err := h.stagesRepo.UpdateParticipant(ctx, msg.ParticipantID, &stagesdomain.ParticipantUpdateInput{
		IsAudience: msg.IsAudience,
	}); err != nil {
		h.logger.Error("failed to update participant", err)
		return
	}

	h.broadcastStageState(ctx, connCtx.GroupID)
}

// handleSelectPlay handles selecting a play for the stage.
func (h *Handler) handleSelectPlay(ctx context.Context, connCtx *platformws.ConnectionContext, msg *stagesdomain.ClientMessage) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	if connCtx.GroupID == "" {
		return
	}

	playID := msg.PlayID
	if playID == "" {
		return
	}

	if err := h.stagesRepo.UpdateStage(ctx, connCtx.GroupID, &stagesdomain.StageUpdateInput{
		PlayScriptID: &playID,
	}); err != nil {
		h.logger.Error("failed to update stage", err)
		return
	}

	h.broadcastStageState(ctx, connCtx.GroupID)
}

// handleSetReady handles setting a participant's ready status.
func (h *Handler) handleSetReady(ctx context.Context, connCtx *platformws.ConnectionContext, msg *stagesdomain.ClientMessage) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	if connCtx.GroupID == "" {
		return
	}

	if msg.IsReady == nil {
		return
	}

	if err := h.stagesRepo.UpdateParticipant(ctx, msg.ParticipantID, &stagesdomain.ParticipantUpdateInput{
		IsReady: msg.IsReady,
	}); err != nil {
		h.logger.Error("failed to update participant", err)
		return
	}

	h.broadcastStageState(ctx, connCtx.GroupID)
}

// handleSendChat handles sending a chat message.
func (h *Handler) handleSendChat(ctx context.Context, connCtx *platformws.ConnectionContext, msg *stagesdomain.ClientMessage) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	if connCtx.GroupID == "" {
		return
	}

	// Get sender info
	participant, err := h.stagesRepo.GetParticipant(ctx, msg.ParticipantID)
	if err != nil {
		h.logger.Error("failed to get participant", err)
		return
	}

	chatMessage := &stagesdomain.ChatMessage{
		ID:         identifiers.New(),
		SenderID:   participant.ID,
		SenderName: participant.DisplayName,
		Text:       msg.Text,
		CastOnly:   msg.CastOnly,
		Timestamp:  time.Now().UnixMilli(),
	}

	serverMsg := &stagesdomain.ServerMessage{
		Type:    stagesdomain.ServerMessageTypeChatMessage,
		Message: chatMessage,
	}

	if msg.CastOnly {
		// Only send to non-audience members
		h.connections.BroadcastToGroupFiltered(ctx, connCtx.GroupID, serverMsg, func(participantID string) bool {
			var p *stagesdomain.StageParticipant
			p, err = h.stagesRepo.GetParticipant(ctx, participantID)
			if err != nil {
				return false // Exclude if we can't fetch
			}
			return !p.IsAudience
		})
	} else {
		h.connections.BroadcastToGroup(ctx, connCtx.GroupID, serverMsg)
	}
}

// handleStartPlay handles starting the play.
func (h *Handler) handleStartPlay(ctx context.Context, connCtx *platformws.ConnectionContext) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	if connCtx.GroupID == "" {
		return
	}

	countdownPhase := stagesdomain.StagePhaseCountdown
	if err := h.stagesRepo.UpdateStage(ctx, connCtx.GroupID, &stagesdomain.StageUpdateInput{
		Phase: &countdownPhase,
	}); err != nil {
		h.logger.Error("failed to update stage", err)
		return
	}

	// Broadcast countdown start
	h.connections.BroadcastToGroup(ctx, connCtx.GroupID, &stagesdomain.ServerMessage{
		Type:     stagesdomain.ServerMessageTypeCountdownStart,
		Duration: 3, // 3 second countdown
	})

	// Start a goroutine to transition to performing after countdown
	go func() {
		time.Sleep(3 * time.Second)

		performingPhase := stagesdomain.StagePhasePerforming
		lineIndex := 0
		if err := h.stagesRepo.UpdateStage(ctx, connCtx.GroupID, &stagesdomain.StageUpdateInput{
			Phase:            &performingPhase,
			CurrentLineIndex: &lineIndex,
		}); err != nil {
			h.logger.Error("failed to update stage to performing", err)
			return
		}

		h.connections.BroadcastToGroup(ctx, connCtx.GroupID, &stagesdomain.ServerMessage{
			Type: stagesdomain.ServerMessageTypePlayStarted,
		})

		h.broadcastStageState(ctx, connCtx.GroupID)
	}()

	h.broadcastStageState(ctx, connCtx.GroupID)
}

// handleNextLine handles advancing to the next line.
func (h *Handler) handleNextLine(ctx context.Context, connCtx *platformws.ConnectionContext) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	logger := h.logger.WithValue("groupID", connCtx.GroupID)
	logger.Debug("handleNextLine received")

	if connCtx.GroupID == "" {
		logger.Info("nextLine ignored: connection has no GroupID (join may not have completed)")
		return
	}

	// Get current stage state
	stage, err := h.stagesRepo.GetStageByID(ctx, connCtx.GroupID)
	if err != nil {
		logger.Error("failed to get stage", err)
		return
	}

	// Get the play to check number of lines
	if stage.PlayScriptID == nil || *stage.PlayScriptID == "" {
		logger.WithValue("stageID", connCtx.GroupID).Info("nextLine ignored: no play selected on stage")
		return
	}

	play, err := h.playScriptsRepo.GetPlayScript(ctx, *stage.PlayScriptID, false)
	if err != nil {
		logger.Error("failed to get play", err)
		return
	}

	newLineIndex := stage.CurrentLineIndex + 1

	// Check if we've reached the end
	if newLineIndex >= len(play.Lines) {
		// End the play
		discussionPhase := stagesdomain.StagePhaseDiscussion
		if err = h.stagesRepo.UpdateStage(ctx, connCtx.GroupID, &stagesdomain.StageUpdateInput{
			Phase: &discussionPhase,
		}); err != nil {
			logger.Error("failed to update stage to discussion", err)
			return
		}

		playEndedMsg := &stagesdomain.ServerMessage{
			Type: stagesdomain.ServerMessageTypePlayEnded,
		}
		// If this act has a next act, include it so the frontend can offer "Continue to Act 2"
		if play.PrecedesID != nil && *play.PrecedesID != "" {
			var nextPlay *playscripts.PlayScript
			nextPlay, err = h.playScriptsRepo.GetPlayScript(ctx, *play.PrecedesID, false)
			if err == nil {
				playEndedMsg.NextPlay = nextPlay
			}
		}
		h.connections.BroadcastToGroup(ctx, connCtx.GroupID, playEndedMsg)
	} else {
		// Advance to next line
		if err = h.stagesRepo.UpdateStage(ctx, connCtx.GroupID, &stagesdomain.StageUpdateInput{
			CurrentLineIndex: &newLineIndex,
		}); err != nil {
			logger.Error("failed to advance line", err)
			return
		}

		logger.WithValue("stageID", connCtx.GroupID).WithValue("lineIndex", newLineIndex).Debug("advanced to next line")
		h.connections.BroadcastToGroup(ctx, connCtx.GroupID, &stagesdomain.ServerMessage{
			Type:             stagesdomain.ServerMessageTypeLineAdvanced,
			CurrentLineIndex: newLineIndex,
		})
	}

	h.broadcastStageState(ctx, connCtx.GroupID)
}

// handleResetPlay handles resetting the play.
func (h *Handler) handleResetPlay(ctx context.Context, connCtx *platformws.ConnectionContext) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	if connCtx.GroupID == "" {
		return
	}

	lobbyPhase := stagesdomain.StagePhaseLobby
	lineIndex := 0
	if err := h.stagesRepo.UpdateStage(ctx, connCtx.GroupID, &stagesdomain.StageUpdateInput{
		Phase:            &lobbyPhase,
		CurrentLineIndex: &lineIndex,
	}); err != nil {
		h.logger.Error("failed to reset play", err)
		return
	}

	// Also reset all participants' ready status
	participants, err := h.stagesRepo.GetActiveParticipantsForStage(ctx, connCtx.GroupID)
	if err != nil {
		observability.AcknowledgeError(err, h.logger, span, "failed to get active participants")
	}

	for _, p := range participants {
		isReady := false
		if err = h.stagesRepo.UpdateParticipant(ctx, p.ID, &stagesdomain.ParticipantUpdateInput{
			IsReady: &isReady,
		}); err != nil {
			observability.AcknowledgeError(err, h.logger, span, "failed to update participant")
		}
	}

	h.broadcastStageState(ctx, connCtx.GroupID)
}

// sendAvailablePlays sends the list of available plays to a participant.
func (h *Handler) sendAvailablePlays(ctx context.Context, connCtx *platformws.ConnectionContext) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	// Get public plays
	filter := filtering.DefaultQueryFilter()
	result, err := h.playScriptsRepo.GetPublicPlayScripts(ctx, filter)
	if err != nil {
		h.logger.Error("failed to get public plays", err)
		return
	}

	plays := make([]*playscripts.PlayScript, len(result.Data))
	copy(plays, result.Data)

	if err = connCtx.WriteJSON(&stagesdomain.ServerMessage{
		Type:  stagesdomain.ServerMessageTypeAvailablePlays,
		Plays: plays,
	}); err != nil {
		observability.AcknowledgeError(err, h.logger, span, "failed to send available plays")
	}
}

// broadcastCastMembersUpdated broadcasts updated cast member assignments.
func (h *Handler) broadcastCastMembersUpdated(ctx context.Context, stageID string) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	participants, err := h.stagesRepo.GetActiveParticipantsForStage(ctx, stageID)
	if err != nil {
		h.logger.Error("failed to get participants for broadcast", err)
		return
	}

	h.connections.BroadcastToGroup(ctx, stageID, &stagesdomain.ServerMessage{
		Type:         stagesdomain.ServerMessageTypeCastMembersUpdated,
		Participants: participants,
	})
}
