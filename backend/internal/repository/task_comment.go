package repository

import (
	"context"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
)

type taskCommentRepo struct {
	db DB
}

func NewTaskCommentRepo(db DB) TaskCommentRepo {
	return &taskCommentRepo{db: db}
}

func (r *taskCommentRepo) WithTx(tx DB) TaskCommentRepo {
	return &taskCommentRepo{db: tx}
}

func (r *taskCommentRepo) Add(ctx context.Context, comment *model.TaskComment) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO task_comment (id, task_instance_id, author, content, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		comment.ID, comment.TaskInstanceID, comment.Author, comment.Content, comment.CreatedAt,
	)
	return err
}

func (r *taskCommentRepo) ListByTask(ctx context.Context, taskID uuid.UUID) ([]*model.TaskComment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, task_instance_id, author, content, created_at
		 FROM task_comment WHERE task_instance_id = $1 ORDER BY created_at`,
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.TaskComment
	for rows.Next() {
		var c model.TaskComment
		if err := rows.Scan(&c.ID, &c.TaskInstanceID, &c.Author, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}
	return comments, rows.Err()
}
