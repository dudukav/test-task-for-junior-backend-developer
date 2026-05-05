package task

import (
	"context"

	"example.com/taskservice/internal/domain"
	taskdomain "example.com/taskservice/internal/domain/task"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*taskdomain.Task, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type CreateInput struct {
	Title       string
	Description string
	Status      domain.Status
}

type UpdateInput struct {
	Title       string
	Description string
	Status      domain.Status
}
