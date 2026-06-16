package dto

import (
	"encoding/json"
	"time"
)

type WorkflowDefResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Version     int             `json:"version"`
	Description string          `json:"description"`
	Definition  json.RawMessage `json:"definition,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}

type WorkflowDefListResponse struct {
	Items []*WorkflowDefResponse `json:"items"`
	Total int                    `json:"total"`
}

type ReloadDefsResponse struct {
	Loaded   []ReloadedDef   `json:"loaded"`
	Warnings []ReloadWarning `json:"warnings"`
	Errors   []ReloadError   `json:"errors"`
}

type ReloadedDef struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type ReloadWarning struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
	File    string `json:"file"`
	Message string `json:"message"`
}

type ReloadError struct {
	File    string `json:"file"`
	Message string `json:"message"`
}
