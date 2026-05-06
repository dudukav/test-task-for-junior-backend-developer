package handlers

import (
	"time"

	"example.com/taskservice/internal/domain"
	taskdomain "example.com/taskservice/internal/domain/task"
	"github.com/google/uuid"
)

type taskMutationDTO struct {
	Title       string            	`json:"title"`
	Description string            	`json:"description"`
	Status      domain.Status 		`json:"status"`
}

type taskDTO struct {
	ID          uuid.UUID           `json:"id"`
	Title       string            	`json:"title"`
	Description string            	`json:"description"`
	Status      domain.Status 		`json:"status"`
	CreatedAt   time.Time         	`json:"created_at"`
	UpdatedAt   time.Time         	`json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
