package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"workflow-engine/backend/internal/model"
)

func TestTaskInstance_DefaultPriority(t *testing.T) {
	task := model.TaskInstance{Priority: "medium"}
	assert.Equal(t, "medium", task.Priority)
	assert.False(t, task.IsAdhoc)
}

func TestTaskComment_Fields(t *testing.T) {
	c := model.TaskComment{Author: "alice", Content: "looks good"}
	assert.Equal(t, "alice", c.Author)
	assert.Equal(t, "looks good", c.Content)
}

func TestTaskAttachment_Fields(t *testing.T) {
	a := model.TaskAttachment{Filename: "doc.pdf", SizeBytes: 1024}
	assert.Equal(t, "doc.pdf", a.Filename)
	assert.Equal(t, 1024, a.SizeBytes)
}
