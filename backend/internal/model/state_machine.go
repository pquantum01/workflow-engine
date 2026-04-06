package model

import "fmt"

type StateMachine struct {
	States         []State      `json:"states"`
	Transitions    []Transition `json:"transitions"`
	InitialState   string       `json:"initial_state"`
	TerminalStates []string     `json:"terminal_states"`
}

type State struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type Transition struct {
	Name           string          `json:"name"`
	FromState      string          `json:"from_state"`
	ToState        string          `json:"to_state"`
	CompletionRule *CompletionRule `json:"completion_rule,omitempty"`
}

type CompletionRule struct {
	Type     string   `json:"type"`
	TaskRefs []string `json:"task_refs"`
	N        int      `json:"n,omitempty"`
}

var validCompletionRuleTypes = map[string]bool{
	"single": true,
	"all_of": true,
	"any_of": true,
	"n_of_m": true,
}

func (sm *StateMachine) Validate() error {
	if len(sm.States) == 0 {
		return fmt.Errorf("at least one state is required")
	}

	stateNames := make(map[string]bool, len(sm.States))
	for _, s := range sm.States {
		if s.Name == "" {
			return fmt.Errorf("state name is required")
		}
		if stateNames[s.Name] {
			return fmt.Errorf("duplicate state name: %s", s.Name)
		}
		stateNames[s.Name] = true
	}

	if sm.InitialState == "" {
		return fmt.Errorf("initial_state is required")
	}
	if !stateNames[sm.InitialState] {
		return fmt.Errorf("initial_state %q does not reference a valid state", sm.InitialState)
	}

	for _, ts := range sm.TerminalStates {
		if !stateNames[ts] {
			return fmt.Errorf("terminal_state %q does not reference a valid state", ts)
		}
	}

	transitionNames := make(map[string]bool, len(sm.Transitions))
	for _, t := range sm.Transitions {
		if t.Name == "" {
			return fmt.Errorf("transition name is required")
		}
		if transitionNames[t.Name] {
			return fmt.Errorf("duplicate transition name: %s", t.Name)
		}
		transitionNames[t.Name] = true

		if !stateNames[t.FromState] {
			return fmt.Errorf("transition %q from_state %q does not reference a valid state", t.Name, t.FromState)
		}
		if !stateNames[t.ToState] {
			return fmt.Errorf("transition %q to_state %q does not reference a valid state", t.Name, t.ToState)
		}

		if t.CompletionRule != nil {
			if err := t.CompletionRule.Validate(); err != nil {
				return fmt.Errorf("transition %q completion_rule: %w", t.Name, err)
			}
		}
	}

	return nil
}

func (cr *CompletionRule) Validate() error {
	if !validCompletionRuleTypes[cr.Type] {
		return fmt.Errorf("type must be one of: single, all_of, any_of, n_of_m")
	}
	if len(cr.TaskRefs) == 0 {
		return fmt.Errorf("task_refs must not be empty")
	}
	if cr.Type == "single" && len(cr.TaskRefs) != 1 {
		return fmt.Errorf("single completion rule must have exactly 1 task_ref")
	}
	if cr.Type == "n_of_m" {
		if cr.N <= 0 {
			return fmt.Errorf("n must be > 0 for n_of_m rule")
		}
		if cr.N > len(cr.TaskRefs) {
			return fmt.Errorf("n (%d) must be <= number of task_refs (%d)", cr.N, len(cr.TaskRefs))
		}
	}
	return nil
}
