package stages

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	pusherpkg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/pusher"

	"github.com/go-chi/chi/v5"
)

// --- Request/Response types ---

type CreateStageResponse struct {
	StageID string `json:"stageId"`
	Code    string `json:"code"`
}

type StageExistsResponse struct {
	Exists bool `json:"exists"`
}

type JoinStageRequest struct {
	StageID         string `json:"stageId"`
	ParticipantName string `json:"participantName"`
}

type JoinStageResponse struct {
	StageState     *stagesdomain.StageState  `json:"stageState"`
	ParticipantID  string                    `json:"participantId"`
	PusherKey      string                    `json:"pusherKey"`
	PusherCluster  string                    `json:"pusherCluster"`
	AvailablePlays []*playscripts.PlayScript `json:"availablePlays"`
}

type LeaveStageRequest struct {
	ParticipantID string `json:"participantId"`
}

type ToggleCastMemberRequest struct {
	ParticipantID string `json:"participantId"`
	CastMemberID  string `json:"castMemberId"`
}

type SetDirectorRequest struct {
	ParticipantID string `json:"participantId"`
	IsDirector    bool   `json:"isDirector"`
}

type SetAudienceRequest struct {
	ParticipantID string `json:"participantId"`
	IsAudience    bool   `json:"isAudience"`
}

type SetReadyRequest struct {
	ParticipantID string `json:"participantId"`
	IsReady       bool   `json:"isReady"`
}

type SelectPlayRequest struct {
	PlayID string `json:"playId"`
}

type ChatRequest struct {
	ParticipantID string `json:"participantId"`
	Text          string `json:"text"`
	CastOnly      bool   `json:"castOnly"`
}

type PusherAuthRequest struct {
	SocketID      string `json:"socket_id"`
	ChannelName   string `json:"channel_name"`
	ParticipantID string `json:"participant_id"`
}

// --- Helper ---

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Error("failed to encode response", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, v any) error {
	defer func() { _ = r.Body.Close() }() //nolint:errcheck // response body close error is not actionable
	body, err := io.ReadAll(io.LimitReader(r.Body, 512*1024))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// --- Endpoints ---

// CreateStageHTTP handles POST /api/stages to create a new stage.
func (h *Handler) CreateStageHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stage, err := h.stagesRepo.CreateStage(ctx, &stagesdomain.StageCreationInput{
		ID:   identifiers.New(),
		Code: generateStageCode(),
	})
	if err != nil {
		h.logger.Error("failed to create stage", err)
		h.writeError(w, http.StatusInternalServerError, "failed to create stage")
		return
	}

	h.logger.WithValue("stageID", stage.ID).WithValue("code", stage.Code).Debug("created stage via HTTP")
	h.writeJSON(w, http.StatusCreated, &CreateStageResponse{
		StageID: stage.ID,
		Code:    stage.Code,
	})
}

// CheckStageExists handles GET /api/stages/{code}/exists.
func (h *Handler) CheckStageExists(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	_, err := h.stagesRepo.GetStageByCode(ctx, code)
	exists := true
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			exists = false
		} else {
			h.logger.Error("failed to check stage existence", err)
			h.writeError(w, http.StatusInternalServerError, "failed to check stage existence")
			return
		}
	}

	h.writeJSON(w, http.StatusOK, &StageExistsResponse{Exists: exists})
}

