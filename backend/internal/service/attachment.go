package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
	"workflow-engine/backend/internal/storage"
)

const maxUploadBytes = 2 * 1024 * 1024 // 2MB

var ErrFeatureDisabled = errors.New("file upload is not enabled")
var ErrFileTooLarge    = errors.New("file exceeds 2MB limit")

type AttachmentService struct {
	store   storage.FileStore
	attRepo repository.TaskAttachmentRepo
	log     *logrus.Logger
}

func NewAttachmentService(store storage.FileStore, attRepo repository.TaskAttachmentRepo, log *logrus.Logger) *AttachmentService {
	return &AttachmentService{store: store, attRepo: attRepo, log: log}
}

func (s *AttachmentService) Upload(ctx context.Context, taskID uuid.UUID, filename, contentType string, data []byte) (*model.TaskAttachment, error) {
	if len(data) > maxUploadBytes {
		return nil, ErrFileTooLarge
	}

	key := fmt.Sprintf("attachments/%s/%s", taskID, filename)
	s3URL, err := s.store.Upload(ctx, key, contentType, bytes.NewReader(data))
	if err != nil {
		if errors.Is(err, storage.ErrFeatureDisabled) {
			return nil, ErrFeatureDisabled
		}
		s.log.WithField("err", err).Error("attachmentService.Upload: store error")
		return nil, err
	}

	now := time.Now()
	att := &model.TaskAttachment{
		ID:             uuid.New(),
		TaskInstanceID: taskID,
		Filename:       filename,
		ContentType:    contentType,
		SizeBytes:      len(data),
		S3URL:          s3URL,
		UploadedAt:     now,
		UpdatedAt:      now,
	}
	if err := s.attRepo.Upsert(ctx, att); err != nil {
		s.log.WithField("err", err).Error("attachmentService.Upload: upsert error")
		return nil, err
	}
	return att, nil
}

func (s *AttachmentService) Download(ctx context.Context, taskID uuid.UUID) (io.ReadCloser, *model.TaskAttachment, error) {
	att, err := s.attRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	rc, err := s.store.Download(ctx, att.S3URL)
	if err != nil {
		if errors.Is(err, storage.ErrFeatureDisabled) {
			return nil, nil, ErrFeatureDisabled
		}
		return nil, nil, err
	}
	return rc, att, nil
}

func (s *AttachmentService) Delete(ctx context.Context, taskID uuid.UUID) error {
	att, err := s.attRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	if err := s.store.Delete(ctx, att.S3URL); err != nil && !errors.Is(err, storage.ErrFeatureDisabled) {
		return err
	}
	return s.attRepo.DeleteByTaskID(ctx, taskID)
}

func (s *AttachmentService) GetMeta(ctx context.Context, taskID uuid.UUID) (*model.TaskAttachment, error) {
	att, err := s.attRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return att, nil
}
