package tasktemplate

import (
	"encoding/json"
	"time"

	"example.com/taskservice/internal/domain"
	"github.com/google/uuid"
)

type TaskTemplate struct {
	ID          		uuid.UUID       `json:"id"`
	Title       		string          `json:"title"`
	Description 		string          `json:"description"`
	AssigneeID 			uuid.UUID       `json:"assignee_id"`
	RecurrenceType   	RecurrenceType  `json:"recurrence_type"`
	RecurrenceConfig 	json.RawMessage `json:"recurrence_config"`
	StartDate 			time.Time  		`json:"start_date"`
	EndDate   			*time.Time 		`json:"end_date,omitempty"`
	Status   			domain.Status  	`json:"status"`
	CreatedAt 			time.Time 		`json:"created_at"`
	UpdatedAt 			time.Time 		`json:"updated_at"`
}

type RecurrenceType string

const (
	Daily 			RecurrenceType = "DAILY"
	Monthly 		RecurrenceType = "MONTHLY"
	SpecificDates 	RecurrenceType = "SPECIFIC_DATES"
	DayParity		RecurrenceType = "DAY_PARITY"
)