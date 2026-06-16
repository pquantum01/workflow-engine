package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var slugRe = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

type WorkflowDef struct {
	ID          uuid.UUID
	Name        string
	Version     int
	Description string
	Definition  WorkflowDefinition
	CreatedAt   time.Time
	DeletedAt   *time.Time
}

type WorkflowDefinition struct {
	Name        string          `json:"name"`
	Version     int             `json:"version"`
	Description string          `json:"description"`
	StartState  string          `json:"start_state"`
	EndState    string          `json:"end_state"`
	States      []StateDef      `json:"states"`
	Transitions []TransitionDef `json:"transitions"`
}

type StateDef struct {
	ID    string    `json:"id"`
	Label string    `json:"label"`
	Tasks []TaskDef `json:"tasks"`
}

type TaskDef struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

type TransitionDef struct {
	FromState string             `json:"from_state"`
	ToState   string             `json:"to_state"`
	Trigger   *TransitionTrigger `json:"trigger,omitempty"`
}

type TransitionTrigger struct {
	TaskKey string `json:"task_key"`
}

type TransitionMode int

const (
	ModeLinear    TransitionMode = iota
	ModeTaskBound                // all outgoing have trigger.task_key
	ModeUserPicks                // multiple outgoing, no triggers, single task
)

var validTaskTypes = map[string]bool{
	"approval":  true,
	"form":      true,
	"checklist": true,
	"manual":    true,
}

func (d *WorkflowDefinition) Validate() (warnings []string, err error) {
	var errs []string

	if !slugRe.MatchString(d.Name) {
		errs = append(errs, fmt.Sprintf("name %q must match ^[a-z][a-z0-9_-]*$", d.Name))
	}
	if d.Version < 1 {
		errs = append(errs, "version must be >= 1")
	}
	if d.StartState == "" {
		errs = append(errs, "start_state is required")
	}
	if d.EndState == "" {
		errs = append(errs, "end_state is required")
	}

	stateIndex := make(map[string]*StateDef, len(d.States))
	for i := range d.States {
		s := &d.States[i]
		if _, dup := stateIndex[s.ID]; dup {
			errs = append(errs, fmt.Sprintf("duplicate state id %q", s.ID))
			continue
		}
		if !slugRe.MatchString(s.ID) {
			errs = append(errs, fmt.Sprintf("state id %q must match ^[a-z][a-z0-9_-]*$", s.ID))
		}
		if s.Label == "" {
			errs = append(errs, fmt.Sprintf("state %q: label is required", s.ID))
		}
		taskKeys := make(map[string]bool)
		for _, t := range s.Tasks {
			if taskKeys[t.Key] {
				errs = append(errs, fmt.Sprintf("state %q: duplicate task key %q", s.ID, t.Key))
			}
			taskKeys[t.Key] = true
			if !slugRe.MatchString(t.Key) {
				errs = append(errs, fmt.Sprintf("state %q task %q: key must match ^[a-z][a-z0-9_-]*$", s.ID, t.Key))
			}
			if t.Label == "" {
				errs = append(errs, fmt.Sprintf("state %q task %q: label is required", s.ID, t.Key))
			}
			if !validTaskTypes[t.Type] {
				errs = append(errs, fmt.Sprintf("state %q task %q: unknown type %q", s.ID, t.Key, t.Type))
			}
		}
		stateIndex[s.ID] = s
	}

	// check start/end exist
	if d.StartState != "" && stateIndex[d.StartState] == nil {
		errs = append(errs, fmt.Sprintf("start_state %q not found in states", d.StartState))
	}
	if d.EndState != "" {
		if stateIndex[d.EndState] == nil {
			errs = append(errs, fmt.Sprintf("end_state %q not found in states", d.EndState))
		}
	}

	// validate transitions
	outgoing := make(map[string][]TransitionDef)
	for _, tr := range d.Transitions {
		if stateIndex[tr.FromState] == nil {
			errs = append(errs, fmt.Sprintf("transition from_state %q not found", tr.FromState))
		}
		if stateIndex[tr.ToState] == nil {
			errs = append(errs, fmt.Sprintf("transition to_state %q not found", tr.ToState))
		}
		if tr.Trigger != nil && tr.Trigger.TaskKey == "" {
			errs = append(errs, fmt.Sprintf("transition from %q to %q: trigger.task_key is empty", tr.FromState, tr.ToState))
		}
		outgoing[tr.FromState] = append(outgoing[tr.FromState], tr)
	}

	// validate per-state outgoing rules
	for id, s := range stateIndex {
		outs := outgoing[id]
		isTerminal := len(outs) == 0 && len(s.Tasks) == 0

		if isTerminal {
			// valid leaf state (end_state or alternative terminal like "denied")
			continue
		}

		// non-terminal: must have both tasks and outgoing transitions
		if len(outs) == 0 {
			errs = append(errs, fmt.Sprintf("state %q: has tasks but no outgoing transitions", id))
		}
		if len(s.Tasks) == 0 {
			errs = append(errs, fmt.Sprintf("state %q: has outgoing transitions but no tasks", id))
		}

		if len(outs) > 1 {
			triggeredCount := 0
			for _, tr := range outs {
				if tr.Trigger != nil {
					triggeredCount++
				}
			}
			if triggeredCount > 0 && triggeredCount < len(outs) {
				errs = append(errs, fmt.Sprintf("state %q: mixed triggers — all or none outgoing transitions must have a trigger", id))
			}
			if triggeredCount == 0 {
				// user-picks: must have exactly 1 task
				if len(s.Tasks) != 1 {
					errs = append(errs, fmt.Sprintf("state %q: user-picks (multiple outgoing, no triggers) must have exactly 1 task", id))
				}
			} else {
				// task-bound: each trigger.task_key must exist in state tasks, no unreferenced tasks
				taskKeys := make(map[string]bool)
				for _, t := range s.Tasks {
					taskKeys[t.Key] = true
				}
				boundKeys := make(map[string]bool)
				for _, tr := range outs {
					if tr.Trigger != nil {
						if !taskKeys[tr.Trigger.TaskKey] {
							errs = append(errs, fmt.Sprintf("state %q: trigger task_key %q not found in state tasks", id, tr.Trigger.TaskKey))
						}
						if boundKeys[tr.Trigger.TaskKey] {
							errs = append(errs, fmt.Sprintf("state %q: task_key %q bound to multiple transitions", id, tr.Trigger.TaskKey))
						}
						boundKeys[tr.Trigger.TaskKey] = true
					}
				}
				for key := range taskKeys {
					if !boundKeys[key] {
						errs = append(errs, fmt.Sprintf("state %q: task %q not bound to any transition", id, key))
					}
				}
			}
		}
	}

	// cycle + reachability checks (only if no structural errors so far)
	if len(errs) == 0 && len(stateIndex) > 0 {
		if cycleErr := detectCycle(d, stateIndex); cycleErr != nil {
			errs = append(errs, cycleErr.Error())
		}
		if d.StartState != "" {
			unreachable := findUnreachable(d, stateIndex)
			for _, id := range unreachable {
				errs = append(errs, fmt.Sprintf("state %q is unreachable from start_state", id))
			}
		}
		if d.EndState != "" {
			noPath := findNoPathToEnd(d, stateIndex)
			for _, id := range noPath {
				errs = append(errs, fmt.Sprintf("state %q has no path to end_state", id))
			}
		}
	}

	if len(errs) > 0 {
		return warnings, errors.New(joinErrors(errs))
	}
	return warnings, nil
}

