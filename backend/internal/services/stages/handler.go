package stages

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	pusherpkg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/pusher"
)

// Handler handles HTTP requests for stages.
type Handler struct {
	stagesRepo      stagesdomain.Repository
	playScriptsRepo playscripts.Repository
	pusher          *pusherpkg.Client
	logger          logging.Logger
	tracer          tracing.Tracer
}

// NewHandler creates a new Handler.
func NewHandler(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	stagesRepo stagesdomain.Repository,
	playScriptsRepo playscripts.Repository,
	pusher *pusherpkg.Client,
) *Handler {
	return &Handler{
		stagesRepo:      stagesRepo,
		playScriptsRepo: playScriptsRepo,
		pusher:          pusher,
		logger:          logging.NewNamedLogger(logger, "stages"),
		tracer:          tracing.NewNamedTracer(tracerProvider, "stages"),
	}
}

// broadcastStageState broadcasts the current stage state to all connected clients via Pusher.
func (h *Handler) broadcastStageState(ctx context.Context, stageID string) {
	stage, err := h.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		h.logger.Error("failed to get stage for broadcast", err)
		return
	}

	stageState := h.buildStageState(ctx, stage)
	if err = h.pusher.TriggerStageEvent(ctx, stageID, "stage-state", stageState); err != nil {
		h.logger.Error("failed to broadcast stage state", err)
	}
}

// buildStageState constructs a StageState from a Stage.
func (h *Handler) buildStageState(ctx context.Context, stage *stagesdomain.Stage) *stagesdomain.StageState {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	state := &stagesdomain.StageState{
		ID:               stage.ID,
		Code:             stage.Code,
		Participants:     stage.Participants,
		CurrentLineIndex: stage.CurrentLineIndex,
		Phase:            stage.Phase,
	}

	if stage.PlayScriptID == nil || *stage.PlayScriptID == "" {
		h.logger.WithValue("stageID", stage.ID).Debug("stage has no play selected")
	}
	if stage.PlayScriptID != nil && *stage.PlayScriptID != "" {
		play, err := h.playScriptsRepo.GetPlayScript(ctx, *stage.PlayScriptID, false)
		if err != nil {
			h.logger.WithValue("stageID", stage.ID).WithValue("playScriptID", *stage.PlayScriptID).Error("failed to fetch selected play for stage state", err)
		} else {
			state.SelectedPlay = play
		}
	}

	return state
}

// looksLikeStageCode returns true if s appears to be a stage code (6 alphanumeric chars)
// rather than an xid-style stage ID (~20 chars). Used to choose lookup order.
func looksLikeStageCode(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

// generateStageCode generates a 6-character stage code using cryptographic randomness.
func generateStageCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		for i := range randomBytes {
			randomBytes[i] = byte(time.Now().UnixNano() >> (i * 8))
		}
	}
	code := make([]byte, 6)
	for i := 0; i < 6; i++ {
		code[i] = chars[int(randomBytes[i])%len(chars)]
	}
	return string(code)
}
