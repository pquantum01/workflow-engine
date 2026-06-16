package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/workflow-engine/backend/internal/api"
	"github.com/workflow-engine/backend/internal/engine"
	"github.com/workflow-engine/backend/internal/repository"
	"github.com/workflow-engine/backend/internal/service"
	"github.com/workflow-engine/backend/internal/workflowloader"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	log := logrus.New()
	log.SetOutput(io.Discard)

	defRepo := repository.NewWorkflowDefRepo(testDB)
	instRepo := repository.NewWorkflowInstanceRepo(testDB)
	taskRepo := repository.NewTaskInstanceRepo(testDB)
	transRepo := repository.NewStateTransitionRepo(testDB)

	loader := workflowloader.New("../../workflows", defRepo, log)
	eng := engine.New(testDB, defRepo, instRepo, taskRepo, transRepo, log)
	defSvc := service.NewWorkflowDefService(defRepo, instRepo, loader, log)
	instSvc := service.NewInstanceService(eng, instRepo, taskRepo, transRepo, log)
	taskSvc := service.NewTaskService(eng, taskRepo, log)

	handler := api.NewRouter(testDB, defSvc, instSvc, taskSvc, log, "*")
	return httptest.NewServer(handler)
}

func doRequest(t *testing.T, srv *httptest.Server, method, path string, body any) *http.Response {
	t.Helper()
	var b *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		b = bytes.NewBuffer(data)
	} else {
		b = bytes.NewBuffer(nil)
	}
	req, err := http.NewRequest(method, srv.URL+path, b)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(v))
}

func seedLinearDef(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	defRepo := repository.NewWorkflowDefRepo(testDB)
	def := linearDefForEngine()
	require.NoError(t, defRepo.Create(ctx, def))
	return def.ID.String()
}

func TestAPI_Healthz(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/healthz", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAPI_Readyz(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/readyz", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAPI_ListWorkflowDefs_Empty(t *testing.T) {
	truncateTables(t)
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/v1/workflow-defs", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	decodeJSON(t, resp, &result)
	assert.Equal(t, float64(0), result["total"])
}

func TestAPI_ListWorkflowDefs_WithDefs(t *testing.T) {
	truncateTables(t)
	seedLinearDef(t)
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/v1/workflow-defs", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	decodeJSON(t, resp, &result)
	assert.Equal(t, float64(1), result["total"])
}

func TestAPI_GetWorkflowDef_NotFound(t *testing.T) {
	truncateTables(t)
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/v1/workflow-defs/00000000-0000-0000-0000-000000000000", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAPI_GetWorkflowDef_Found(t *testing.T) {
	truncateTables(t)
	defID := seedLinearDef(t)
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/v1/workflow-defs/"+defID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	decodeJSON(t, resp, &result)
	assert.Equal(t, defID, result["id"])
	assert.NotNil(t, result["definition"])
}

func TestAPI_StartInstance(t *testing.T) {
	truncateTables(t)
	defID := seedLinearDef(t)
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/v1/workflow-instances", map[string]any{
		"workflow_def_id": defID,
		"created_by":      "test-user",
	})
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]any
	decodeJSON(t, resp, &result)
	assert.Equal(t, "active", result["status"])
	assert.Equal(t, "s1", result["current_state_id"])
}