// JoinStageHTTP handles POST /api/stages/join.
func (h *Handler) JoinStageHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	var req JoinStageRequest
	if err := decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ParticipantName == "" {
		h.writeError(w, http.StatusBadRequest, "participantName is required")
		return
	}

	stageID := strings.TrimSpace(req.StageID)

	var stage *stagesdomain.Stage
	var err error

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
				stage, err = h.stagesRepo.GetStageByCode(ctx, strings.ToUpper(stageID))
			}
		}
	}

	if stage == nil || errors.Is(err, sql.ErrNoRows) {
		stage, err = h.stagesRepo.CreateStage(ctx, &stagesdomain.StageCreationInput{
			ID:   identifiers.New(),
			Code: generateStageCode(),
		})
		if err != nil {
			h.logger.Error("failed to create stage", err)
			h.writeError(w, http.StatusInternalServerError, "failed to create stage")
			return
		}
	} else if err != nil {
		h.logger.Error("failed to get stage", err)
		h.writeError(w, http.StatusInternalServerError, "failed to get stage")
		return
	}

	// Try rejoin first
	participant, err := h.stagesRepo.RejoinParticipant(ctx, stage.ID, req.ParticipantName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		h.logger.Error("failed to rejoin participant", err)
		h.writeError(w, http.StatusInternalServerError, "failed to join stage")
		return
	}
	if participant == nil {
		participant, err = h.stagesRepo.AddParticipant(ctx, &stagesdomain.ParticipantCreationInput{
			ID:          identifiers.New(),
			StageID:     stage.ID,
			DisplayName: req.ParticipantName,
		})
		if err != nil {
			h.logger.Error("failed to add participant", err)
			h.writeError(w, http.StatusInternalServerError, "failed to join stage")
			return
		}
	}

	// Refresh stage
	stage, err = h.stagesRepo.GetStageByID(ctx, stage.ID)
	if err != nil {
		h.logger.Error("failed to get stage", err)
		h.writeError(w, http.StatusInternalServerError, "failed to get stage")
		return
	}

	stageState := h.buildStageState(ctx, stage)

	// Get available plays
	filter := filtering.DefaultQueryFilter()
	result, err := h.playScriptsRepo.GetPublicPlayScripts(ctx, filter)
	if err != nil {
		h.logger.Error("failed to get public plays", err)
	}
	var availablePlays []*playscripts.PlayScript
	if result != nil {
		availablePlays = result.Data
	}

	// Broadcast updated state to existing participants
	go h.broadcastStageState(ctx, stage.ID)

	h.writeJSON(w, http.StatusOK, &JoinStageResponse{
		ParticipantID:  participant.ID,
		StageState:     stageState,
		AvailablePlays: availablePlays,
		PusherKey:      h.pusher.Key(),
		PusherCluster:  h.pusher.Cluster(),
	})
}

// LeaveStageHTTP handles POST /api/stages/{stageID}/leave.
func (h *Handler) LeaveStageHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")

	var req LeaveStageRequest
	if err := decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ParticipantID != "" {
		if err := h.stagesRepo.SetParticipantLeft(ctx, req.ParticipantID); err != nil {
			h.logger.Error("failed to mark participant as left", err)
		}
	}

	// Check if stage should be archived
	participants, err := h.stagesRepo.GetActiveParticipantsForStage(ctx, stageID)
	if err != nil {
		h.logger.Error("failed to get active participants", err)
	} else if len(participants) == 0 {
		if err = h.stagesRepo.ArchiveStage(ctx, stageID); err != nil {
			h.logger.Error("failed to archive stage", err)
		}
	}

	go h.broadcastStageState(ctx, stageID)

	w.WriteHeader(http.StatusNoContent)
}

// ToggleCastMemberHTTP handles POST /api/stages/{stageID}/toggle-cast-member.
func (h *Handler) ToggleCastMemberHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")

	var req ToggleCastMemberRequest
	if err := decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ParticipantID == "" || req.CastMemberID == "" {
		h.writeError(w, http.StatusBadRequest, "participantId and castMemberId are required")
		return
	}

	roles, err := h.stagesRepo.GetParticipantRoles(ctx, req.ParticipantID)
	if err != nil {
		observability.AcknowledgeError(err, h.logger, span, "failed to get participant roles")
		h.writeError(w, http.StatusInternalServerError, "failed to get participant roles")
		return
	}

	hasRole := false
	for _, role := range roles {
		if role == req.CastMemberID {
			hasRole = true
			break
		}
	}

	if hasRole {
		if err = h.stagesRepo.RemoveParticipantRole(ctx, req.ParticipantID, req.CastMemberID); err != nil {
			h.logger.Error("failed to remove participant role", err)
			h.writeError(w, http.StatusInternalServerError, "failed to update role")
			return
		}
	} else {
		if err = h.stagesRepo.AddParticipantRole(ctx, req.ParticipantID, req.CastMemberID); err != nil {
			h.logger.Error("failed to add participant role", err)
			h.writeError(w, http.StatusInternalServerError, "failed to update role")
			return
		}
	}

	// Broadcast cast members updated
	participants, err := h.stagesRepo.GetActiveParticipantsForStage(ctx, stageID)
	if err != nil {
		h.logger.Error("failed to get participants for broadcast", err)
	} else {
		if triggerErr := h.pusher.TriggerStageEvent(ctx, stageID, "cast-members-updated", map[string]any{
			"participants": participants,
		}); triggerErr != nil {
			h.logger.Error("failed to trigger cast-members-updated event", triggerErr)
		}
	}

	go h.broadcastStageState(ctx, stageID)

	w.WriteHeader(http.StatusNoContent)
}

