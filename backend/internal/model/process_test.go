package model

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validProcess() *Process {
	now := time.Now()
	return &Process{
		ID:          uuid.New(),
		Name:        "Bug Triage",
		Description: "Process for triaging bugs",
		Version:     1,
		Status:      "draft",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestProcess_Validate_Valid(t *testing.T) {
	p := validProcess()
	if err := p.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProcess_Validate_AllStatuses(t *testing.T) {
	for _, status := range []string{"draft", "active", "archived"} {
		p := validProcess()
		p.Status = status
		if err := p.Validate(); err != nil {
			t.Errorf("status %q should be valid, got %v", status, err)
		}
	}
}

func TestProcess_Validate_EmptyName(t *testing.T) {
	p := validProcess()
	p.Name = ""
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcess_Validate_LongName(t *testing.T) {
	p := validProcess()
	p.Name = strings.Repeat("a", 256)
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for long name")
	}
	if !strings.Contains(err.Error(), "255") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcess_Validate_InvalidStatus(t *testing.T) {
	p := validProcess()
	p.Status = "invalid"
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestProcess_Validate_ZeroVersion(t *testing.T) {
	p := validProcess()
	p.Version = 0
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for zero version")
	}
}

func TestProcess_Validate_NegativeVersion(t *testing.T) {
	p := validProcess()
	p.Version = -1
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for negative version")
	}
}
