package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/workflow-engine/backend/internal/api/dto"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/service"
)

type TaskHandler struct {
	svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	task, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeNotFound(w)
			return
		}
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) StartTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	task, err := h.svc.Start(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeNotFound(w)
			return
		}
		if errors.Is(err, service.ErrTaskNotCompletable) {
			writeError(w, http.StatusConflict, "TASK_NOT_COMPLETABLE", "task cannot be started", nil)
			return
		}
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) Complete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	var req dto.CompleteTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}
	if req.Status == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "status is required", nil)
		return
	}

	inst, _, err := h.svc.Complete(r.Context(), id, model.TaskStatus(req.Status), req.NextState, req.CompletedBy)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeNotFound(w)
		case errors.Is(err, service.ErrInstanceTerminal):
			writeError(w, http.StatusConflict, "INSTANCE_TERMINAL", "instance is in terminal state", nil)
		case errors.Is(err, service.ErrTaskNotCompletable):
			writeError(w, http.StatusConflict, "TASK_NOT_COMPLETABLE", "task is not in a completable state", nil)
		case errors.Is(err, service.ErrInvalidNextState):
			writeError(w, http.StatusUnprocessableEntity, "INVALID_NEXT_STATE", err.Error(), nil)
		case errors.Is(err, service.ErrValidation):
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid status value", nil)
		default:
			writeInternalError(w)
		}
		return
	}
	writeJSON(w, http.StatusOK, toInstanceResponse(inst, nil, nil))
}

func (h *TaskHandler) ListTaskTypes(w http.ResponseWriter, r *http.Request) {
	types := []map[string]string{
		{"type": "approval", "description": "Done/rejected with optional comment"},
		{"type": "form", "description": "Collect structured data on completion"},
		{"type": "checklist", "description": "Multi-item, done when all checked"},
		{"type": "manual", "description": "Generic, mark done or rejected"},
	}
	writeJSON(w, http.StatusOK, types)
}

func toTaskResponse(t *model.TaskInstance) *dto.TaskInstanceResponse {
	return &dto.TaskInstanceResponse{
		ID:                 t.ID.String(),
		WorkflowInstanceID: t.WorkflowInstanceID.String(),
		StateID:            t.StateID,
		TaskKey:            t.TaskKey,
		TaskLabel:          t.TaskLabel,
		TaskType:           t.TaskType,
		Status:             string(t.Status),
		Outcome:            t.Outcome,
		CreatedBy:          t.CreatedBy,
		CreatedAt:          t.CreatedAt,
		CompletedAt:        t.CompletedAt,
	}
}
