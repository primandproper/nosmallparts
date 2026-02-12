package playscripts

import (
	"errors"
	"net/http"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/go-chi/chi/v5"
)

// errorResponse is the JSON shape for error responses.
type errorResponse struct {
	Error string `json:"error"`
}

// HTTPHandler handles HTTP requests for play scripts.
type HTTPHandler struct {
	manager        playscripts.PlayScriptsManager
	encoderDecoder encoding.ServerEncoderDecoder
	logger         logging.Logger
	tracer         tracing.Tracer
}

// NewHTTPHandler creates a new HTTP handler for play scripts.
func NewHTTPHandler(
	manager playscripts.PlayScriptsManager,
	encoderDecoder encoding.ServerEncoderDecoder,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
) *HTTPHandler {
	return &HTTPHandler{
		manager:        manager,
		encoderDecoder: encoderDecoder,
		logger:         logging.EnsureLogger(logger).WithName("play_scripts_http"),
		tracer:         tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer("play_scripts_http")),
	}
}

// CreatePlayScript handles POST /api/play-scripts (auth required).
func (h *HTTPHandler) CreatePlayScript(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	session, err := sessions.FetchContextDataFromRequest(r)
	if err != nil {
		h.logger.Error("create play script: unauthenticated", err)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "unauthenticated"}, http.StatusUnauthorized)
		return
	}

	var input playscripts.PlayScriptCreationRequestInput
	if decodeErr := h.encoderDecoder.DecodeRequest(ctx, r, &input); decodeErr != nil {
		h.logger.Error("create play script: invalid body", decodeErr)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "invalid request body"}, http.StatusBadRequest)
		return
	}

	created, err := h.manager.CreatePlayScript(ctx, &input, session.GetActiveAccountID(), session.GetUserID())
	if err != nil {
		h.logger.Error("create play script", err)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "failed to create play script"}, http.StatusInternalServerError)
		return
	}

	h.encoderDecoder.EncodeResponseWithStatus(ctx, w, created, http.StatusCreated)
}

// GetPlayScript handles GET /api/play-scripts/{id} (auth optional; manager enforces public vs account).
func (h *HTTPHandler) GetPlayScript(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "id is required"}, http.StatusBadRequest)
		return
	}

	accountID := ""
	if session, err := sessions.FetchContextDataFromRequest(r); err == nil {
		accountID = session.GetActiveAccountID()
	}

	play, err := h.manager.GetPlayScript(ctx, id, accountID)
	if err != nil {
		if errors.Is(err, ErrPermissionDenied) {
			h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "permission denied"}, http.StatusForbidden)
			return
		}
		h.logger.Error("get play script", err)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "failed to get play script"}, http.StatusInternalServerError)
		return
	}

	h.encoderDecoder.EncodeResponseWithStatus(ctx, w, play, http.StatusOK)
}

// GetPlayScriptsForAccount handles GET /api/play-scripts or GET /api/play-scripts/account (auth required).
func (h *HTTPHandler) GetPlayScriptsForAccount(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	session, err := sessions.FetchContextDataFromRequest(r)
	if err != nil {
		h.logger.Error("get play scripts for account: unauthenticated", err)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "unauthenticated"}, http.StatusUnauthorized)
		return
	}

	filter := filtering.ExtractQueryFilterFromRequest(r)
	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	result, err := h.manager.GetPlayScriptsForAccount(ctx, session.GetActiveAccountID(), filter)
	if err != nil {
		h.logger.Error("get play scripts for account", err)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "failed to list play scripts"}, http.StatusInternalServerError)
		return
	}

	h.encoderDecoder.EncodeResponseWithStatus(ctx, w, result, http.StatusOK)
}

// GetPublicPlayScripts handles GET /api/play-scripts/public (auth optional).
func (h *HTTPHandler) GetPublicPlayScripts(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	filter := filtering.ExtractQueryFilterFromRequest(r)
	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	result, err := h.manager.GetPublicPlayScripts(ctx, filter)
	if err != nil {
		h.logger.Error("get public play scripts", err)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "failed to list public play scripts"}, http.StatusInternalServerError)
		return
	}

	h.encoderDecoder.EncodeResponseWithStatus(ctx, w, result, http.StatusOK)
}

// UpdatePlayScript handles PATCH /api/play-scripts/{id} (auth required).
func (h *HTTPHandler) UpdatePlayScript(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	session, err := sessions.FetchContextDataFromRequest(r)
	if err != nil {
		h.logger.Error("update play script: unauthenticated", err)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "unauthenticated"}, http.StatusUnauthorized)
		return
	}

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "id is required"}, http.StatusBadRequest)
		return
	}

	var input playscripts.PlayScriptUpdateRequestInput
	if decodeErr := h.encoderDecoder.DecodeRequest(ctx, r, &input); decodeErr != nil {
		h.logger.Error("update play script: invalid body", decodeErr)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "invalid request body"}, http.StatusBadRequest)
		return
	}

	updated, err := h.manager.UpdatePlayScript(ctx, id, &input, session.GetActiveAccountID())
	if err != nil {
		if errors.Is(err, ErrPermissionDenied) {
			h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "permission denied"}, http.StatusForbidden)
			return
		}
		h.logger.Error("update play script", err)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "failed to update play script"}, http.StatusInternalServerError)
		return
	}

	h.encoderDecoder.EncodeResponseWithStatus(ctx, w, updated, http.StatusOK)
}

// ArchivePlayScript handles DELETE /api/play-scripts/{id} (auth required).
func (h *HTTPHandler) ArchivePlayScript(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	session, err := sessions.FetchContextDataFromRequest(r)
	if err != nil {
		h.logger.Error("archive play script: unauthenticated", err)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "unauthenticated"}, http.StatusUnauthorized)
		return
	}

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "id is required"}, http.StatusBadRequest)
		return
	}

	if archiveErr := h.manager.ArchivePlayScript(ctx, id, session.GetActiveAccountID()); archiveErr != nil {
		if errors.Is(archiveErr, ErrPermissionDenied) {
			h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "permission denied"}, http.StatusForbidden)
			return
		}
		h.logger.Error("archive play script", archiveErr)
		h.encoderDecoder.EncodeResponseWithStatus(ctx, w, &errorResponse{Error: "failed to archive play script"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
