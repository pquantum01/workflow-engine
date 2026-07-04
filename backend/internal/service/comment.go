package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type CommentService struct {
	commentRepo repository.TaskCommentRepo
	log         *logrus.Logger
}

func NewCommentService(commentRepo repository.TaskCommentRepo, log *logrus.Logger) *CommentService {
	return &CommentService{commentRepo: commentRepo, log: log}
}

func (s *CommentService) Add(ctx context.Context, taskID uuid.UUID, author, content string) (*model.TaskComment, error) {
	c := &model.TaskComment{
		ID:             uuid.New(),
		TaskInstanceID: taskID,
		Author:         author,
		Content:        content,
		CreatedAt:      time.Now(),
	}
	if err := s.commentRepo.Add(ctx, c); err != nil {
		s.log.WithField("err", err).Error("commentService.Add")
		return nil, err
	}
	return c, nil
}

func (s *CommentService) List(ctx context.Context, taskID uuid.UUID) ([]*model.TaskComment, error) {
	return s.commentRepo.ListByTask(ctx, taskID)
}
