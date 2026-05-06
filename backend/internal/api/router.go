package api

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"workflow-engine/backend/internal/api/handler"
	"workflow-engine/backend/internal/api/middleware"
	"workflow-engine/backend/internal/service"
)

// NewRouter creates an http.Handler with all API routes and middleware.
// CORS wraps the router externally so OPTIONS preflight requests are handled
// before gorilla/mux route matching.
func NewRouter(
	processSvc *service.ProcessService,
	workflowSvc *service.WorkflowService,
	taskSvc *service.TaskService,
	processInstanceSvc *service.ProcessInstanceService,
	workflowInstanceSvc *service.WorkflowInstanceService,
	taskInstanceSvc *service.TaskInstanceService,
	logger *slog.Logger,
) http.Handler {
	r := mux.NewRouter()

	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logging(logger))

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	v1 := r.PathPrefix("/api/v1").Subrouter()

	// Process definitions
	ph := handler.NewProcessHandler(processSvc, logger)
	v1.HandleFunc("/processes", ph.Create).Methods("POST")
	v1.HandleFunc("/processes", ph.List).Methods("GET")
	v1.HandleFunc("/processes/{id}", ph.GetByID).Methods("GET")
	v1.HandleFunc("/processes/{id}", ph.Update).Methods("PUT")
	v1.HandleFunc("/processes/{id}", ph.Delete).Methods("DELETE")

	// Workflow definitions
	wh := handler.NewWorkflowHandler(workflowSvc, logger)
	v1.HandleFunc("/processes/{id}/workflows", wh.Create).Methods("POST")
	v1.HandleFunc("/processes/{id}/workflows", wh.ListByProcessID).Methods("GET")
	v1.HandleFunc("/workflows/{id}", wh.GetByID).Methods("GET")
	v1.HandleFunc("/workflows/{id}", wh.Update).Methods("PUT")
	v1.HandleFunc("/workflows/{id}", wh.Delete).Methods("DELETE")

	// Task definitions
	th := handler.NewTaskHandler(taskSvc, logger)
	v1.HandleFunc("/workflows/{id}/tasks", th.Create).Methods("POST")
	v1.HandleFunc("/workflows/{id}/tasks", th.ListByWorkflowID).Methods("GET")
	v1.HandleFunc("/tasks/{id}", th.GetByID).Methods("GET")
	v1.HandleFunc("/tasks/{id}", th.Update).Methods("PUT")
	v1.HandleFunc("/tasks/{id}", th.Delete).Methods("DELETE")

	// Process instances
	pih := handler.NewProcessInstanceHandler(processInstanceSvc, logger)
	v1.HandleFunc("/process-instances", pih.Start).Methods("POST")
	v1.HandleFunc("/process-instances", pih.List).Methods("GET")
	v1.HandleFunc("/process-instances/{id}", pih.GetByID).Methods("GET")
	v1.HandleFunc("/process-instances/{id}/cancel", pih.Cancel).Methods("POST")
	v1.HandleFunc("/process-instances/{id}/children", pih.ListChildren).Methods("GET")

	// Workflow instances
	wih := handler.NewWorkflowInstanceHandler(workflowInstanceSvc, logger)
	v1.HandleFunc("/process-instances/{id}/workflows", wih.ListByProcessInstanceID).Methods("GET")
	v1.HandleFunc("/workflow-instances/{id}", wih.GetByID).Methods("GET")
	v1.HandleFunc("/workflow-instances/{id}/transition", wih.TriggerTransition).Methods("POST")
	v1.HandleFunc("/workflow-instances/{id}/transitions", wih.ListTransitions).Methods("GET")

	// Task instances
	tih := handler.NewTaskInstanceHandler(taskInstanceSvc, logger)
	v1.HandleFunc("/workflow-instances/{id}/tasks", tih.ListByWorkflowInstanceID).Methods("GET")
	v1.HandleFunc("/task-instances/{id}", tih.GetByID).Methods("GET")
	v1.HandleFunc("/task-instances/{id}/status", tih.UpdateStatus).Methods("PUT")
	v1.HandleFunc("/task-instances/{id}/assign", tih.Assign).Methods("PUT")

	return middleware.CORS(r)
}
