package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"workflow-engine/backend/internal/service"
)

type WorkflowInstanceHandler struct {
	service *service.WorkflowInstanceService
	logger  *slog.Logger
}

func NewWorkflowInstanceHandler(svc *service.WorkflowInstanceService, logger *slog.Logger) *WorkflowInstanceHandler {
	return &WorkflowInstanceHandler{service: svc, logger: logger}
}

func (h *WorkflowInstanceHandler) ListByProcessInstanceID(w http.ResponseWriter, r *http.Request) {
	processInstanceID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid process instance id")
		return
	}

	results, err := h.service.ListByProcessInstanceID(r.Context(), processInstanceID)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, results)
}

func (h *WorkflowInstanceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	wi, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, wi)
}

func (h *WorkflowInstanceHandler) ListTransitions(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	results, err := h.service.ListTransitions(r.Context(), id)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, results)
}

func (h *WorkflowInstanceHandler) TriggerTransition(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input service.TriggerTransitionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.TriggerTransition(r.Context(), id, input)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}