func (d *WorkflowDefinition) TransitionMode(stateID string) TransitionMode {
	outs := d.OutgoingTransitions(stateID)
	if len(outs) <= 1 {
		return ModeLinear
	}
	for _, tr := range outs {
		if tr.Trigger != nil {
			return ModeTaskBound
		}
	}
	return ModeUserPicks
}

func (d *WorkflowDefinition) OutgoingTransitions(stateID string) []TransitionDef {
	var out []TransitionDef
	for _, tr := range d.Transitions {
		if tr.FromState == stateID {
			out = append(out, tr)
		}
	}
	return out
}

func (d *WorkflowDefinition) FindState(stateID string) (*StateDef, bool) {
	for i := range d.States {
		if d.States[i].ID == stateID {
			return &d.States[i], true
		}
	}
	return nil, false
}

func (d *WorkflowDefinition) MarshalJSON() ([]byte, error) {
	type Alias WorkflowDefinition
	return json.Marshal((*Alias)(d))
}

func detectCycle(d *WorkflowDefinition, stateIndex map[string]*StateDef) error {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var dfs func(id string) bool
	dfs = func(id string) bool {
		visited[id] = true
		inStack[id] = true
		for _, tr := range d.OutgoingTransitions(id) {
			if !visited[tr.ToState] {
				if dfs(tr.ToState) {
					return true
				}
			} else if inStack[tr.ToState] {
				return true
			}
		}
		inStack[id] = false
		return false
	}

	for id := range stateIndex {
		if !visited[id] {
			if dfs(id) {
				return errors.New("workflow definition contains a cycle")
			}
		}
	}
	return nil
}

func findUnreachable(d *WorkflowDefinition, stateIndex map[string]*StateDef) []string {
	reachable := make(map[string]bool)
	queue := []string{d.StartState}
	reachable[d.StartState] = true
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, tr := range d.OutgoingTransitions(cur) {
			if !reachable[tr.ToState] {
				reachable[tr.ToState] = true
				queue = append(queue, tr.ToState)
			}
		}
	}
	var unreachable []string
	for id := range stateIndex {
		if !reachable[id] {
			unreachable = append(unreachable, id)
		}
	}
	return unreachable
}

func findNoPathToEnd(d *WorkflowDefinition, stateIndex map[string]*StateDef) []string {
	// collect all terminal states (zero tasks + zero outgoing) as valid end nodes
	outgoing := make(map[string]int)
	incoming := make(map[string][]string)
	for _, tr := range d.Transitions {
		outgoing[tr.FromState]++
		incoming[tr.ToState] = append(incoming[tr.ToState], tr.FromState)
	}

	terminals := []string{}
	for id, s := range stateIndex {
		if outgoing[id] == 0 && len(s.Tasks) == 0 {
			terminals = append(terminals, id)
		}
	}

	// reverse BFS from all terminal states
	canReachTerminal := make(map[string]bool)
	queue := terminals
	for _, t := range terminals {
		canReachTerminal[t] = true
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, from := range incoming[cur] {
			if !canReachTerminal[from] {
				canReachTerminal[from] = true
				queue = append(queue, from)
			}
		}
	}

	var noPath []string
	for id := range stateIndex {
		if canReachTerminal[id] {
			continue
		}
		noPath = append(noPath, id)
	}
	return noPath
}

func joinErrors(errs []string) string {
	if len(errs) == 1 {
		return errs[0]
	}
	result := fmt.Sprintf("%d validation errors:", len(errs))
	for i, e := range errs {
		result += fmt.Sprintf("\n  %d. %s", i+1, e)
	}
	return result
}
