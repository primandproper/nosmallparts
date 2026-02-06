package livekit

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/internalerrors"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/livekit/protocol/auth"
)

// TokenRequest represents the request body for token generation.
type TokenRequest struct {
	StageID         string `json:"stageId"`
	ParticipantID   string `json:"participantId"`
	ParticipantName string `json:"participantName"`
}

// TokenResponse represents the response body for token generation.
type TokenResponse struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Handler handles LiveKit token generation requests.
type Handler struct {
	config *Config
	logger logging.Logger
	tracer tracing.Tracer
}

// NewHandler creates a new LiveKit handler.
func NewHandler(
	ctx context.Context,
	config *Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) (*Handler, error) {
	if config == nil {
		return nil, internalerrors.NilConfigError("livekit")
	}

	if err := config.ValidateWithContext(ctx); err != nil {
		return nil, err
	}

	return &Handler{
		config: config,
		logger: logging.EnsureLogger(logger).WithName("livekit"),
		tracer: tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer("livekit")),
	}, nil
}

// Enabled returns true if LiveKit is configured.
func (h *Handler) Enabled() bool {
	return h.config.Enabled()
}

// GenerateToken handles POST requests to generate a LiveKit access token.
func (h *Handler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	logger := h.logger.Clone()

	// Check if LiveKit is configured
	if !h.config.Enabled() {
		logger.Debug("LiveKit not configured, returning 503")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "LiveKit not configured"})
		return
	}

	// Parse request body
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate required fields
	if req.StageID == "" || req.ParticipantID == "" || req.ParticipantName == "" {
		logger.Debug("missing required fields in token request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "stageId, participantId, and participantName are required"})
		return
	}

	logger = logger.WithValue("stageId", req.StageID).WithValue("participantId", req.ParticipantID)
	tracing.AttachToSpan(span, "stage_id", req.StageID)
	tracing.AttachToSpan(span, "participant_id", req.ParticipantID)

	// Generate LiveKit access token
	// The room name is based on the stage ID to group participants
	roomName := "stage-" + req.StageID

	at := auth.NewAccessToken(h.config.APIKey, h.config.APISecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.AddGrant(grant).
		SetIdentity(req.ParticipantID).
		SetName(req.ParticipantName).
		SetValidFor(24 * time.Hour) // Token valid for 24 hours

	token, err := at.ToJWT()
	if err != nil {
		logger.Error("failed to generate JWT token", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to generate token"})
		return
	}

	logger.Debug("generated LiveKit token")

	// Return token and LiveKit server URL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(TokenResponse{
		Token: token,
		URL:   h.config.URL,
	})
}