// SetDirectorHTTP handles POST /api/stages/{stageID}/set-director.
func (h *Handler) SetDirectorHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")

	var req SetDirectorRequest
	if err := decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.stagesRepo.UpdateParticipant(ctx, req.ParticipantID, &stagesdomain.ParticipantUpdateInput{
		IsDirector: &req.IsDirector,
	}); err != nil {
		h.logger.Error("failed to update participant", err)
		h.writeError(w, http.StatusInternalServerError, "failed to update participant")
		return
	}

	go h.broadcastStageState(ctx, stageID)

	w.WriteHeader(http.StatusNoContent)
}

// SetAudienceHTTP handles POST /api/stages/{stageID}/set-audience.
func (h *Handler) SetAudienceHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")

	var req SetAudienceRequest
	if err := decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.stagesRepo.UpdateParticipant(ctx, req.ParticipantID, &stagesdomain.ParticipantUpdateInput{
		IsAudience: &req.IsAudience,
	}); err != nil {
		h.logger.Error("failed to update participant", err)
		h.writeError(w, http.StatusInternalServerError, "failed to update participant")
		return
	}

	go h.broadcastStageState(ctx, stageID)

	w.WriteHeader(http.StatusNoContent)
}

// SetReadyHTTP handles POST /api/stages/{stageID}/set-ready.
func (h *Handler) SetReadyHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")

	var req SetReadyRequest
	if err := decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.stagesRepo.UpdateParticipant(ctx, req.ParticipantID, &stagesdomain.ParticipantUpdateInput{
		IsReady: &req.IsReady,
	}); err != nil {
		h.logger.Error("failed to update participant", err)
		h.writeError(w, http.StatusInternalServerError, "failed to update participant")
		return
	}

	go h.broadcastStageState(ctx, stageID)

	w.WriteHeader(http.StatusNoContent)
}

// SelectPlayHTTP handles POST /api/stages/{stageID}/select-play.
func (h *Handler) SelectPlayHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")

	var req SelectPlayRequest
	if err := decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlayID == "" {
		h.writeError(w, http.StatusBadRequest, "playId is required")
		return
	}

	if err := h.stagesRepo.UpdateStage(ctx, stageID, &stagesdomain.StageUpdateInput{
		PlayScriptID: &req.PlayID,
	}); err != nil {
		h.logger.Error("failed to update stage", err)
		h.writeError(w, http.StatusInternalServerError, "failed to select play")
		return
	}

	go h.broadcastStageState(ctx, stageID)

	w.WriteHeader(http.StatusNoContent)
}

