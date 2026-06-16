package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/workflow-engine/backend/internal/api/dto"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
	"github.com/workflow-engine/backend/internal/service"
)

type InstanceHandler struct {
	svc *service.InstanceService
}

func NewInstanceHandler(svc *service.InstanceService) *InstanceHandler {
	return &InstanceHandler{svc: svc}
}

func (h *InstanceHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req dto.StartInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}
	if req.WorkflowDefID == "" || req.CreatedBy == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "workflow_def_id and created_by are required", nil)
		return
	}
	defID, err := parseUUID(req.WorkflowDefID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid workflow_def_id", nil)
		return
	}

	inst, _, err := h.svc.Start(r.Context(), defID, req.CreatedBy)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeNotFound(w)
			return
		}
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusCreated, toInstanceResponse(inst, nil, nil))
}

func (h *InstanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	inst, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeNotFound(w)
			return
		}
		writeInternalError(w)
		return
	}
	tasks, _ := h.svc.ListTasks(r.Context(), id, nil)
	history, _ := h.svc.ListHistory(r.Context(), id)
	writeJSON(w, http.StatusOK, toInstanceResponse(inst, tasks, history))
}

func (h *InstanceHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := repository.InstanceFilter{}

	if defIDStr := q.Get("def_id"); defIDStr != "" {
		if id, err := parseUUID(defIDStr); err == nil {
			filter.DefID = &id
		}
	}
	if statusStr := q.Get("status"); statusStr != "" {
		s := model.InstanceStatus(statusStr)
		filter.Status = &s
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit <= 0 {
		limit = 50
	}
	filter.Limit = limit
	filter.Offset = offset

	insts, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeInternalError(w)
		return
	}
	items := make([]*dto.WorkflowInstanceResponse, 0, len(insts))
	for _, i := range insts {
		items = append(items, toInstanceResponse(i, nil, nil))
	}
	writeJSON(w, http.StatusOK, dto.InstanceListResponse{Items: items, Total: total, Limit: limit, Offset: offset})
}

func (h *InstanceHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	inst, _, err := h.svc.Cancel(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeNotFound(w)
			return
		}
		if errors.Is(err, service.ErrInstanceTerminal) {
			writeError(w, http.StatusConflict, "INSTANCE_TERMINAL", "instance is already in terminal state", nil)
			return
		}
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, toInstanceResponse(inst, nil, nil))
}

func (h *InstanceHandler) SoftDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	if err := h.svc.SoftDelete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeNotFound(w)
			return
		}
		writeInternalError(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *InstanceHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	var statusFilter []model.TaskStatus
	if s := r.URL.Query().Get("status"); s != "" {
		statusFilter = []model.TaskStatus{model.TaskStatus(s)}
	}
	tasks, err := h.svc.ListTasks(r.Context(), id, statusFilter)
	if err != nil {
		writeInternalError(w)
		return
	}
	items := make([]*dto.TaskInstanceResponse, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, toTaskResponse(t))
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *InstanceHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	transitions, err := h.svc.ListHistory(r.Context(), id)
	if err != nil {
		writeInternalError(w)
		return
	}
	items := make([]*dto.StateTransitionResponse, 0, len(transitions))
	for _, t := range transitions {
		items = append(items, &dto.StateTransitionResponse{
			ID: t.ID.String(), FromState: t.FromState, ToState: t.ToState, TransitionedAt: t.TransitionedAt,
		})
	}
	writeJSON(w, http.StatusOK, items)
}

func toInstanceResponse(inst *model.WorkflowInstance, tasks []*model.TaskInstance, history []*model.StateTransition) *dto.WorkflowInstanceResponse {
	resp := &dto.WorkflowInstanceResponse{
		ID:             inst.ID.String(),
		WorkflowDefID:  inst.WorkflowDefID.String(),
		CurrentStateID: inst.CurrentStateID,
		Status:         string(inst.Status),
		CreatedBy:      inst.CreatedBy,
		Context:        inst.Context,
		StartedAt:      inst.StartedAt,
		CompletedAt:    inst.CompletedAt,
	}
	// Always set non-nil slices so JSON encodes [] not null
	resp.Tasks = make([]*dto.TaskInstanceResponse, 0, len(tasks))
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, toTaskResponse(t))
	}
	resp.History = make([]*dto.StateTransitionResponse, 0, len(history))
	for _, tr := range history {
		resp.History = append(resp.History, &dto.StateTransitionResponse{
			ID: tr.ID.String(), FromState: tr.FromState, ToState: tr.ToState, TransitionedAt: tr.TransitionedAt,
		})
	}
	return resp
}
