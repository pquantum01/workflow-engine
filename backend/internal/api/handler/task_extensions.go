package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"workflow-engine/backend/internal/api/dto"
	"workflow-engine/backend/internal/service"
)

type TaskExtHandler struct {
	taskSvc    *service.TaskService
	commentSvc *service.CommentService
	attSvc     *service.AttachmentService
}

func NewTaskExtHandler(taskSvc *service.TaskService, commentSvc *service.CommentService, attSvc *service.AttachmentService) *TaskExtHandler {
	return &TaskExtHandler{taskSvc: taskSvc, commentSvc: commentSvc, attSvc: attSvc}
}

func (h *TaskExtHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	var req dto.PatchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}

	err = h.taskSvc.UpdateFields(r.Context(), id, service.TaskFieldPatch{
		Priority:               req.Priority,
		ExpectedCompletionDate: req.ExpectedCompletionDate,
		ExpectedCompletionTime: req.ExpectedCompletionTime,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeNotFound(w)
		case errors.Is(err, service.ErrTaskCompleted):
			writeError(w, http.StatusConflict, "TASK_COMPLETED", "task is terminal, fields locked", nil)
		default:
			writeInternalError(w)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskExtHandler) CreateAdhoc(w http.ResponseWriter, r *http.Request) {
	instID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	var req dto.CreateAdhocTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}
	if req.StateID == "" || req.Label == "" || req.TaskType == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "state_id, label, task_type are required", nil)
		return
	}

	task, err := h.taskSvc.CreateAdhoc(r.Context(), instID, service.AdhocTaskRequest{
		StateID:                req.StateID,
		Label:                  req.Label,
		TaskType:               req.TaskType,
		CreatedBy:              req.CreatedBy,
		Priority:               req.Priority,
		ExpectedCompletionDate: req.ExpectedCompletionDate,
		ExpectedCompletionTime: req.ExpectedCompletionTime,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeNotFound(w)
		case errors.Is(err, service.ErrInstanceNotActive):
			writeError(w, http.StatusConflict, "INSTANCE_TERMINAL", "instance is not active", nil)
		case errors.Is(err, service.ErrInvalidState):
			writeError(w, http.StatusUnprocessableEntity, "INVALID_STATE", "state_id not in workflow definition", nil)
		default:
			writeInternalError(w)
		}
		return
	}
	writeJSON(w, http.StatusCreated, toTaskResponse(task))
}

func (h *TaskExtHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	var req dto.AddCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}
	if req.Author == "" || req.Content == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "author and content are required", nil)
		return
	}
	if len(req.Content) > 2048 {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "content exceeds 2048 characters", nil)
		return
	}

	comment, err := h.commentSvc.Add(r.Context(), taskID, req.Author, req.Content)
	if err != nil {
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusCreated, dto.CommentResponse{
		ID:             comment.ID.String(),
		TaskInstanceID: comment.TaskInstanceID.String(),
		Author:         comment.Author,
		Content:        comment.Content,
		CreatedAt:      comment.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *TaskExtHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	comments, err := h.commentSvc.List(r.Context(), taskID)
	if err != nil {
		writeInternalError(w)
		return
	}
	out := make([]dto.CommentResponse, 0, len(comments))
	for _, c := range comments {
		out = append(out, dto.CommentResponse{
			ID:             c.ID.String(),
			TaskInstanceID: c.TaskInstanceID.String(),
			Author:         c.Author,
			Content:        c.Content,
			CreatedAt:      c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *TaskExtHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	var req dto.UploadAttachmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON", nil)
		return
	}
	if req.Filename == "" || req.ContentType == "" || req.Data == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "filename, content_type, data are required", nil)
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "data is not valid base64", nil)
		return
	}

	att, err := h.attSvc.Upload(r.Context(), taskID, req.Filename, req.ContentType, data)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFeatureDisabled):
			writeError(w, http.StatusServiceUnavailable, "FEATURE_DISABLED", "file upload is not enabled", nil)
		case errors.Is(err, service.ErrFileTooLarge):
			writeError(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "file exceeds 2MB limit", nil)
		default:
			writeInternalError(w)
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto.AttachmentMetaResponse{
		ID:          att.ID.String(),
		Filename:    att.Filename,
		ContentType: att.ContentType,
		SizeBytes:   att.SizeBytes,
		UploadedAt:  att.UploadedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   att.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *TaskExtHandler) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}

	rc, att, err := h.attSvc.Download(r.Context(), taskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeNotFound(w)
		case errors.Is(err, service.ErrFeatureDisabled):
			writeError(w, http.StatusServiceUnavailable, "FEATURE_DISABLED", "file download is not enabled", nil)
		default:
			writeInternalError(w)
		}
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", att.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", att.SizeBytes))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, rc) //nolint:errcheck // headers already sent
}

func (h *TaskExtHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}
	err = h.attSvc.Delete(r.Context(), taskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			writeNotFound(w)
		case errors.Is(err, service.ErrFeatureDisabled):
			writeError(w, http.StatusServiceUnavailable, "FEATURE_DISABLED", "file storage is not enabled", nil)
		default:
			writeInternalError(w)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
