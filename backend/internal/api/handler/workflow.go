package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"workflow-engine/backend/internal/service"
)

type WorkflowHandler struct {
	service *service.WorkflowService
	logger  *slog.Logger
}

func NewWorkflowHandler(svc *service.WorkflowService, logger *slog.Logger) *WorkflowHandler {
	return &WorkflowHandler{service: svc, logger: logger}
}

func (h *WorkflowHandler) Create(w http.ResponseWriter, r *http.Request) {
	processID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid process id")
		return
	}

	var input service.CreateWorkflowInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wf, err := h.service.Create(r.Context(), processID, input)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusCreated, wf)
}

func (h *WorkflowHandler) ListByProcessID(w http.ResponseWriter, r *http.Request) {
	processID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid process id")
		return
	}

	results, err := h.service.ListByProcessID(r.Context(), processID)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, results)
}

func (h *WorkflowHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	wf, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, wf)
}

func (h *WorkflowHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input service.UpdateWorkflowInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wf, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, wf)
}

func (h *WorkflowHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
