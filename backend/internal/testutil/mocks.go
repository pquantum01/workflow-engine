package testutil

import (
	"context"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/engine"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

// MockProcessRepo implements repository.ProcessRepository for testing
type MockProcessRepo struct {
	Processes map[uuid.UUID]*model.Process
}

func NewMockProcessRepo() *MockProcessRepo {
	return &MockProcessRepo{Processes: make(map[uuid.UUID]*model.Process)}
}

func (m *MockProcessRepo) Create(_ context.Context, p *model.Process) error {
	m.Processes[p.ID] = p
	return nil
}

func (m *MockProcessRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Process, error) {
	p, ok := m.Processes[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return p, nil
}

func (m *MockProcessRepo) List(_ context.Context, offset, limit int) ([]*model.Process, error) {
	var result []*model.Process
	i := 0
	for _, p := range m.Processes {
		if i >= offset && len(result) < limit {
			result = append(result, p)
		}
		i++
	}
	return result, nil
}

func (m *MockProcessRepo) Update(_ context.Context, p *model.Process) error {
	if _, ok := m.Processes[p.ID]; !ok {
		return repository.ErrNotFound
	}
	m.Processes[p.ID] = p
	return nil
}

func (m *MockProcessRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.Processes[id]; !ok {
		return repository.ErrNotFound
	}
	delete(m.Processes, id)
	return nil
}

// MockWorkflowRepo implements repository.WorkflowRepository for testing
type MockWorkflowRepo struct {
	Workflows map[uuid.UUID]*model.Workflow
}

func NewMockWorkflowRepo() *MockWorkflowRepo {
	return &MockWorkflowRepo{Workflows: make(map[uuid.UUID]*model.Workflow)}
}

func (m *MockWorkflowRepo) Create(_ context.Context, w *model.Workflow) error {
	m.Workflows[w.ID] = w
	return nil
}

func (m *MockWorkflowRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Workflow, error) {
	w, ok := m.Workflows[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return w, nil
}

func (m *MockWorkflowRepo) ListByProcessID(_ context.Context, processID uuid.UUID) ([]*model.Workflow, error) {
	var result []*model.Workflow
	for _, w := range m.Workflows {
		if w.ProcessID == processID {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *MockWorkflowRepo) Update(_ context.Context, w *model.Workflow) error {
	if _, ok := m.Workflows[w.ID]; !ok {
		return repository.ErrNotFound
	}
	m.Workflows[w.ID] = w
	return nil
}

func (m *MockWorkflowRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.Workflows[id]; !ok {
		return repository.ErrNotFound
	}
	delete(m.Workflows, id)
	return nil
}

// MockTaskRepo implements repository.TaskRepository for testing
type MockTaskRepo struct {
	Tasks map[uuid.UUID]*model.Task
}

func NewMockTaskRepo() *MockTaskRepo {
	return &MockTaskRepo{Tasks: make(map[uuid.UUID]*model.Task)}
}

func (m *MockTaskRepo) Create(_ context.Context, t *model.Task) error {
	m.Tasks[t.ID] = t
	return nil
}

func (m *MockTaskRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Task, error) {
	t, ok := m.Tasks[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return t, nil
}

func (m *MockTaskRepo) ListByWorkflowID(_ context.Context, wfID uuid.UUID) ([]*model.Task, error) {
	var result []*model.Task
	for _, t := range m.Tasks {
		if t.WorkflowID == wfID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MockTaskRepo) Update(_ context.Context, t *model.Task) error {
	if _, ok := m.Tasks[t.ID]; !ok {
		return repository.ErrNotFound
	}
	m.Tasks[t.ID] = t
	return nil
}

func (m *MockTaskRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.Tasks[id]; !ok {
		return repository.ErrNotFound
	}
	delete(m.Tasks, id)
	return nil
}

// MockProcessInstanceRepo implements repository.ProcessInstanceRepository for testing
type MockProcessInstanceRepo struct {
	Instances map[uuid.UUID]*model.ProcessInstance
}

func NewMockProcessInstanceRepo() *MockProcessInstanceRepo {
	return &MockProcessInstanceRepo{Instances: make(map[uuid.UUID]*model.ProcessInstance)}
}

func (m *MockProcessInstanceRepo) Create(_ context.Context, pi *model.ProcessInstance) error {
	m.Instances[pi.ID] = pi
	return nil
}

func (m *MockProcessInstanceRepo) GetByID(_ context.Context, id uuid.UUID) (*model.ProcessInstance, error) {
	pi, ok := m.Instances[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return pi, nil
}

func (m *MockProcessInstanceRepo) List(_ context.Context, offset, limit int) ([]*model.ProcessInstance, error) {
	var result []*model.ProcessInstance
	i := 0
	for _, pi := range m.Instances {
		if i >= offset && len(result) < limit {
			result = append(result, pi)
		}
		i++
	}
	return result, nil
}

func (m *MockProcessInstanceRepo) ListByParentID(_ context.Context, parentID uuid.UUID) ([]*model.ProcessInstance, error) {
	var result []*model.ProcessInstance
	for _, pi := range m.Instances {
		if pi.ParentID != nil && *pi.ParentID == parentID {
			result = append(result, pi)
		}
	}
	return result, nil
}

func (m *MockProcessInstanceRepo) Update(_ context.Context, pi *model.ProcessInstance) error {
	if _, ok := m.Instances[pi.ID]; !ok {
		return repository.ErrNotFound
	}
	m.Instances[pi.ID] = pi
	return nil
}

// MockWorkflowInstanceRepo implements repository.WorkflowInstanceRepository for testing
type MockWorkflowInstanceRepo struct {
	Instances map[uuid.UUID]*model.WorkflowInstance
}

func NewMockWorkflowInstanceRepo() *MockWorkflowInstanceRepo {
	return &MockWorkflowInstanceRepo{Instances: make(map[uuid.UUID]*model.WorkflowInstance)}
}

func (m *MockWorkflowInstanceRepo) Create(_ context.Context, wi *model.WorkflowInstance) error {
	m.Instances[wi.ID] = wi
	return nil
}

func (m *MockWorkflowInstanceRepo) GetByID(_ context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	wi, ok := m.Instances[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return wi, nil
}

func (m *MockWorkflowInstanceRepo) ListByProcessInstanceID(_ context.Context, piID uuid.UUID) ([]*model.WorkflowInstance, error) {
	var result []*model.WorkflowInstance
	for _, wi := range m.Instances {
		if wi.ProcessInstanceID == piID {
			result = append(result, wi)
		}
	}
	return result, nil
}

func (m *MockWorkflowInstanceRepo) Update(_ context.Context, wi *model.WorkflowInstance) error {
	if _, ok := m.Instances[wi.ID]; !ok {
		return repository.ErrNotFound
	}
	m.Instances[wi.ID] = wi
	return nil
}

// MockTaskInstanceRepo implements repository.TaskInstanceRepository for testing
type MockTaskInstanceRepo struct {
	Instances map[uuid.UUID]*model.TaskInstance
}

func NewMockTaskInstanceRepo() *MockTaskInstanceRepo {
	return &MockTaskInstanceRepo{Instances: make(map[uuid.UUID]*model.TaskInstance)}
}

func (m *MockTaskInstanceRepo) Create(_ context.Context, ti *model.TaskInstance) error {
	m.Instances[ti.ID] = ti
	return nil
}

func (m *MockTaskInstanceRepo) GetByID(_ context.Context, id uuid.UUID) (*model.TaskInstance, error) {
	ti, ok := m.Instances[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return ti, nil
}

func (m *MockTaskInstanceRepo) ListByWorkflowInstanceID(_ context.Context, wiID uuid.UUID) ([]*model.TaskInstance, error) {
	var result []*model.TaskInstance
	for _, ti := range m.Instances {
		if ti.WorkflowInstanceID == wiID {
			result = append(result, ti)
		}
	}
	return result, nil
}

func (m *MockTaskInstanceRepo) Update(_ context.Context, ti *model.TaskInstance) error {
	if _, ok := m.Instances[ti.ID]; !ok {
		return repository.ErrNotFound
	}
	m.Instances[ti.ID] = ti
	return nil
}

// MockStateTransitionRepo implements repository.StateTransitionRepository for testing
type MockStateTransitionRepo struct {
	Transitions []*model.StateTransition
}

func NewMockStateTransitionRepo() *MockStateTransitionRepo {
	return &MockStateTransitionRepo{}
}

func (m *MockStateTransitionRepo) Create(_ context.Context, st *model.StateTransition) error {
	m.Transitions = append(m.Transitions, st)
	return nil
}

func (m *MockStateTransitionRepo) ListByWorkflowInstanceID(_ context.Context, wiID uuid.UUID) ([]*model.StateTransition, error) {
	var result []*model.StateTransition
	for _, st := range m.Transitions {
		if st.WorkflowInstanceID == wiID {
			result = append(result, st)
		}
	}
	return result, nil
}

// MockEngine implements engine.Engine for testing
type MockEngine struct {
	EvaluateCalls int
}

func (m *MockEngine) Evaluate(_ context.Context, _ uuid.UUID) (*engine.TransitionResult, error) {
	m.EvaluateCalls++
	return nil, nil
}

func (m *MockEngine) TriggerTransition(_ context.Context, _ uuid.UUID, _, _, _ string) (*engine.TransitionResult, error) {
	return nil, nil
}
