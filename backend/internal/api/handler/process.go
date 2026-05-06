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

type ProcessHandler struct {
	service *service.ProcessService
	logger  *slog.Logger
}

func NewProcessHandler(svc *service.ProcessService, logger *slog.Logger) *ProcessHandler {
	return &ProcessHandler{service: svc, logger: logger}
}

func (h *ProcessHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateProcessInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.service.Create(r.Context(), input)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusCreated, p)
}

func (h *ProcessHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	p, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, p)
}

func (h *ProcessHandler) List(w http.ResponseWriter, r *http.Request) {
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

func (h *ProcessHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input service.UpdateProcessInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		status, msg := MapServiceError(err)
		WriteError(w, status, msg)
		return
	}

	WriteJSON(w, http.StatusOK, p)
}

func (h *ProcessHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