func TestAPI_StartInstance_MissingFields(t *testing.T) {
	truncateTables(t)
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/v1/workflow-instances", map[string]any{})
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestAPI_GetInstance_WithTasksAndHistory(t *testing.T) {
	truncateTables(t)
	defID := seedLinearDef(t)
	srv := newTestServer(t)
	defer srv.Close()

	// start instance
	startResp := doRequest(t, srv, http.MethodPost, "/api/v1/workflow-instances", map[string]any{
		"workflow_def_id": defID,
		"created_by":      "test-user",
	})
	var startResult map[string]any
	decodeJSON(t, startResp, &startResult)
	instID := startResult["id"].(string)

	resp := doRequest(t, srv, http.MethodGet, "/api/v1/workflow-instances/"+instID, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	decodeJSON(t, resp, &result)
	assert.NotNil(t, result["tasks"])
	assert.NotNil(t, result["history"])
}

func TestAPI_CompleteTask_Linear_Completes(t *testing.T) {
	truncateTables(t)
	defID := seedLinearDef(t)
	srv := newTestServer(t)
	defer srv.Close()

	// Start
	startResp := doRequest(t, srv, http.MethodPost, "/api/v1/workflow-instances", map[string]any{
		"workflow_def_id": defID,
		"created_by":      "test-user",
	})
	var startResult map[string]any
	decodeJSON(t, startResp, &startResult)
	instID := startResult["id"].(string)

	// List tasks
	tasksResp := doRequest(t, srv, http.MethodGet, "/api/v1/workflow-instances/"+instID+"/tasks", nil)
	var tasks []map[string]any
	decodeJSON(t, tasksResp, &tasks)
	require.Len(t, tasks, 1)
	taskID := tasks[0]["id"].(string)

	// Complete
	resp := doRequest(t, srv, http.MethodPost, "/api/v1/task-instances/"+taskID+"/complete", map[string]any{
		"status":       "done",
		"completed_by": "test-user",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	decodeJSON(t, resp, &result)
	assert.Equal(t, "completed", result["status"])
}

func TestAPI_CompleteTask_Reject_CancelsInstance(t *testing.T) {
	truncateTables(t)
	defID := seedLinearDef(t)
	srv := newTestServer(t)
	defer srv.Close()

	startResp := doRequest(t, srv, http.MethodPost, "/api/v1/workflow-instances", map[string]any{
		"workflow_def_id": defID, "created_by": "test-user",
	})
	var startResult map[string]any
	decodeJSON(t, startResp, &startResult)
	instID := startResult["id"].(string)

	tasksResp := doRequest(t, srv, http.MethodGet, "/api/v1/workflow-instances/"+instID+"/tasks", nil)
	var tasks []map[string]any
	decodeJSON(t, tasksResp, &tasks)
	taskID := tasks[0]["id"].(string)

	resp := doRequest(t, srv, http.MethodPost, "/api/v1/task-instances/"+taskID+"/complete", map[string]any{
		"status": "rejected", "completed_by": "test-user",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	decodeJSON(t, resp, &result)
	assert.Equal(t, "cancelled", result["status"])
}

func TestAPI_CompleteTask_OnTerminalInstance_Returns409(t *testing.T) {
	truncateTables(t)
	defID := seedLinearDef(t)
	srv := newTestServer(t)
	defer srv.Close()

	startResp := doRequest(t, srv, http.MethodPost, "/api/v1/workflow-instances", map[string]any{
		"workflow_def_id": defID, "created_by": "test-user",
	})
	var startResult map[string]any
	decodeJSON(t, startResp, &startResult)
	instID := startResult["id"].(string)

	// Cancel instance
	doRequest(t, srv, http.MethodPost, "/api/v1/workflow-instances/"+instID+"/cancel", nil)

	// Try to complete task
	tasksResp := doRequest(t, srv, http.MethodGet, "/api/v1/workflow-instances/"+instID+"/tasks", nil)
	var tasks []map[string]any
	decodeJSON(t, tasksResp, &tasks)
	require.NotEmpty(t, tasks)
	taskID := tasks[0]["id"].(string)

	resp := doRequest(t, srv, http.MethodPost, "/api/v1/task-instances/"+taskID+"/complete", map[string]any{
		"status": "done",
	})
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestAPI_CancelInstance(t *testing.T) {
	truncateTables(t)
	defID := seedLinearDef(t)
	srv := newTestServer(t)
	defer srv.Close()

	startResp := doRequest(t, srv, http.MethodPost, "/api/v1/workflow-instances", map[string]any{
		"workflow_def_id": defID, "created_by": "test-user",
	})
	var startResult map[string]any
	decodeJSON(t, startResp, &startResult)
	instID := startResult["id"].(string)

	resp := doRequest(t, srv, http.MethodPost, "/api/v1/workflow-instances/"+instID+"/cancel", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	decodeJSON(t, resp, &result)
	assert.Equal(t, "cancelled", result["status"])
}

func TestAPI_ListInstances_FilterByStatus(t *testing.T) {
	truncateTables(t)
	defID := seedLinearDef(t)
	srv := newTestServer(t)
	defer srv.Close()

	doRequest(t, srv, http.MethodPost, "/api/v1/workflow-instances", map[string]any{
		"workflow_def_id": defID, "created_by": "user1",
	})

	resp := doRequest(t, srv, http.MethodGet, "/api/v1/workflow-instances?status=active", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	decodeJSON(t, resp, &result)
	assert.Equal(t, float64(1), result["total"])
}

func TestAPI_TaskTypes(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/v1/task-types", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var types []map[string]any
	decodeJSON(t, resp, &types)
	assert.Len(t, types, 4)
}

func TestAPI_ReloadDefs(t *testing.T) {
	truncateTables(t)
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodPost, "/api/v1/admin/reload-defs", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	decodeJSON(t, resp, &result)
	// loaded, warnings, errors always present as arrays (may be empty)
	_, hasLoaded := result["loaded"]
	_, hasErrors := result["errors"]
	assert.True(t, hasLoaded)
	assert.True(t, hasErrors)
}

func TestAPI_GetTask_NotFound(t *testing.T) {
	truncateTables(t)
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, srv, http.MethodGet, "/api/v1/task-instances/00000000-0000-0000-0000-000000000000", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
