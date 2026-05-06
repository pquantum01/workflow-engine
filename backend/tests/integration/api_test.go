//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"workflow-engine/backend/internal/api"
	"workflow-engine/backend/internal/engine"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository/postgres"
	"workflow-engine/backend/internal/service"
)

// setupTestServer creates a test HTTP server backed by real PostgreSQL.
func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	cleanupTables(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	processRepo := postgres.NewProcessRepo(testDB)
	workflowRepo := postgres.NewWorkflowRepo(testDB)
	taskRepo := postgres.NewTaskRepo(testDB)
	processInstanceRepo := postgres.NewProcessInstanceRepo(testDB)
	workflowInstanceRepo := postgres.NewWorkflowInstanceRepo(testDB)
	taskInstanceRepo := postgres.NewTaskInstanceRepo(testDB)
	stateTransitionRepo := postgres.NewStateTransitionRepo(testDB)

	eng := engine.New(workflowRepo, workflowInstanceRepo, taskRepo, taskInstanceRepo, stateTransitionRepo, logger)

	processSvc := service.NewProcessService(processRepo, logger)
	workflowSvc := service.NewWorkflowService(workflowRepo, processRepo, logger)
	taskSvc := service.NewTaskService(taskRepo, workflowRepo, processRepo, logger)
	processInstanceSvc := service.NewProcessInstanceService(processRepo, workflowRepo, taskRepo, processInstanceRepo, workflowInstanceRepo, taskInstanceRepo, logger)
	workflowInstanceSvc := service.NewWorkflowInstanceService(workflowInstanceRepo, stateTransitionRepo, eng, logger)
	taskInstanceSvc := service.NewTaskInstanceService(taskInstanceRepo, workflowInstanceRepo, eng, logger)

	router := api.NewRouter(processSvc, workflowSvc, taskSvc, processInstanceSvc, workflowInstanceSvc, taskInstanceSvc, logger)
	return httptest.NewServer(router)
}

func postJSON(url string, body any) (*http.Response, error) {
	b, _ := json.Marshal(body)
	return http.Post(url, "application/json", bytes.NewReader(b))
}

