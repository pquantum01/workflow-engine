package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
)

type taskAttachmentRepo struct {
	db DB
}

func NewTaskAttachmentRepo(db DB) TaskAttachmentRepo {
	return &taskAttachmentRepo{db: db}
}

func (r *taskAttachmentRepo) WithTx(tx DB) TaskAttachmentRepo {
	return &taskAttachmentRepo{db: tx}
}

func (r *taskAttachmentRepo) Upsert(ctx context.Context, att *model.TaskAttachment) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO task_attachment
		  (id, task_instance_id, filename, content_type, size_bytes, s3_url, uploaded_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (task_instance_id) DO UPDATE SET
		  id           = EXCLUDED.id,
		  filename     = EXCLUDED.filename,
		  content_type = EXCLUDED.content_type,
		  size_bytes   = EXCLUDED.size_bytes,
		  s3_url       = EXCLUDED.s3_url,
		  updated_at   = EXCLUDED.updated_at`,
		att.ID, att.TaskInstanceID, att.Filename, att.ContentType,
		att.SizeBytes, att.S3URL, att.UploadedAt, att.UpdatedAt,
	)
	return err
}

func (r *taskAttachmentRepo) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*model.TaskAttachment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, task_instance_id, filename, content_type, size_bytes, s3_url, uploaded_at, updated_at
		 FROM task_attachment WHERE task_instance_id = $1`,
		taskID,
	)
	var a model.TaskAttachment
	err := row.Scan(&a.ID, &a.TaskInstanceID, &a.Filename, &a.ContentType,
		&a.SizeBytes, &a.S3URL, &a.UploadedAt, &a.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *taskAttachmentRepo) DeleteByTaskID(ctx context.Context, taskID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM task_attachment WHERE task_instance_id = $1`, taskID)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}
