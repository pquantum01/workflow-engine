package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"workflow-engine/backend/internal/service"
)

type ProcessInstanceHandler struct {
	service *service.ProcessInstanceService
	logger  *slog.Logger
}

func NewProcessInstanceHandler(svc *service.ProcessInstanceService, logger *slog.Logger) *ProcessInstanceHandler {
	return &ProcessInstanceHandler{service: svc, logger: logger}
}

func (h *ProcessInstanceHandler) Start(w http.ResponseWriter, r *http.Request) {
	var input service.StartProcessInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pi, err := h.service.Start(r.Context(), input)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusCreated, pi)
}

func (h *ProcessInstanceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	pi, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, pi)
}

func (h *ProcessInstanceHandler) List(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	results, err := h.service.List(r.Context(), offset, limit)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, results)
}

func (h *ProcessInstanceHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.Cancel(r.Context(), id); err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProcessInstanceHandler) ListChildren(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	results, err := h.service.ListChildren(r.Context(), id)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, results)
}
