package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type TaskTemplateGenerator interface {
	GenerateTasksForDate(ctx context.Context, date time.Time) error
}

type Scheduler struct {
	generator TaskTemplateGenerator
	logger    *slog.Logger
	location  *time.Location
	cronSpec  string
}

func New(
	generator TaskTemplateGenerator,
	logger *slog.Logger,
	location *time.Location,
	cronSpec string,
) *Scheduler {
	return &Scheduler{
		generator: generator,
		logger:    logger,
		location:  location,
		cronSpec:  cronSpec,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	c := cron.New(cron.WithLocation(s.location))

	_, err := c.AddFunc(s.cronSpec, func() {
		now := time.Now().In(s.location)

		if err := s.generator.GenerateTasksForDate(ctx, now); err != nil {
			s.logger.Error("generate tasks from templates", "error", err)
		}
	})

	if err != nil {
		return fmt.Errorf("%w: failed to add func to cron", err)
	}

	if err = s.generator.GenerateTasksForDate(ctx, time.Now().In(s.location)); err != nil {
		s.logger.Error("generate tasks from templates on startup", "error", err)
	}

	c.Start()

	<-ctx.Done()

	stopCtx := c.Stop()
	<-stopCtx.Done()

	return nil
}
