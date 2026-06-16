package workflowloader_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
	"github.com/workflow-engine/backend/internal/testutil"
	"github.com/workflow-engine/backend/internal/workflowloader"
)

var silentLog = func() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l
}()

func writeJSON(t *testing.T, dir string, name string, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, data, 0644))
	return path
}

func validDef(name string, version int) model.WorkflowDefinition {
	return model.WorkflowDefinition{
		Name: name, Version: version, Description: "test",
		StartState: "s1", EndState: "s2",
		States: []model.StateDef{
			{ID: "s1", Label: "S1", Tasks: []model.TaskDef{{Key: "t1", Label: "T1", Type: "manual"}}},
			{ID: "s2", Label: "S2", Tasks: []model.TaskDef{}},
		},
		Transitions: []model.TransitionDef{{FromState: "s1", ToState: "s2"}},
	}
}

func TestLoader_ValidFile_Loads(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, dir, "flow.json", validDef("test-flow", 1))

	created := []*model.WorkflowDef{}
	repo := &testutil.MockWorkflowDefRepo{
		GetByNameVersionFn: func(_ context.Context, _ string, _ int) (*model.WorkflowDef, error) {
			return nil, repository.ErrNotFound
		},
		CreateFn: func(_ context.Context, def *model.WorkflowDef) error {
			created = append(created, def)
			return nil
		},
	}

	result, err := workflowloader.New(dir, repo, silentLog).Load(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.Loaded, 1)
	assert.Len(t, result.Errors, 0)
	assert.Len(t, created, 1)
}

func TestLoader_InvalidJSON_Error(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{invalid json"), 0644))

	repo := &testutil.MockWorkflowDefRepo{
		GetByNameVersionFn: func(_ context.Context, _ string, _ int) (*model.WorkflowDef, error) {
			return nil, repository.ErrNotFound
		},
	}

	result, err := workflowloader.New(dir, repo, silentLog).Load(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "invalid JSON")
}

func TestLoader_UnknownTaskType_Error(t *testing.T) {
	dir := t.TempDir()
	def := validDef("bad-type", 1)
	def.States[0].Tasks[0].Type = "unknown_type"
	writeJSON(t, dir, "flow.json", def)

	repo := &testutil.MockWorkflowDefRepo{
		GetByNameVersionFn: func(_ context.Context, _ string, _ int) (*model.WorkflowDef, error) {
			return nil, repository.ErrNotFound
		},
	}

	result, err := workflowloader.New(dir, repo, silentLog).Load(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "validation")
}

func TestLoader_SameContentNoOp(t *testing.T) {
	dir := t.TempDir()
	def := validDef("test-flow", 1)
	writeJSON(t, dir, "flow.json", def)

	existing := &model.WorkflowDef{ID: uuid.New(), Name: "test-flow", Version: 1, Definition: def}
	createCalled := false
	repo := &testutil.MockWorkflowDefRepo{
		GetByNameVersionFn: func(_ context.Context, _ string, _ int) (*model.WorkflowDef, error) {
			return existing, nil
		},
		CreateFn: func(_ context.Context, _ *model.WorkflowDef) error {
			createCalled = true
			return nil
		},
	}

	result, err := workflowloader.New(dir, repo, silentLog).Load(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.Loaded, 1)
	assert.Len(t, result.Errors, 0)
	assert.False(t, createCalled, "Create should not be called for identical content")
}

func TestLoader_DifferentContent_VersionConflict(t *testing.T) {
	dir := t.TempDir()
	def := validDef("test-flow", 1)
	writeJSON(t, dir, "flow.json", def)

	differentDef := validDef("test-flow", 1)
	differentDef.Description = "different description"
	existing := &model.WorkflowDef{ID: uuid.New(), Name: "test-flow", Version: 1, Definition: differentDef}
	repo := &testutil.MockWorkflowDefRepo{
		GetByNameVersionFn: func(_ context.Context, _ string, _ int) (*model.WorkflowDef, error) {
			return existing, nil
		},
	}

	result, err := workflowloader.New(dir, repo, silentLog).Load(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "version conflict")
}

func TestLoader_NonexistentDir_ReturnsEmpty(t *testing.T) {
	repo := &testutil.MockWorkflowDefRepo{}
	result, err := workflowloader.New("/nonexistent/path", repo, silentLog).Load(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.Loaded, 0)
	assert.Len(t, result.Errors, 0)
}

func TestLoader_SkipsNonJSONFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0644))
	writeJSON(t, dir, "flow.json", validDef("ok-flow", 1))

	created := 0
	repo := &testutil.MockWorkflowDefRepo{
		GetByNameVersionFn: func(_ context.Context, _ string, _ int) (*model.WorkflowDef, error) {
			return nil, repository.ErrNotFound
		},
		CreateFn: func(_ context.Context, _ *model.WorkflowDef) error {
			created++
			return nil
		},
	}

	result, err := workflowloader.New(dir, repo, silentLog).Load(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.Loaded, 1)
	assert.Equal(t, 1, created)
}

func TestLoader_MultipleFiles_BadOneSkipped(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, dir, "good.json", validDef("good-flow", 1))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{invalid"), 0644))

	repo := &testutil.MockWorkflowDefRepo{
		GetByNameVersionFn: func(_ context.Context, _ string, _ int) (*model.WorkflowDef, error) {
			return nil, repository.ErrNotFound
		},
		CreateFn: func(_ context.Context, _ *model.WorkflowDef) error { return nil },
	}

	result, err := workflowloader.New(dir, repo, silentLog).Load(context.Background())
	require.NoError(t, err)
	assert.Len(t, result.Loaded, 1)
	assert.Len(t, result.Errors, 1)
}
