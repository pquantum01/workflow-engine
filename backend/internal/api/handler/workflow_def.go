package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/workflow-engine/backend/internal/api/dto"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/service"
)


type WorkflowDefHandler struct {
	svc *service.WorkflowDefService
}

func NewWorkflowDefHandler(svc *service.WorkflowDefService) *WorkflowDefHandler {
	return &WorkflowDefHandler{svc: svc}
}

func (h *WorkflowDefHandler) List(w http.ResponseWriter, r *http.Request) {
	includeDeleted, _ := strconv.ParseBool(r.URL.Query().Get("include_deleted"))
	defs, err := h.svc.List(r.Context(), includeDeleted)
	if err != nil {
		writeInternalError(w)
		return
	}
	items := make([]*dto.WorkflowDefResponse, 0, len(defs))
	for _, d := range defs {
		items = append(items, toDefResponse(d, false))
	}
	writeJSON(w, http.StatusOK, dto.WorkflowDefListResponse{Items: items, Total: len(items)})
}

func (h *WorkflowDefHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	def, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeNotFound(w)
			return
		}
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, toDefResponse(def, true))
}

func (h *WorkflowDefHandler) ReloadDefs(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.Reload(r.Context())
	if err != nil {
		writeInternalError(w)
		return
	}
	resp := dto.ReloadDefsResponse{
		Loaded:   make([]dto.ReloadedDef, 0),
		Warnings: make([]dto.ReloadWarning, 0),
		Errors:   make([]dto.ReloadError, 0),
	}
	for _, l := range result.Loaded {
		resp.Loaded = append(resp.Loaded, dto.ReloadedDef{Name: l.Name, Version: l.Version})
	}
	for _, w2 := range result.Warnings {
		resp.Warnings = append(resp.Warnings, dto.ReloadWarning{Name: w2.Name, Version: w2.Version, File: w2.File, Message: w2.Message})
	}
	for _, e := range result.Errors {
		resp.Errors = append(resp.Errors, dto.ReloadError{File: e.File, Message: e.Message})
	}
	writeJSON(w, http.StatusOK, resp)
}

func toDefResponse(d *model.WorkflowDef, includeDefinition bool) *dto.WorkflowDefResponse {
	resp := &dto.WorkflowDefResponse{
		ID:          d.ID.String(),
		Name:        d.Name,
		Version:     d.Version,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		DeletedAt:   d.DeletedAt,
	}
	if includeDefinition {
		b, _ := json.Marshal(d.Definition)
		resp.Definition = b
	}
	return resp
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

