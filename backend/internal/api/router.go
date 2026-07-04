package api

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"workflow-engine/backend/internal/api/handler"
	"workflow-engine/backend/internal/api/middleware"
	"workflow-engine/backend/internal/service"
)

func NewRouter(
	db *sql.DB,
	defSvc *service.WorkflowDefService,
	instSvc *service.InstanceService,
	taskSvc *service.TaskService,
	commentSvc *service.CommentService,
	attSvc *service.AttachmentService,
	log *logrus.Logger,
	corsOrigins string,
) http.Handler {
	r := mux.NewRouter()

	defHandler := handler.NewWorkflowDefHandler(defSvc)
	instHandler := handler.NewInstanceHandler(instSvc)
	taskHandler := handler.NewTaskHandler(taskSvc)
	taskExtHandler := handler.NewTaskExtHandler(taskSvc, commentSvc, attSvc)
	healthHandler := handler.NewHealthHandler(db)

	api := r.PathPrefix("/api/v1").Subrouter()

	// Workflow defs
	api.HandleFunc("/workflow-defs", defHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/workflow-defs/{id}", defHandler.Get).Methods(http.MethodGet)

	// Admin
	api.HandleFunc("/admin/reload-defs", defHandler.ReloadDefs).Methods(http.MethodPost)

	// Instances
	api.HandleFunc("/workflow-instances", instHandler.Start).Methods(http.MethodPost)
	api.HandleFunc("/workflow-instances", instHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/workflow-instances/{id}", instHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/workflow-instances/{id}/cancel", instHandler.Cancel).Methods(http.MethodPost)
	api.HandleFunc("/workflow-instances/{id}", instHandler.SoftDelete).Methods(http.MethodDelete)
	api.HandleFunc("/workflow-instances/{id}/tasks", instHandler.ListTasks).Methods(http.MethodGet)
	api.HandleFunc("/workflow-instances/{id}/history", instHandler.GetHistory).Methods(http.MethodGet)

	// Tasks
	api.HandleFunc("/task-instances/{id}", taskHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/task-instances/{id}", taskExtHandler.Patch).Methods(http.MethodPatch)
	api.HandleFunc("/task-instances/{id}/start", taskHandler.StartTask).Methods(http.MethodPost)
	api.HandleFunc("/task-instances/{id}/complete", taskHandler.Complete).Methods(http.MethodPost)
	api.HandleFunc("/task-instances/{id}/comments", taskExtHandler.ListComments).Methods(http.MethodGet)
	api.HandleFunc("/task-instances/{id}/comments", taskExtHandler.AddComment).Methods(http.MethodPost)
	api.HandleFunc("/task-instances/{id}/attachment", taskExtHandler.UploadAttachment).Methods(http.MethodPost)
	api.HandleFunc("/task-instances/{id}/attachment", taskExtHandler.DownloadAttachment).Methods(http.MethodGet)
	api.HandleFunc("/task-instances/{id}/attachment", taskExtHandler.DeleteAttachment).Methods(http.MethodDelete)
	api.HandleFunc("/workflow-instances/{id}/tasks", taskExtHandler.CreateAdhoc).Methods(http.MethodPost)
	api.HandleFunc("/task-types", taskHandler.ListTaskTypes).Methods(http.MethodGet)

	// Health
	r.HandleFunc("/healthz", healthHandler.Liveness).Methods(http.MethodGet)
	r.HandleFunc("/readyz", healthHandler.Readiness).Methods(http.MethodGet)

	// SPA fallback
	RegisterStatic(r)

	// Middleware chain
	chain := middleware.RequestID(
		middleware.Logging(log)(
			middleware.Recovery(log)(
				middleware.CORS(corsOrigins)(r),
			),
		),
	)
	return chain
}
