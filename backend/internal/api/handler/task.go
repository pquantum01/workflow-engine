package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"workflow-engine/backend/internal/service"
)

type TaskHandler struct {
	service *service.TaskService
	logger  *slog.Logger
}

func NewTaskHandler(svc *service.TaskService, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{service: svc, logger: logger}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid workflow id")
		return
	}

	var input service.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.service.Create(r.Context(), workflowID, input)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusCreated, t)
}

func (h *TaskHandler) ListByWorkflowID(w http.ResponseWriter, r *http.Request) {
	workflowID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid workflow id")
		return
	}

	results, err := h.service.ListByWorkflowID(r.Context(), workflowID)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, results)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	t, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, t)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input service.UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, t)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