// SendChatHTTP handles POST /api/stages/{stageID}/chat.
func (h *Handler) SendChatHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")

	var req ChatRequest
	if err := decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	participant, err := h.stagesRepo.GetParticipant(ctx, req.ParticipantID)
	if err != nil {
		h.logger.Error("failed to get participant", err)
		h.writeError(w, http.StatusInternalServerError, "failed to send chat")
		return
	}

	chatMessage := &stagesdomain.ChatMessage{
		ID:         identifiers.New(),
		SenderID:   participant.ID,
		SenderName: participant.DisplayName,
		Text:       req.Text,
		CastOnly:   req.CastOnly,
		Timestamp:  time.Now().UnixMilli(),
	}

	eventData := map[string]any{"message": chatMessage}

	if req.CastOnly {
		if err = h.pusher.TriggerCastEvent(ctx, stageID, "chat-message", eventData); err != nil {
			h.logger.Error("failed to trigger cast chat event", err)
		}
	} else {
		if err = h.pusher.TriggerStageEvent(ctx, stageID, "chat-message", eventData); err != nil {
			h.logger.Error("failed to trigger chat event", err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// StartPlayHTTP handles POST /api/stages/{stageID}/start.
func (h *Handler) StartPlayHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")

	countdownPhase := stagesdomain.StagePhaseCountdown
	if err := h.stagesRepo.UpdateStage(ctx, stageID, &stagesdomain.StageUpdateInput{
		Phase: &countdownPhase,
	}); err != nil {
		h.logger.Error("failed to update stage", err)
		h.writeError(w, http.StatusInternalServerError, "failed to start play")
		return
	}

	if err := h.pusher.TriggerStageEvent(ctx, stageID, "countdown-start", map[string]any{"duration": 3}); err != nil {
		h.logger.Error("failed to trigger countdown-start event", err)
	}
	h.broadcastStageState(ctx, stageID)

	// Transition to performing after countdown
	go func() {
		time.Sleep(3 * time.Second)

		performingPhase := stagesdomain.StagePhasePerforming
		lineIndex := 0
		if err := h.stagesRepo.UpdateStage(ctx, stageID, &stagesdomain.StageUpdateInput{
			Phase:            &performingPhase,
			CurrentLineIndex: &lineIndex,
		}); err != nil {
			h.logger.Error("failed to update stage to performing", err)
			return
		}

		if err := h.pusher.TriggerStageEvent(ctx, stageID, "play-started", map[string]any{}); err != nil {
			h.logger.Error("failed to trigger play-started event", err)
		}
		h.broadcastStageState(ctx, stageID)
	}()

	w.WriteHeader(http.StatusNoContent)
}

// NextLineHTTP handles POST /api/stages/{stageID}/next-line.
func (h *Handler) NextLineHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")
	logger := h.logger.WithValue("stageID", stageID)

	stage, err := h.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		logger.Error("failed to get stage", err)
		h.writeError(w, http.StatusInternalServerError, "failed to get stage")
		return
	}

	if stage.PlayScriptID == nil || *stage.PlayScriptID == "" {
		logger.Info("nextLine ignored: no play selected")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	play, err := h.playScriptsRepo.GetPlayScript(ctx, *stage.PlayScriptID, false)
	if err != nil {
		logger.Error("failed to get play", err)
		h.writeError(w, http.StatusInternalServerError, "failed to get play")
		return
	}

	turnIndices := playscripts.GetSortedUniqueLineIndices(play)
	newTurnIndex := stage.CurrentLineIndex + 1

	if newTurnIndex >= len(turnIndices) {
		// End the play
		discussionPhase := stagesdomain.StagePhaseDiscussion
		if err = h.stagesRepo.UpdateStage(ctx, stageID, &stagesdomain.StageUpdateInput{
			Phase: &discussionPhase,
		}); err != nil {
			logger.Error("failed to update stage to discussion", err)
			h.writeError(w, http.StatusInternalServerError, "failed to end play")
			return
		}

		playEndedData := map[string]any{}
		if play.PrecedesID != nil && *play.PrecedesID != "" {
			nextPlay, nextErr := h.playScriptsRepo.GetPlayScript(ctx, *play.PrecedesID, false)
			if nextErr == nil {
				playEndedData["nextPlay"] = nextPlay
			}
		}
		if triggerErr := h.pusher.TriggerStageEvent(ctx, stageID, "play-ended", playEndedData); triggerErr != nil {
			logger.Error("failed to trigger play-ended event", triggerErr)
		}
	} else {
		if err = h.stagesRepo.UpdateStage(ctx, stageID, &stagesdomain.StageUpdateInput{
			CurrentLineIndex: &newTurnIndex,
		}); err != nil {
			logger.Error("failed to advance line", err)
			h.writeError(w, http.StatusInternalServerError, "failed to advance line")
			return
		}

		if triggerErr := h.pusher.TriggerStageEvent(ctx, stageID, "line-advanced", map[string]any{
			"currentLineIndex": newTurnIndex,
		}); triggerErr != nil {
			logger.Error("failed to trigger line-advanced event", triggerErr)
		}
	}

	go h.broadcastStageState(ctx, stageID)

	w.WriteHeader(http.StatusNoContent)
}

// ResetPlayHTTP handles POST /api/stages/{stageID}/reset.
func (h *Handler) ResetPlayHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")

	lobbyPhase := stagesdomain.StagePhaseLobby
	lineIndex := 0
	if err := h.stagesRepo.UpdateStage(ctx, stageID, &stagesdomain.StageUpdateInput{
		Phase:            &lobbyPhase,
		CurrentLineIndex: &lineIndex,
	}); err != nil {
		h.logger.Error("failed to reset play", err)
		h.writeError(w, http.StatusInternalServerError, "failed to reset play")
		return
	}

	// Reset ready status
	participants, err := h.stagesRepo.GetActiveParticipantsForStage(ctx, stageID)
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

	go h.broadcastStageState(ctx, stageID)

	w.WriteHeader(http.StatusNoContent)
}

