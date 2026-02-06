package stages

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	platformws "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/websocket"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade connection", err)
		return
	}

	ctx := platformws.NewConnectionContext(conn)
	defer h.cleanup(ctx)

	// Configure connection
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		ctx.MarkAlive()
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start heartbeat
	go h.heartbeat(ctx)

	// Handle incoming messages
	h.readPump(ctx)
}

// heartbeat sends periodic ping messages to keep the connection alive.
func (h *Handler) heartbeat(ctx *platformws.ConnectionContext) {
	ticker := time.NewTicker(pingPeriod)
	ctx.SetHeartbeatTicker(ticker)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !ctx.CheckAlive() {
				h.logger.Debug("connection dead, closing")
				ctx.Conn.Close()
				return
			}
			ctx.MarkDead()
			ctx.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ctx.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// readPump reads messages from the WebSocket connection.
func (h *Handler) readPump(ctx *platformws.ConnectionContext) {
	for {
		var msg stagesdomain.ClientMessage
		if err := ctx.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("unexpected close error", err)
			}
			return
		}

		h.handleMessage(context.Background(), ctx, &msg)
	}
}

// cleanup handles connection cleanup when a client disconnects.
func (h *Handler) cleanup(ctx *platformws.ConnectionContext) {
	ctx.Close()

	if ctx.GroupID != "" && ctx.MemberID != "" {
		h.logger.WithValue("stageID", ctx.GroupID).WithValue("participantID", ctx.MemberID).Debug("participant disconnected")

		// Mark participant as left in database
		if err := h.stagesRepo.SetParticipantLeft(context.Background(), ctx.MemberID); err != nil {
			h.logger.Error("failed to mark participant as left", err)
		}

		// Remove from connection manager
		h.connections.Remove(ctx.GroupID, ctx.MemberID)

		// Broadcast updated state to remaining participants
		h.broadcastStageState(ctx.GroupID)

		// Archive stage if no more connections
		if !h.connections.GroupHasConnections(ctx.GroupID) {
			h.logger.WithValue("stageID", ctx.GroupID).Debug("no more connections, archiving stage")
			if err := h.stagesRepo.ArchiveStage(context.Background(), ctx.GroupID); err != nil {
				h.logger.Error("failed to archive stage", err)
			}
		}
	}
}

// handleMessage routes incoming messages to the appropriate handler.
func (h *Handler) handleMessage(ctx context.Context, connCtx *platformws.ConnectionContext, msg *stagesdomain.ClientMessage) {
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
		h.handleStartPlay(ctx, connCtx, msg)
	case stagesdomain.ClientMessageTypeNextLine:
		h.handleNextLine(ctx, connCtx, msg)
	case stagesdomain.ClientMessageTypeResetPlay:
		h.handleResetPlay(ctx, connCtx, msg)
	default:
		h.sendError(connCtx, "unknown message type")
	}
}

// sendError sends an error message to the client.
func (h *Handler) sendError(ctx *platformws.ConnectionContext, message string) {
	_ = ctx.WriteJSON(&stagesdomain.ServerMessage{
		Type:  stagesdomain.ServerMessageTypeError,
		Error: message,
	})
}

// broadcastStageState broadcasts the current stage state to all connected clients.
func (h *Handler) broadcastStageState(stageID string) {
	stage, err := h.stagesRepo.GetStageByID(context.Background(), stageID)
	if err != nil {
		h.logger.Error("failed to get stage for broadcast", err)
		return
	}

	stageState := h.buildStageState(context.Background(), stage)

	h.connections.BroadcastToGroup(stageID, &stagesdomain.ServerMessage{
		Type:  stagesdomain.ServerMessageTypeStageState,
		State: stageState,
	})
}

// buildStageState constructs a StageState from a Stage.
func (h *Handler) buildStageState(ctx context.Context, stage *stagesdomain.Stage) *stagesdomain.StageState {
	state := &stagesdomain.StageState{
		ID:               stage.ID,
		Code:             stage.Code,
		Participants:     stage.Participants,
		CurrentLineIndex: stage.CurrentLineIndex,
		Phase:            stage.Phase,
	}

	// Fetch selected play if set
	if stage.PlayScriptID != nil && *stage.PlayScriptID != "" {
		play, err := h.playScriptsRepo.GetPlayScript(ctx, *stage.PlayScriptID)
		if err == nil {
			state.SelectedPlay = play
		}
	}

	return state
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
