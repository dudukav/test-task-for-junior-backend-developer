package task

import (
	"time"

	"example.com/taskservice/internal/domain"
	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID     	`json:"id"`
	Title       string    		`json:"title"`
	Description string    		`json:"description"`
	Status      domain.Status   `json:"status"`

	CreatedAt   time.Time 		`json:"created_at"`
	UpdatedAt   time.Time 		`json:"updated_at"`
}
