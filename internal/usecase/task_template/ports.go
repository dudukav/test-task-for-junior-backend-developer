package tasktemplate

import (
	"context"
	"encoding/json"
	"time"

	"example.com/taskservice/internal/domain"
	taskdomain "example.com/taskservice/internal/domain/task"
	tasktemplatedomain "example.com/taskservice/internal/domain/task_template"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, template tasktemplatedomain.TaskTemplate) (*tasktemplatedomain.TaskTemplate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*tasktemplatedomain.TaskTemplate, error)
	Update(ctx context.Context, template *tasktemplatedomain.TaskTemplate) (*tasktemplatedomain.TaskTemplate, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*tasktemplatedomain.TaskTemplate, error)
	ListActive(ctx context.Context) ([]*tasktemplatedomain.TaskTemplate, error)
}

type GenerationRepository interface {
	CreateTaskFromTemplate(
		ctx context.Context,
		template tasktemplatedomain.TaskTemplate,
		date time.Time,
		slot string,
		task taskdomain.Task,
	) (bool, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*tasktemplatedomain.TaskTemplate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*tasktemplatedomain.TaskTemplate, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*tasktemplatedomain.TaskTemplate, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*tasktemplatedomain.TaskTemplate, error)
	ListActive(ctx context.Context) ([]*tasktemplatedomain.TaskTemplate, error)
	GenerateTasksForDate(ctx context.Context, date time.Time) error
}

type CreateInput struct {
	Title            string
	Description      string
	AssignedID       uuid.UUID
	RecurrenceType   tasktemplatedomain.RecurrenceType
	RecurrenceConfig json.RawMessage
	StartDate        time.Time
	EndDate          *time.Time
	Status           domain.Status
}

type UpdateInput struct {
	Title            *string
	Description      *string
	AssignedID       *uuid.UUID
	RecurrenceType   *tasktemplatedomain.RecurrenceType
	RecurrenceConfig *json.RawMessage
	StartDate        *time.Time
	EndDate          *time.Time
	Status           *domain.Status
}
