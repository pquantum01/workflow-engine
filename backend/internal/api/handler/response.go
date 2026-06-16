package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string, details any) {
	writeJSON(w, status, ErrorResponse{Error: ErrorBody{Code: code, Message: message, Details: details}})
}

func writeNotFound(w http.ResponseWriter) {
	writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
}

func writeInternalError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, "INTERNAL", "internal server error", nil)
}
