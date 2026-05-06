package tasktemplate

import (
	"encoding/json"
	"time"

	"example.com/taskservice/internal/domain"
	tasktemplatedomain "example.com/taskservice/internal/domain/task_template"
	"github.com/google/uuid"
)

type taskTemplateMutationDTO struct {
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	AssignedID       uuid.UUID       `json:"assignee_id"`
	RecurrenceType   string          `json:"recurrence_type"`
	RecurrenceConfig json.RawMessage `json:"recurrence_config"`
	StartDate        time.Time       `json:"start_date"`
	EndDate          *time.Time      `json:"end_date,omitempty"`
	Status           domain.Status   `json:"status"`
}

type taskTemplateDTO struct {
	ID               uuid.UUID       `json:"id"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	AssignedID       uuid.UUID       `json:"assignee_id"`
	RecurrenceType   string          `json:"recurrence_type"`
	RecurrenceConfig json.RawMessage `json:"recurrence_config"`
	StartDate        time.Time       `json:"start_date"`
	EndDate          *time.Time      `json:"end_date,omitempty"`
	Status           domain.Status   `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func newTaskTemplateDTO(template *tasktemplatedomain.TaskTemplate) taskTemplateDTO {
	return taskTemplateDTO{
		ID:               template.ID,
		Title:            template.Title,
		Description:      template.Description,
		AssignedID:       template.AssignedID,
		RecurrenceType:   string(template.RecurrenceType),
		RecurrenceConfig: template.RecurrenceConfig,
		StartDate:        template.StartDate,
		EndDate:          template.EndDate,
		Status:           template.Status,
		CreatedAt:        template.CreatedAt,
		UpdatedAt:        template.UpdatedAt,
	}
}
