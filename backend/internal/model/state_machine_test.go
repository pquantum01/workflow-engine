package model

import (
	"strings"
	"testing"
)

func validStateMachine() StateMachine {
	return StateMachine{
		States: []State{
			{Name: "open", Description: "Ticket is open"},
			{Name: "in_progress", Description: "Work in progress"},
			{Name: "done", Description: "Work completed"},
		},
		Transitions: []Transition{
			{Name: "start", FromState: "open", ToState: "in_progress"},
			{Name: "complete", FromState: "in_progress", ToState: "done"},
		},
		InitialState:   "open",
		TerminalStates: []string{"done"},
	}
}

func TestStateMachine_Validate_Valid(t *testing.T) {
	sm := validStateMachine()
	if err := sm.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStateMachine_Validate_NoStates(t *testing.T) {
	sm := validStateMachine()
	sm.States = nil
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for no states")
	}
	if !strings.Contains(err.Error(), "at least one state") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStateMachine_Validate_EmptyStateName(t *testing.T) {
	sm := validStateMachine()
	sm.States[0].Name = ""
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for empty state name")
	}
}

func TestStateMachine_Validate_DuplicateStateName(t *testing.T) {
	sm := validStateMachine()
	sm.States[1].Name = "open"
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for duplicate state name")
	}
	if !strings.Contains(err.Error(), "duplicate state name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStateMachine_Validate_EmptyInitialState(t *testing.T) {
	sm := validStateMachine()
	sm.InitialState = ""
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for empty initial state")
	}
}

func TestStateMachine_Validate_InvalidInitialState(t *testing.T) {
	sm := validStateMachine()
	sm.InitialState = "nonexistent"
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for invalid initial state")
	}
	if !strings.Contains(err.Error(), "does not reference a valid state") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStateMachine_Validate_InvalidTerminalState(t *testing.T) {
	sm := validStateMachine()
	sm.TerminalStates = []string{"nonexistent"}
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for invalid terminal state")
	}
}

func TestStateMachine_Validate_EmptyTransitionName(t *testing.T) {
	sm := validStateMachine()
	sm.Transitions[0].Name = ""
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for empty transition name")
	}
}

func TestStateMachine_Validate_DuplicateTransitionName(t *testing.T) {
	sm := validStateMachine()
	sm.Transitions[1].Name = "start"
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for duplicate transition name")
	}
}

func TestStateMachine_Validate_InvalidFromState(t *testing.T) {
	sm := validStateMachine()
	sm.Transitions[0].FromState = "nonexistent"
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for invalid from_state")
	}
}

func TestStateMachine_Validate_InvalidToState(t *testing.T) {
	sm := validStateMachine()
	sm.Transitions[0].ToState = "nonexistent"
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for invalid to_state")
	}
}

func TestStateMachine_Validate_WithCompletionRule(t *testing.T) {
	sm := validStateMachine()
	sm.Transitions[0].CompletionRule = &CompletionRule{
		Type:     "all_of",
		TaskRefs: []string{"task_a", "task_b"},
	}
	if err := sm.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestStateMachine_Validate_InvalidCompletionRule(t *testing.T) {
	sm := validStateMachine()
	sm.Transitions[0].CompletionRule = &CompletionRule{
		Type:     "invalid_type",
		TaskRefs: []string{"task_a"},
	}
	err := sm.Validate()
	if err == nil {
		t.Fatal("expected error for invalid completion rule type")
	}
}

// CompletionRule tests

func TestCompletionRule_Validate_Single(t *testing.T) {
	cr := &CompletionRule{Type: "single", TaskRefs: []string{"task_a"}}
	if err := cr.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCompletionRule_Validate_SingleMultipleRefs(t *testing.T) {
	cr := &CompletionRule{Type: "single", TaskRefs: []string{"task_a", "task_b"}}
	err := cr.Validate()
	if err == nil {
		t.Fatal("expected error for single rule with multiple task_refs")
	}
}

func TestCompletionRule_Validate_AllOf(t *testing.T) {
	cr := &CompletionRule{Type: "all_of", TaskRefs: []string{"task_a", "task_b"}}
	if err := cr.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCompletionRule_Validate_AnyOf(t *testing.T) {
	cr := &CompletionRule{Type: "any_of", TaskRefs: []string{"task_a", "task_b"}}
	if err := cr.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCompletionRule_Validate_NOfM_Valid(t *testing.T) {
	cr := &CompletionRule{Type: "n_of_m", TaskRefs: []string{"a", "b", "c"}, N: 2}
	if err := cr.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCompletionRule_Validate_NOfM_ZeroN(t *testing.T) {
	cr := &CompletionRule{Type: "n_of_m", TaskRefs: []string{"a", "b"}, N: 0}
	err := cr.Validate()
	if err == nil {
		t.Fatal("expected error for n=0")
	}
}

func TestCompletionRule_Validate_NOfM_NGreaterThanM(t *testing.T) {
	cr := &CompletionRule{Type: "n_of_m", TaskRefs: []string{"a", "b"}, N: 3}
	err := cr.Validate()
	if err == nil {
		t.Fatal("expected error for n > len(task_refs)")
	}
}

func TestCompletionRule_Validate_EmptyTaskRefs(t *testing.T) {
	cr := &CompletionRule{Type: "all_of", TaskRefs: nil}
	err := cr.Validate()
	if err == nil {
		t.Fatal("expected error for empty task_refs")
	}
}

func TestCompletionRule_Validate_InvalidType(t *testing.T) {
	cr := &CompletionRule{Type: "bogus", TaskRefs: []string{"a"}}
	err := cr.Validate()
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}