// GetStageStateHTTP handles GET /api/stages/{stageID}/state.
func (h *Handler) GetStageStateHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stageID := chi.URLParam(r, "stageID")

	stage, err := h.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.writeError(w, http.StatusNotFound, "stage not found")
			return
		}
		h.logger.Error("failed to get stage", err)
		h.writeError(w, http.StatusInternalServerError, "failed to get stage")
		return
	}

	stageState := h.buildStageState(ctx, stage)
	h.writeJSON(w, http.StatusOK, stageState)
}

// PusherAuthHTTP handles POST /api/pusher/auth.
func (h *Handler) PusherAuthHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	var req PusherAuthRequest
	if err := decodeJSON(r, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SocketID == "" || req.ChannelName == "" {
		h.writeError(w, http.StatusBadRequest, "socket_id and channel_name are required")
		return
	}

	// For private-cast channels, verify participant is not audience
	if strings.HasPrefix(req.ChannelName, "private-cast-") {
		if req.ParticipantID == "" {
			h.writeError(w, http.StatusForbidden, "participant_id required for cast channel")
			return
		}

		participant, err := h.stagesRepo.GetParticipant(ctx, req.ParticipantID)
		if err != nil {
			h.logger.Error("failed to get participant for auth", err)
			h.writeError(w, http.StatusForbidden, "unauthorized")
			return
		}
		if participant.IsAudience {
			h.writeError(w, http.StatusForbidden, "audience members cannot join cast channel")
			return
		}

		response, err := h.pusher.AuthorizePrivateChannel(req.SocketID, req.ChannelName)
		if err != nil {
			h.logger.Error("failed to authorize private channel", err)
			h.writeError(w, http.StatusInternalServerError, "authorization failed")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, writeErr := w.Write(response); writeErr != nil {
			h.logger.Error("failed to write response", writeErr)
		}
		return
	}

	// For presence channels
	if strings.HasPrefix(req.ChannelName, "presence-stage-") {
		userData := pusherpkg.PresenceUserData{
			UserID: req.ParticipantID,
		}
		if req.ParticipantID != "" {
			participant, err := h.stagesRepo.GetParticipant(ctx, req.ParticipantID)
			if err == nil {
				userData.UserInfo = map[string]string{
					"name": participant.DisplayName,
				}
			}
		}

		response, err := h.pusher.AuthorizePresenceChannel(req.SocketID, req.ChannelName, userData)
		if err != nil {
			h.logger.Error("failed to authorize presence channel", err)
			h.writeError(w, http.StatusInternalServerError, "authorization failed")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, writeErr := w.Write(response); writeErr != nil {
			h.logger.Error("failed to write response", writeErr)
		}
		return
	}

	h.writeError(w, http.StatusForbidden, "unknown channel type")
}

// PusherWebhookHTTP handles POST /api/pusher/webhook.
func (h *Handler) PusherWebhookHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	body, err := io.ReadAll(io.LimitReader(r.Body, 512*1024))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			h.logger.Error("closing request body", closeErr)
		}
	}()

	signature := r.Header.Get("X-Pusher-Signature")
	if !h.pusher.ValidateWebhook(signature, body) {
		h.writeError(w, http.StatusForbidden, "invalid signature")
		return
	}

	payload, err := pusherpkg.ParseWebhookPayload(body)
	if err != nil {
		h.logger.Error("failed to parse webhook payload", err)
		h.writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	for i := range payload.Events {
		event := &payload.Events[i]
		if event.Name != "member_removed" || !strings.HasPrefix(event.Channel, "presence-stage-") {
			continue
		}

		stageID := strings.TrimPrefix(event.Channel, "presence-stage-")
		participantID := event.UserID

		if participantID != "" {
			if err = h.stagesRepo.SetParticipantLeft(ctx, participantID); err != nil {
				h.logger.Error(fmt.Sprintf("failed to mark participant %s as left", participantID), err)
			}
		}

		// Archive stage if empty
		participants, pErr := h.stagesRepo.GetActiveParticipantsForStage(ctx, stageID)
		switch {
		case pErr != nil:
			h.logger.Error("failed to get active participants", pErr)
		case len(participants) == 0:
			if err = h.stagesRepo.ArchiveStage(ctx, stageID); err != nil {
				h.logger.Error("failed to archive stage", err)
			}
		default:
			go h.broadcastStageState(ctx, stageID)
		}
	}

	w.WriteHeader(http.StatusOK)
}