func putJSON(url string, body any) (*http.Response, error) {
	b, _ := json.Marshal(body)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func doDelete(url string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

// --- Process CRUD Tests ---

func TestAPI_ProcessCRUD(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create
	resp, err := postJSON(ts.URL+"/api/v1/processes", map[string]string{
		"name": "Test Process", "description": "desc",
	})
	if err != nil {
		t.Fatalf("Create request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var created model.Process
	decodeJSON(t, resp, &created)
	if created.Name != "Test Process" {
		t.Errorf("Name = %q, want %q", created.Name, "Test Process")
	}
	if created.Status != "draft" {
		t.Errorf("Status = %q, want %q", created.Status, "draft")
	}

	// GetByID
	resp, err = http.Get(ts.URL + "/api/v1/processes/" + created.ID.String())
	if err != nil {
		t.Fatalf("Get request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var fetched model.Process
	decodeJSON(t, resp, &fetched)
	if fetched.ID != created.ID {
		t.Errorf("ID mismatch")
	}

	// List
	resp, err = http.Get(ts.URL + "/api/v1/processes")
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var list []model.Process
	decodeJSON(t, resp, &list)
	if len(list) != 1 {
		t.Errorf("List length = %d, want 1", len(list))
	}

	// Update
	resp, err = putJSON(ts.URL+"/api/v1/processes/"+created.ID.String(), map[string]string{
		"name": "Updated",
	})
	if err != nil {
		t.Fatalf("Update request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Update status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var updated model.Process
	decodeJSON(t, resp, &updated)
	if updated.Name != "Updated" {
		t.Errorf("Name = %q, want %q", updated.Name, "Updated")
	}

	// Delete
	resp, err = doDelete(ts.URL + "/api/v1/processes/" + created.ID.String())
	if err != nil {
		t.Fatalf("Delete request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Delete status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	resp.Body.Close()

	// Verify deleted
	resp, err = http.Get(ts.URL + "/api/v1/processes/" + created.ID.String())
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Get after delete status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	resp.Body.Close()
}

func TestAPI_Process_CreateEmptyName(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, err := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": ""})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	resp.Body.Close()
}

func TestAPI_Process_DeleteActiveBlocked(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create and activate
	resp, _ := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	putJSON(ts.URL+"/api/v1/processes/"+p.ID.String(), map[string]string{"status": "active"})

	resp, _ = doDelete(ts.URL + "/api/v1/processes/" + p.ID.String())
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("Delete active status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
	resp.Body.Close()
}

// --- Workflow CRUD Tests ---

func TestAPI_WorkflowCRUD(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create process first
	resp, _ := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}

	// Create workflow
	resp, err := postJSON(ts.URL+"/api/v1/processes/"+p.ID.String()+"/workflows", map[string]any{
		"name": "WF", "state_machine": sm,
	})
	if err != nil {
		t.Fatalf("Create workflow failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var wf model.Workflow
	decodeJSON(t, resp, &wf)

	// List by process
	resp, _ = http.Get(ts.URL + "/api/v1/processes/" + p.ID.String() + "/workflows")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var wfList []model.Workflow
	decodeJSON(t, resp, &wfList)
	if len(wfList) != 1 {
		t.Errorf("List length = %d, want 1", len(wfList))
	}

	// Get by ID
	resp, _ = http.Get(ts.URL + "/api/v1/workflows/" + wf.ID.String())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// Update
	resp, _ = putJSON(ts.URL+"/api/v1/workflows/"+wf.ID.String(), map[string]string{"name": "Updated WF"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Update status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// Delete
	resp, _ = doDelete(ts.URL + "/api/v1/workflows/" + wf.ID.String())
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Delete status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	resp.Body.Close()
}

// --- Task CRUD Tests ---

func TestAPI_TaskCRUD(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Setup: process + workflow
	resp, _ := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	resp, _ = postJSON(ts.URL+"/api/v1/processes/"+p.ID.String()+"/workflows", map[string]any{
		"name": "WF", "state_machine": sm,
	})
	var wf model.Workflow
	decodeJSON(t, resp, &wf)

	// Create task
	resp, err := postJSON(ts.URL+"/api/v1/workflows/"+wf.ID.String()+"/tasks", map[string]string{
		"name": "Review", "task_type": "manual",
	})
	if err != nil {
		t.Fatalf("Create task failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Create status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var task model.Task
	decodeJSON(t, resp, &task)

	// List by workflow
	resp, _ = http.Get(ts.URL + "/api/v1/workflows/" + wf.ID.String() + "/tasks")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var taskList []model.Task
	decodeJSON(t, resp, &taskList)
	if len(taskList) != 1 {
		t.Errorf("List length = %d, want 1", len(taskList))
	}

	// Get
	resp, _ = http.Get(ts.URL + "/api/v1/tasks/" + task.ID.String())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// Update
	resp, _ = putJSON(ts.URL+"/api/v1/tasks/"+task.ID.String(), map[string]string{"name": "Code Review"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Update status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// Delete
	resp, _ = doDelete(ts.URL + "/api/v1/tasks/" + task.ID.String())
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Delete status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	resp.Body.Close()
}

// --- Empty List Tests ---

func TestAPI_EmptyListReturnsArray(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Processes
	resp, err := http.Get(ts.URL + "/api/v1/processes")
	if err != nil {
		t.Fatalf("List processes request failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List processes status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if string(body) != "[]\n" {
		t.Errorf("List processes body = %q, want %q", string(body), "[]\n")
	}

	// Process instances
	resp, err = http.Get(ts.URL + "/api/v1/process-instances")
	if err != nil {
		t.Fatalf("List process instances request failed: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "[]\n" {
		t.Errorf("List process instances body = %q, want %q", string(body), "[]\n")
	}

	// Workflows for a process (create a process first to get a valid ID)
	resp, _ = postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	resp, err = http.Get(ts.URL + "/api/v1/processes/" + p.ID.String() + "/workflows")
	if err != nil {
		t.Fatalf("List workflows request failed: %v", err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "[]\n" {
		t.Errorf("List workflows body = %q, want %q", string(body), "[]\n")
	}
}

// --- Process Instance Lifecycle Tests ---

func TestAPI_ProcessInstanceLifecycle(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Setup: create active process with workflow and task
	resp, _ := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	resp, _ = postJSON(ts.URL+"/api/v1/processes/"+p.ID.String()+"/workflows", map[string]any{
		"name": "WF", "state_machine": sm,
	})
	var wf model.Workflow
	decodeJSON(t, resp, &wf)

	resp, _ = postJSON(ts.URL+"/api/v1/workflows/"+wf.ID.String()+"/tasks", map[string]string{
		"name": "Review", "task_type": "manual",
	})
	resp.Body.Close()

	// Activate the process
	resp, _ = putJSON(ts.URL+"/api/v1/processes/"+p.ID.String(), map[string]string{"status": "active"})
	resp.Body.Close()

	// Start process instance
	resp, err := postJSON(ts.URL+"/api/v1/process-instances", map[string]string{
		"process_id": p.ID.String(),
	})
	if err != nil {
		t.Fatalf("Start request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Start status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var pi model.ProcessInstance
	decodeJSON(t, resp, &pi)
	if pi.Status != "running" {
		t.Errorf("Status = %q, want %q", pi.Status, "running")
	}

	// Get process instance
	resp, _ = http.Get(ts.URL + "/api/v1/process-instances/" + pi.ID.String())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// List process instances
	resp, _ = http.Get(ts.URL + "/api/v1/process-instances")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var piList []model.ProcessInstance
	decodeJSON(t, resp, &piList)
	if len(piList) != 1 {
		t.Errorf("List length = %d, want 1", len(piList))
	}

	// List workflow instances for process instance
	resp, _ = http.Get(ts.URL + "/api/v1/process-instances/" + pi.ID.String() + "/workflows")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List workflows status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var wiList []model.WorkflowInstance
	decodeJSON(t, resp, &wiList)
	if len(wiList) != 1 {
		t.Errorf("Workflow instances = %d, want 1", len(wiList))
	}

	// Cancel
	resp, _ = postJSON(ts.URL+"/api/v1/process-instances/"+pi.ID.String()+"/cancel", nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Cancel status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	resp.Body.Close()

	// Verify cancelled
	resp, _ = http.Get(ts.URL + "/api/v1/process-instances/" + pi.ID.String())
	var cancelled model.ProcessInstance
	decodeJSON(t, resp, &cancelled)
	if cancelled.Status != "cancelled" {
		t.Errorf("Status = %q, want %q", cancelled.Status, "cancelled")
	}
}

func TestAPI_ProcessInstance_StartDraftBlocked(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, _ := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	resp, _ = postJSON(ts.URL+"/api/v1/process-instances", map[string]string{
		"process_id": p.ID.String(),
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("Start draft status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
	resp.Body.Close()
}

func TestAPI_ProcessInstance_ParentChild(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Setup active process
	resp, _ := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	resp, _ = postJSON(ts.URL+"/api/v1/processes/"+p.ID.String()+"/workflows", map[string]any{
		"name": "WF", "state_machine": sm,
	})
	resp.Body.Close()
	putJSON(ts.URL+"/api/v1/processes/"+p.ID.String(), map[string]string{"status": "active"})

	// Start parent
	resp, _ = postJSON(ts.URL+"/api/v1/process-instances", map[string]string{
		"process_id": p.ID.String(),
	})
	var parent model.ProcessInstance
	decodeJSON(t, resp, &parent)

	// Start child
	resp, _ = postJSON(ts.URL+"/api/v1/process-instances", map[string]any{
		"process_id": p.ID.String(),
		"parent_id":  parent.ID.String(),
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Start child status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var child model.ProcessInstance
	decodeJSON(t, resp, &child)

	// List children
	resp, _ = http.Get(ts.URL + "/api/v1/process-instances/" + parent.ID.String() + "/children")
	var children []model.ProcessInstance
	decodeJSON(t, resp, &children)
	if len(children) != 1 {
		t.Errorf("Children = %d, want 1", len(children))
	}

	// Grandchild should be blocked
	resp, _ = postJSON(ts.URL+"/api/v1/process-instances", map[string]any{
		"process_id": p.ID.String(),
		"parent_id":  child.ID.String(),
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("Grandchild status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
	resp.Body.Close()
}

// --- Workflow Instance + Task Instance Tests ---

func TestAPI_TaskInstanceStatusAndTransition(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Setup: active process with workflow (completion rule) and task
	resp, _ := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	sm := model.StateMachine{
		States:      []model.State{{Name: "open"}, {Name: "done"}},
		Transitions: []model.Transition{{Name: "finish", FromState: "open", ToState: "done", CompletionRule: &model.CompletionRule{Type: "single", TaskRefs: []string{"Review"}}}},
		InitialState:   "open",
		TerminalStates: []string{"done"},
	}
	resp, _ = postJSON(ts.URL+"/api/v1/processes/"+p.ID.String()+"/workflows", map[string]any{
		"name": "WF", "state_machine": sm,
	})
	var wf model.Workflow
	decodeJSON(t, resp, &wf)

	resp, _ = postJSON(ts.URL+"/api/v1/workflows/"+wf.ID.String()+"/tasks", map[string]string{
		"name": "Review", "task_type": "manual",
	})
	resp.Body.Close()

	putJSON(ts.URL+"/api/v1/processes/"+p.ID.String(), map[string]string{"status": "active"})

	// Start instance
	resp, _ = postJSON(ts.URL+"/api/v1/process-instances", map[string]string{"process_id": p.ID.String()})
	var pi model.ProcessInstance
	decodeJSON(t, resp, &pi)

	// Get workflow instance
	resp, _ = http.Get(ts.URL + "/api/v1/process-instances/" + pi.ID.String() + "/workflows")
	var wiList []model.WorkflowInstance
	decodeJSON(t, resp, &wiList)
	wi := wiList[0]

	if wi.CurrentState != "open" {
		t.Errorf("CurrentState = %q, want %q", wi.CurrentState, "open")
	}

	// Get task instances
	resp, _ = http.Get(ts.URL + "/api/v1/workflow-instances/" + wi.ID.String() + "/tasks")
	var tiList []model.TaskInstance
	decodeJSON(t, resp, &tiList)
	if len(tiList) != 1 {
		t.Fatalf("Task instances = %d, want 1", len(tiList))
	}
	ti := tiList[0]

	// Assign task
	resp, _ = putJSON(ts.URL+"/api/v1/task-instances/"+ti.ID.String()+"/assign", map[string]string{
		"assigned_to": "user-1",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Assign status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var assigned model.TaskInstance
	decodeJSON(t, resp, &assigned)
	if assigned.AssignedTo != "user-1" {
		t.Errorf("AssignedTo = %q, want %q", assigned.AssignedTo, "user-1")
	}

	// Move to in_progress
	resp, _ = putJSON(ts.URL+"/api/v1/task-instances/"+ti.ID.String()+"/status", map[string]string{
		"status": "in_progress",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Status update to in_progress = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// Complete the task — should auto-evaluate and transition workflow
	resp, _ = putJSON(ts.URL+"/api/v1/task-instances/"+ti.ID.String()+"/status", map[string]string{
		"status": "completed", "triggered_by": "user-1", "trigger_type": "human",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Status update to completed = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// Verify workflow transitioned to done (terminal)
	resp, _ = http.Get(ts.URL + "/api/v1/workflow-instances/" + wi.ID.String())
	var updatedWI model.WorkflowInstance
	decodeJSON(t, resp, &updatedWI)
	if updatedWI.CurrentState != "done" {
		t.Errorf("CurrentState = %q, want %q", updatedWI.CurrentState, "done")
	}
	if updatedWI.Status != "completed" {
		t.Errorf("Status = %q, want %q", updatedWI.Status, "completed")
	}
}

func TestAPI_TaskInstance_InvalidTransition(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Setup
	resp, _ := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	resp, _ = postJSON(ts.URL+"/api/v1/processes/"+p.ID.String()+"/workflows", map[string]any{
		"name": "WF", "state_machine": sm,
	})
	var wf model.Workflow
	decodeJSON(t, resp, &wf)

	resp, _ = postJSON(ts.URL+"/api/v1/workflows/"+wf.ID.String()+"/tasks", map[string]string{
		"name": "T", "task_type": "manual",
	})
	resp.Body.Close()

	putJSON(ts.URL+"/api/v1/processes/"+p.ID.String(), map[string]string{"status": "active"})

	resp, _ = postJSON(ts.URL+"/api/v1/process-instances", map[string]string{"process_id": p.ID.String()})
	var pi model.ProcessInstance
	decodeJSON(t, resp, &pi)

	resp, _ = http.Get(ts.URL + "/api/v1/process-instances/" + pi.ID.String() + "/workflows")
	var wiList []model.WorkflowInstance
	decodeJSON(t, resp, &wiList)

	resp, _ = http.Get(ts.URL + "/api/v1/workflow-instances/" + wiList[0].ID.String() + "/tasks")
	var tiList []model.TaskInstance
	decodeJSON(t, resp, &tiList)

	// Try to jump from pending to completed (invalid)
	resp, _ = putJSON(ts.URL+"/api/v1/task-instances/"+tiList[0].ID.String()+"/status", map[string]string{
		"status": "completed",
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("Invalid transition status = %d, want %d", resp.StatusCode, http.StatusConflict)
	}
	resp.Body.Close()
}

func TestAPI_WorkflowInstance_ManualTransition(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Setup: active process with unguarded transition
	resp, _ := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "close", FromState: "open", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	resp, _ = postJSON(ts.URL+"/api/v1/processes/"+p.ID.String()+"/workflows", map[string]any{
		"name": "WF", "state_machine": sm,
	})
	resp.Body.Close()

	putJSON(ts.URL+"/api/v1/processes/"+p.ID.String(), map[string]string{"status": "active"})

	resp, _ = postJSON(ts.URL+"/api/v1/process-instances", map[string]string{"process_id": p.ID.String()})
	var pi model.ProcessInstance
	decodeJSON(t, resp, &pi)

	resp, _ = http.Get(ts.URL + "/api/v1/process-instances/" + pi.ID.String() + "/workflows")
	var wiList []model.WorkflowInstance
	decodeJSON(t, resp, &wiList)
	wi := wiList[0]

	// Trigger manual transition
	resp, err := postJSON(ts.URL+"/api/v1/workflow-instances/"+wi.ID.String()+"/transition", map[string]string{
		"transition_name": "close", "triggered_by": "admin", "trigger_type": "human",
	})
	if err != nil {
		t.Fatalf("Trigger transition failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Trigger transition status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]any
	decodeJSON(t, resp, &result)
	if result["from_state"] != "open" || result["to_state"] != "closed" {
		t.Errorf("Transition = %v → %v, want open → closed", result["from_state"], result["to_state"])
	}
}

// --- State Transitions List Test ---

func TestAPI_StateTransitionsList(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Setup: active process with unguarded transition
	resp, _ := postJSON(ts.URL+"/api/v1/processes", map[string]string{"name": "P"})
	var p model.Process
	decodeJSON(t, resp, &p)

	sm := model.StateMachine{
		States:         []model.State{{Name: "open"}, {Name: "review"}, {Name: "closed"}},
		Transitions:    []model.Transition{{Name: "start_review", FromState: "open", ToState: "review"}, {Name: "close", FromState: "review", ToState: "closed"}},
		InitialState:   "open",
		TerminalStates: []string{"closed"},
	}
	resp, _ = postJSON(ts.URL+"/api/v1/processes/"+p.ID.String()+"/workflows", map[string]any{
		"name": "WF", "state_machine": sm,
	})
	resp.Body.Close()

	putJSON(ts.URL+"/api/v1/processes/"+p.ID.String(), map[string]string{"status": "active"})

	resp, _ = postJSON(ts.URL+"/api/v1/process-instances", map[string]string{"process_id": p.ID.String()})
	var pi model.ProcessInstance
	decodeJSON(t, resp, &pi)

	resp, _ = http.Get(ts.URL + "/api/v1/process-instances/" + pi.ID.String() + "/workflows")
	var wiList []model.WorkflowInstance
	decodeJSON(t, resp, &wiList)
	wi := wiList[0]

	// Initially empty
	resp, err := http.Get(ts.URL + "/api/v1/workflow-instances/" + wi.ID.String() + "/transitions")
	if err != nil {
		t.Fatalf("List transitions request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List transitions status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var emptyList []model.StateTransition
	decodeJSON(t, resp, &emptyList)
	if len(emptyList) != 0 {
		t.Errorf("Initial transitions = %d, want 0", len(emptyList))
	}

	// Trigger a manual transition
	postJSON(ts.URL+"/api/v1/workflow-instances/"+wi.ID.String()+"/transition", map[string]string{
		"transition_name": "start_review", "triggered_by": "admin", "trigger_type": "human",
	})

	// Now should have 1 transition
	resp, _ = http.Get(ts.URL + "/api/v1/workflow-instances/" + wi.ID.String() + "/transitions")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("List transitions status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var transitions []model.StateTransition
	decodeJSON(t, resp, &transitions)
	if len(transitions) != 1 {
		t.Fatalf("Transitions = %d, want 1", len(transitions))
	}
	if transitions[0].FromState != "open" || transitions[0].ToState != "review" {
		t.Errorf("Transition = %s → %s, want open → review", transitions[0].FromState, transitions[0].ToState)
	}
	if transitions[0].TriggeredBy != "admin" {
		t.Errorf("TriggeredBy = %q, want %q", transitions[0].TriggeredBy, "admin")
	}

	// Trigger second transition
	postJSON(ts.URL+"/api/v1/workflow-instances/"+wi.ID.String()+"/transition", map[string]string{
		"transition_name": "close", "triggered_by": "bot", "trigger_type": "ai_agent",
	})

	// Should have 2 transitions
	resp, _ = http.Get(ts.URL + "/api/v1/workflow-instances/" + wi.ID.String() + "/transitions")
	var allTransitions []model.StateTransition
	decodeJSON(t, resp, &allTransitions)
	if len(allTransitions) != 2 {
		t.Fatalf("Transitions = %d, want 2", len(allTransitions))
	}
}

// --- Error Response Tests ---

func TestAPI_NotFoundError(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/api/v1/processes/00000000-0000-0000-0000-000000000001")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	resp.Body.Close()
}

func TestAPI_InvalidUUID(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/api/v1/processes/not-a-uuid")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	resp.Body.Close()
}

func TestAPI_HealthCheck(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()
}

// --- CORS Test ---

func TestAPI_CORSHeaders(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	req, _ := http.NewRequest("OPTIONS", ts.URL+"/api/v1/processes", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("OPTIONS status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("CORS origin header missing")
	}
	resp.Body.Close()
}
