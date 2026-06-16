package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/workflow-engine/backend/internal/api"
	"github.com/workflow-engine/backend/internal/config"
	"github.com/workflow-engine/backend/internal/engine"
	"github.com/workflow-engine/backend/internal/repository"
	"github.com/workflow-engine/backend/internal/service"
	"github.com/workflow-engine/backend/internal/webhook"
	"github.com/workflow-engine/backend/internal/workflowloader"
)

func main() {
	cfg := config.Load()
	log := config.NewLogger(cfg)

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.WithField("err", err).Fatal("failed to open database")
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.WithField("err", err).Fatal("database ping failed")
	}
	log.Info("database connected")

	// Repos
	defRepo := repository.NewWorkflowDefRepo(db)
	instRepo := repository.NewWorkflowInstanceRepo(db)
	taskRepo := repository.NewTaskInstanceRepo(db)
	transRepo := repository.NewStateTransitionRepo(db)
	outboxRepo := repository.NewWebhookOutboxRepo(db)

	// Webhook
	subscribers, err := webhook.LoadRegistry(cfg.WebhookConfig)
	if err != nil {
		log.WithField("err", err).Warn("webhook registry load failed, continuing with no subscribers")
	}
	log.WithField("subscribers", len(subscribers)).Info("webhook registry loaded")
	worker := webhook.NewWorker(outboxRepo, subscribers, cfg.WebhookWorkerInterval, cfg.WebhookMaxAttempts, log)

	// Workflow loader
	loader := workflowloader.New(cfg.WorkflowsDir, defRepo, log)
	result, err := loader.Load(context.Background())
	if err != nil {
		log.WithField("err", err).Fatal("workflow loader failed")
	}
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			log.WithFields(logrus.Fields{"file": e.File, "err": e.Message}).Error("workflow def load error")
		}
		os.Exit(1)
	}
	log.WithFields(logrus.Fields{
		"loaded":   len(result.Loaded),
		"warnings": len(result.Warnings),
	}).Info("workflow definitions loaded at startup")

	// Engine + services
	eng := engine.New(db, defRepo, instRepo, taskRepo, transRepo, log)
	defSvc := service.NewWorkflowDefService(defRepo, instRepo, loader, log)
	instSvc := service.NewInstanceService(eng, instRepo, taskRepo, transRepo, log)
	taskSvc := service.NewTaskService(eng, taskRepo, log)

	// Router
	handler := api.NewRouter(db, defSvc, instSvc, taskSvc, log, cfg.CORSAllowedOrigins)
	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go worker.Run(ctx)

	go func() {
		log.WithField("addr", cfg.HTTPAddr).Info("workflow engine listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithField("err", err).Fatal("server error")
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.WithField("err", err).Error("HTTP shutdown error")
	}
	worker.Run(shutCtx) // drain remaining outbox
	log.Info("shutdown complete")
}
