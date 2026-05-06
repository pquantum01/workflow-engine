package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"workflow-engine/backend/internal/api"
	"workflow-engine/backend/internal/config"
	"workflow-engine/backend/internal/engine"
	"workflow-engine/backend/internal/repository/postgres"
	"workflow-engine/backend/internal/service"
)

func main() {
	cfg := config.Load()
	logger := config.SetupLogger(cfg.LogLevel)

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("database connected")

	// Repositories
	processRepo := postgres.NewProcessRepo(db)
	workflowRepo := postgres.NewWorkflowRepo(db)
	taskRepo := postgres.NewTaskRepo(db)
	processInstanceRepo := postgres.NewProcessInstanceRepo(db)
	workflowInstanceRepo := postgres.NewWorkflowInstanceRepo(db)
	taskInstanceRepo := postgres.NewTaskInstanceRepo(db)
	stateTransitionRepo := postgres.NewStateTransitionRepo(db)

	// Engine
	eng := engine.New(workflowRepo, workflowInstanceRepo, taskRepo, taskInstanceRepo, stateTransitionRepo, logger)

	// Services
	processSvc := service.NewProcessService(processRepo, logger)
	workflowSvc := service.NewWorkflowService(workflowRepo, processRepo, logger)
	taskSvc := service.NewTaskService(taskRepo, workflowRepo, processRepo, logger)
	processInstanceSvc := service.NewProcessInstanceService(processRepo, workflowRepo, taskRepo, processInstanceRepo, workflowInstanceRepo, taskInstanceRepo, logger)
	workflowInstanceSvc := service.NewWorkflowInstanceService(workflowInstanceRepo, stateTransitionRepo, eng, logger)
	taskInstanceSvc := service.NewTaskInstanceService(taskInstanceRepo, workflowInstanceRepo, eng, logger)

	// Router
	router := api.NewRouter(processSvc, workflowSvc, taskSvc, processInstanceSvc, workflowInstanceSvc, taskInstanceSvc, logger)

	addr := ":" + cfg.ServerPort
	logger.Info("starting server", slog.String("addr", addr))
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
