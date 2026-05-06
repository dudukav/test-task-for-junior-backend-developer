package tasktemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"example.com/taskservice/internal/domain"
	tasktemplatedomain "example.com/taskservice/internal/domain/task_template"
	"github.com/google/uuid"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func New(repo Repository) Usecase {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) CreateTemplate(ctx context.Context, input CreateInput) (*tasktemplatedomain.TaskTemplate, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	model := tasktemplatedomain.TaskTemplate{
		ID:               uuid.New(),
		Title:            normalized.Title,
		Description:      normalized.Description,
		AssignedID:       normalized.AssigneeID,
		RecurrenceType:   normalized.RecurrenceType,
		RecurrenceConfig: normalized.RecurrenceConfig,
		StartDate:        normalized.StartDate,
		EndDate:          normalized.EndDate,
		Status:           normalized.Status,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	return s.repo.Create(ctx, model)
}

func (s *Service) GetTemplate(ctx context.Context, id uuid.UUID) (*tasktemplatedomain.TaskTemplate, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid id", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateTemplate(ctx context.Context, id uuid.UUID, input UpdateInput) (*tasktemplatedomain.TaskTemplate, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid id", ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	model := *existing
	if input.Title != nil {
		model.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		model.Description = strings.TrimSpace(*input.Description)
	}
	if input.AssigneeID != nil {
		model.AssignedID = *input.AssigneeID
	}
	if input.RecurrenceType != nil {
		model.RecurrenceType = *input.RecurrenceType
	}
	if input.RecurrenceConfig != nil {
		model.RecurrenceConfig = *input.RecurrenceConfig
	}
	if input.StartDate != nil {
		model.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		model.EndDate = input.EndDate
	}
	if input.Status != nil {
		model.Status = *input.Status
	}

	if err := validateTemplate(model); err != nil {
		return nil, err
	}

	model.UpdatedAt = s.now()
	return s.repo.Update(ctx, &model)
}

func (s *Service) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: invalid id", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) []*tasktemplatedomain.TaskTemplate {
	templates, err := s.repo.List(ctx)
	if err != nil {
		return []*tasktemplatedomain.TaskTemplate{}
	}

	return templates
}

func (s *Service) ListActive(ctx context.Context) []*tasktemplatedomain.TaskTemplate {
	templates, err := s.repo.ListActive(ctx)
	if err != nil {
		return []*tasktemplatedomain.TaskTemplate{}
	}

	return templates
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Status == "" {
		input.Status = domain.StatusNew
	}

	template := tasktemplatedomain.TaskTemplate{
		Title:            input.Title,
		Description:      input.Description,
		AssignedID:       input.AssigneeID,
		RecurrenceType:   input.RecurrenceType,
		RecurrenceConfig: input.RecurrenceConfig,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
		Status:           input.Status,
	}

	if err := validateTemplate(template); err != nil {
		return CreateInput{}, err
	}

	return input, nil
}

func validateTemplate(template tasktemplatedomain.TaskTemplate) error {
	if template.Title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if template.AssignedID == uuid.Nil {
		return fmt.Errorf("%w: assignee_id is required", ErrInvalidInput)
	}
	if template.StartDate.IsZero() {
		return fmt.Errorf("%w: start_date is required", ErrInvalidInput)
	}
	if template.EndDate != nil && template.EndDate.Before(template.StartDate) {
		return fmt.Errorf("%w: end_date must be greater than or equal to start_date", ErrInvalidInput)
	}
	if !template.Status.Valid() {
		return fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	if !validRecurrenceType(template.RecurrenceType) {
		return fmt.Errorf("%w: invalid recurrence_type", ErrInvalidInput)
	}
	if err := validateRecurrenceConfig(template.RecurrenceType, template.RecurrenceConfig); err != nil {
		return err
	}

	return nil
}

func validRecurrenceType(recurrenceType tasktemplatedomain.RecurrenceType) bool {
	switch recurrenceType {
	case tasktemplatedomain.Daily,
		tasktemplatedomain.Monthly,
		tasktemplatedomain.SpecificDates,
		tasktemplatedomain.DayParity:
		return true
	default:
		return false
	}
}

func validateRecurrenceConfig(recurrenceType tasktemplatedomain.RecurrenceType, raw json.RawMessage) error {
	if len(raw) == 0 || !json.Valid(raw) {
		return fmt.Errorf("%w: recurrence_config must be valid json", ErrInvalidInput)
	}

	var config map[string]json.RawMessage
	if err := json.Unmarshal(raw, &config); err != nil {
		return fmt.Errorf("%w: recurrence_config must be an object", ErrInvalidInput)
	}

	if len(config) == 0 {
		return fmt.Errorf("%w: recurrence_config is required", ErrInvalidInput)
	}

	switch recurrenceType {
	case tasktemplatedomain.Daily:
		interval, ok := readInt(config, "interval")
		if !ok || interval < 1 {
			return fmt.Errorf("%w: daily interval must be greater than zero", ErrInvalidInput)
		}
		
	case tasktemplatedomain.Monthly:
		day, ok := readInt(config, "day_of_month", "dayOfMonth", "dayofMonth")
		if !ok || day < 1 || day > 31 {
			return fmt.Errorf("%w: monthly day_of_month must be between 1 and 31", ErrInvalidInput)
		}

	case tasktemplatedomain.SpecificDates:
		date, ok := readString(config, "date")
		if !ok {
			return fmt.Errorf("%w: specific dates config requires date", ErrInvalidInput)
		}
		if _, err := time.Parse(dateLayout, date); err != nil {
			return fmt.Errorf("%w: date must use YYYY-MM-DD format", ErrInvalidInput)
		}

	case tasktemplatedomain.DayParity:
		parity, ok := readString(config, "parity")
		if !ok {
			return fmt.Errorf("%w: day parity config requires parity", ErrInvalidInput)
		}
		parity = strings.ToLower(strings.TrimSpace(parity))
		if parity != "odd" && parity != "even" {
			return fmt.Errorf("%w: parity must be odd or even", ErrInvalidInput)
		}
	}

	times, ok := readStrings(config, "times")
	if !ok || len(times) == 0 {
		return fmt.Errorf("%w: recurrence_config requires times", ErrInvalidInput)
	}
	for _, value := range times {
		if _, err := time.Parse(timeLayout, value); err != nil {
			return fmt.Errorf("%w: times must use HH:MM format", ErrInvalidInput)
		}
	}

	return nil
}

func readInt(config map[string]json.RawMessage, keys ...string) (int, bool) {
	for _, key := range keys {
		raw, ok := config[key]
		if !ok {
			continue
		}

		var value int
		if err := json.Unmarshal(raw, &value); err != nil {
			return 0, false
		}

		return value, true
	}

	return 0, false
}

func readString(config map[string]json.RawMessage, keys ...string) (string, bool) {
	for _, key := range keys {
		raw, ok := config[key]
		if !ok {
			continue
		}

		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", false
		}

		return strings.TrimSpace(value), true
	}

	return "", false
}

func readStrings(config map[string]json.RawMessage, key string) ([]string, bool) {
	raw, ok := config[key]
	if !ok {
		return nil, false
	}

	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, false
	}

	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}

	return values, true
}
