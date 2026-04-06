package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var validProcessStatuses = map[string]bool{
	"draft":    true,
	"active":   true,
	"archived": true,
}

type Process struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *Process) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(p.Name) > 255 {
		return fmt.Errorf("name must be at most 255 characters")
	}
	if !validProcessStatuses[p.Status] {
		return fmt.Errorf("status must be one of: draft, active, archived")
	}
	if p.Version < 1 {
		return fmt.Errorf("version must be >= 1")
	}
	return nil
}
