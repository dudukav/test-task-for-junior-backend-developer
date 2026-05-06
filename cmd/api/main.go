package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	"example.com/taskservice/internal/infrastructure/scheduler"
	postgrestask "example.com/taskservice/internal/repository/postgres/task"
	postgrestemplate "example.com/taskservice/internal/repository/postgres/task_template"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httptaskhandlers "example.com/taskservice/internal/transport/http/handlers/task"
	httptemplatehandlers "example.com/taskservice/internal/transport/http/handlers/task_template"
	"example.com/taskservice/internal/usecase/task"
	tasktemplate "example.com/taskservice/internal/usecase/task_template"
)

const (
	readHeaderTimeoutMin = 5
	contextTimeoutMin    = 10
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := run(logger); err != nil {
		logger.Error("run api", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer pool.Close()

	taskRepo := postgrestask.New(pool)
	taskUsecase := task.NewService(taskRepo)
	templateRepo := postgrestemplate.New(pool)
	generationRepo := postgrestemplate.NewGeneration(pool)
	templateUsecase := tasktemplate.NewWithGenerator(templateRepo, generationRepo)
	taskHandler := httptaskhandlers.NewTaskHandler(taskUsecase)
	docsHandler := swaggerdocs.NewHandler()
	templateHandler := httptemplatehandlers.NewTaskTemplateHandler(templateUsecase)
	router := transporthttp.NewRouter(taskHandler, templateHandler, docsHandler)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeoutMin * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), contextTimeoutMin*time.Second)
		defer cancel()

		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.Error("shutdown http server", "error", shutdownErr)
		}
	}()

	taskTemplateScheduler := scheduler.New(
		templateUsecase,
		logger,
		cfg.SchedulerLocation,
		cfg.SchedulerCronSpec,
	)

	go func() {
		if schedulerErr := taskTemplateScheduler.Run(ctx); schedulerErr != nil {
			logger.Error("scheduler stopped", "error", schedulerErr)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if listenErr := server.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", listenErr)
	}

	return nil
}
