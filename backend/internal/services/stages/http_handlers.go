package stages

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"

	"github.com/go-chi/chi/v5"
)

// CreateStageResponse is the response body for stage creation.
type CreateStageResponse struct {
	StageID string `json:"stageId"`
	Code    string `json:"code"`
}

// StageExistsResponse is the response body for stage existence check.
type StageExistsResponse struct {
	Exists bool `json:"exists"`
}

// CreateStageHTTP handles POST /api/stages to create a new stage.
func (h *Handler) CreateStageHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	stage, err := h.stagesManager.CreateStage(ctx, &stagesdomain.StageCreationInput{
		ID:   identifiers.New(),
		Code: generateStageCode(),
	})
	if err != nil {
		h.logger.Error("failed to create stage", err)
		http.Error(w, `{"error":"failed to create stage"}`, http.StatusInternalServerError)
		return
	}

	h.logger.WithValue("stageID", stage.ID).WithValue("code", stage.Code).Debug("created stage via HTTP")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(&CreateStageResponse{
		StageID: stage.ID,
		Code:    stage.Code,
	}); err != nil {
		h.logger.Error("failed to encode response", err)
	}
}

// CheckStageExists handles GET /api/stages/{code}/exists to check if a stage exists.
func (h *Handler) CheckStageExists(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		http.Error(w, `{"error":"code is required"}`, http.StatusBadRequest)
		return
	}

	_, err := h.stagesManager.GetStageByCode(ctx, code)

	exists := true
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			exists = false
		} else {
			h.logger.Error("failed to check stage existence", err)
			http.Error(w, `{"error":"failed to check stage existence"}`, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(&StageExistsResponse{
		Exists: exists,
	}); err != nil {
		h.logger.Error("failed to encode response", err)
	}
}
