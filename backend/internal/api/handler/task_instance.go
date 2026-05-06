package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"workflow-engine/backend/internal/service"
)

type TaskInstanceHandler struct {
	service *service.TaskInstanceService
	logger  *slog.Logger
}

func NewTaskInstanceHandler(svc *service.TaskInstanceService, logger *slog.Logger) *TaskInstanceHandler {
	return &TaskInstanceHandler{service: svc, logger: logger}
}

func (h *TaskInstanceHandler) ListByWorkflowInstanceID(w http.ResponseWriter, r *http.Request) {
	workflowInstanceID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid workflow instance id")
		return
	}

	results, err := h.service.ListByWorkflowInstanceID(r.Context(), workflowInstanceID)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, results)
}

func (h *TaskInstanceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	ti, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, ti)
}

func (h *TaskInstanceHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input service.UpdateTaskStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ti, err := h.service.UpdateStatus(r.Context(), id, input)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, ti)
}

type assignInput struct {
	AssignedTo string `json:"assigned_to"`
}

func (h *TaskInstanceHandler) Assign(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input assignInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ti, err := h.service.Assign(r.Context(), id, input.AssignedTo)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, ti)
}
