package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"workflow-engine/backend/internal/repository"
)

// WriteJSON writes data as JSON with the given HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response as JSON.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// MapServiceError converts a service/repository error into an HTTP status + message.
func MapServiceError(err error) (int, string) {
	if errors.Is(err, repository.ErrNotFound) {
		return http.StatusNotFound, "resource not found"
	}

	msg := err.Error()

	if strings.HasPrefix(msg, "validation:") {
		return http.StatusBadRequest, msg
	}

	// Business rule violations
	if strings.HasPrefix(msg, "cannot ") ||
		strings.HasPrefix(msg, "can only ") ||
		strings.HasPrefix(msg, "invalid status transition") ||
		strings.HasPrefix(msg, "completion rule not satisfied") {
		return http.StatusConflict, msg
	}

	return http.StatusInternalServerError, "internal server error"
}
