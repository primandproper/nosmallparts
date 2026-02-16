package stages

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	platformws "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/websocket"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Send pings to peer every 30 seconds.
	pingPeriod = 30 * time.Second

	// Circuit breaker: close connection only after this many consecutive pings
	// without a pong. Tolerates brief stalls (e.g. browser tab backgrounded).
	pongFailuresBeforeClose = 5

	// Read deadline: reset on each pong. Slightly longer than
	// pingPeriod*pongFailuresBeforeClose so we don't close via read timeout
	// before the circuit breaker would.
	readDeadline = (pongFailuresBeforeClose + 2) * pingPeriod

	// Maximum message size allowed from peer.
	maxMessageSize = 512 * 1024 // 512KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking for production
		return true
	},
}

// Handler handles WebSocket connections for stages.
type Handler struct {
	stagesRepo      stagesdomain.Repository
	playScriptsRepo playscripts.Repository
	connections     *platformws.ConnectionManager
	logger          logging.Logger
	tracer          tracing.Tracer
}

// NewHandler creates a new WebSocket handler.
func NewHandler(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	stagesRepo stagesdomain.Repository,
	playScriptsRepo playscripts.Repository,
	connections *platformws.ConnectionManager,
) *Handler {
	return &Handler{
		stagesRepo:      stagesRepo,
		playScriptsRepo: playScriptsRepo,
		connections:     connections,
		logger:          logging.EnsureLogger(logger).WithName("stages_websocket"),
		tracer:          tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer("stages_websocket")),
	}
}

// ServeHTTP handles WebSocket upgrade and message processing.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade connection", err)
		return
	}

	connCtx := platformws.NewConnectionContext(conn)
	defer h.cleanup(ctx, connCtx)

	// Configure connection
	conn.SetReadLimit(maxMessageSize)
	if err = conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
		observability.AcknowledgeError(err, h.logger, span, "failed to set read deadline")
	}
	conn.SetPongHandler(func(string) error {
		connCtx.MarkAlive()
		if err = conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
			observability.AcknowledgeError(err, h.logger, span, "failed to set read deadline")
		}
		return nil
	})

	// Start heartbeat
	go h.heartbeat(connCtx)

	// Handle incoming messages
	h.readPump(ctx, connCtx)
}

// heartbeat sends periodic ping messages and closes the connection only after
// pongFailuresBeforeClose consecutive pings without a pong (circuit breaker).
func (h *Handler) heartbeat(connCtx *platformws.ConnectionContext) {
	ticker := time.NewTicker(pingPeriod)
	connCtx.SetHeartbeatTicker(ticker)
	defer ticker.Stop()

	consecutiveMissed := 0

	for {
		select {
		case <-ticker.C:
			if connCtx.CheckAlive() {
				consecutiveMissed = 0
			} else {
				consecutiveMissed++
				if consecutiveMissed >= pongFailuresBeforeClose {
					h.logger.WithValue("consecutiveMissed", consecutiveMissed).Debug("connection dead after consecutive missed pongs, closing")
					if err := connCtx.Conn.Close(); err != nil {
						observability.AcknowledgeError(err, h.logger, nil, "failed to close connection")
					}
					return
				}
			}
			connCtx.MarkDead()
			if err := connCtx.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				observability.AcknowledgeError(err, h.logger, nil, "setting write deadline")
			}

			if err := connCtx.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-connCtx.Done():
			return
		}
	}
}

// readPump reads messages from the WebSocket connection.
func (h *Handler) readPump(ctx context.Context, connCtx *platformws.ConnectionContext) {
	for {
		var msg stagesdomain.ClientMessage
		if err := connCtx.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("unexpected close error", err)
			}
			return
		}

		h.handleMessage(ctx, connCtx, &msg)
	}
}

// cleanup handles connection cleanup when a client disconnects.
func (h *Handler) cleanup(ctx context.Context, connCtx *platformws.ConnectionContext) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	connCtx.Close()

	if connCtx.GroupID != "" && connCtx.MemberID != "" {
		h.logger.WithValue("stageID", connCtx.GroupID).WithValue("participantID", connCtx.MemberID).Debug("participant disconnected")

		// Mark participant as left in database
		if err := h.stagesRepo.SetParticipantLeft(ctx, connCtx.MemberID); err != nil {
			h.logger.Error("failed to mark participant as left", err)
		}

		// Remove from connection manager
		h.connections.Remove(ctx, connCtx.GroupID, connCtx.MemberID)

		// Broadcast updated state to remaining participants
		h.broadcastStageState(ctx, connCtx.GroupID)

		// Archive stage if no more connections
		if !h.connections.GroupHasConnections(ctx, connCtx.GroupID) {
			h.logger.WithValue("stageID", connCtx.GroupID).Debug("no more connections, archiving stage")
			if err := h.stagesRepo.ArchiveStage(ctx, connCtx.GroupID); err != nil {
				h.logger.Error("failed to archive stage", err)
			}
		}
	}
}

// handleMessage routes incoming messages to the appropriate handler.
func (h *Handler) handleMessage(ctx context.Context, connCtx *platformws.ConnectionContext, msg *stagesdomain.ClientMessage) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	switch msg.Type {
	case stagesdomain.ClientMessageTypeJoin:
		h.handleJoin(ctx, connCtx, msg)
	case stagesdomain.ClientMessageTypeToggleCastMember:
		h.handleToggleCastMember(ctx, connCtx, msg)
	case stagesdomain.ClientMessageTypeSetDirector:
		h.handleSetDirector(ctx, connCtx, msg)
	case stagesdomain.ClientMessageTypeSetAudience:
		h.handleSetAudience(ctx, connCtx, msg)
	case stagesdomain.ClientMessageTypeSelectPlay:
		h.handleSelectPlay(ctx, connCtx, msg)
	case stagesdomain.ClientMessageTypeSetReady:
		h.handleSetReady(ctx, connCtx, msg)
	case stagesdomain.ClientMessageTypeSendChat:
		h.handleSendChat(ctx, connCtx, msg)
	case stagesdomain.ClientMessageTypeStartPlay:
		h.handleStartPlay(ctx, connCtx)
	case stagesdomain.ClientMessageTypeNextLine:
		h.handleNextLine(ctx, connCtx)
	case stagesdomain.ClientMessageTypeResetPlay:
		h.handleResetPlay(ctx, connCtx)
	case stagesdomain.ClientMessageTypeHeartbeat:
		h.handleHeartbeat(ctx, connCtx)
	default:
		h.sendError(ctx, connCtx, "unknown message type")
	}
}

// sendError sends an error message to the client.
func (h *Handler) sendError(ctx context.Context, connCtx *platformws.ConnectionContext, message string) {
	_, span := h.tracer.StartSpan(ctx)
	defer span.End()

	if err := connCtx.WriteJSON(&stagesdomain.ServerMessage{
		Type:  stagesdomain.ServerMessageTypeError,
		Error: message,
	}); err != nil {
		observability.AcknowledgeError(err, h.logger, span, "failed to send error message")
	}
}

// broadcastStageState broadcasts the current stage state to all connected clients.
func (h *Handler) broadcastStageState(ctx context.Context, stageID string) {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	stage, err := h.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		h.logger.Error("failed to get stage for broadcast", err)
		return
	}

	stageState := h.buildStageState(ctx, stage)

	h.connections.BroadcastToGroup(ctx, stageID, &stagesdomain.ServerMessage{
		Type:  stagesdomain.ServerMessageTypeStageState,
		State: stageState,
	})
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

	// Fetch selected play if set
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
		// Fallback to less secure but still random approach
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
