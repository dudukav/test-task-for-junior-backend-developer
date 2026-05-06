package tasktemplate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"example.com/taskservice/internal/domain"
	taskdomain "example.com/taskservice/internal/domain/task"
	tasktemplatedomain "example.com/taskservice/internal/domain/task_template"
	"github.com/google/uuid"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04"
	dayHours   = 24
	even       = 2
)

type Service struct {
	repo           Repository
	generationRepo GenerationRepository
	now            func() time.Time
}

func New(repo Repository) Usecase {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func NewWithGenerator(repo Repository, generationRepo GenerationRepository) Usecase {
	return &Service{
		repo:           repo,
		generationRepo: generationRepo,
		now:            func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*tasktemplatedomain.TaskTemplate, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	model := tasktemplatedomain.TaskTemplate{
		ID:               uuid.New(),
		Title:            normalized.Title,
		Description:      normalized.Description,
		AssignedID:       normalized.AssignedID,
		RecurrenceType:   normalized.RecurrenceType,
		RecurrenceConfig: normalized.RecurrenceConfig,
		StartDate:        normalized.StartDate,
		EndDate:          normalized.EndDate,
		Status:           normalized.Status,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	var task *tasktemplatedomain.TaskTemplate
	task, err = s.repo.Create(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create task template", err)
	}

	return task, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*tasktemplatedomain.TaskTemplate, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid id", ErrInvalidInput)
	}

	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get task template", err)
	}

	return task, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*tasktemplatedomain.TaskTemplate, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("%w: invalid id", ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get task template", err)
	}

	model := *existing
	if input.Title != nil {
		model.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		model.Description = strings.TrimSpace(*input.Description)
	}
	if input.AssignedID != nil {
		model.AssignedID = *input.AssignedID
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

	if err = validateTemplate(model); err != nil {
		return nil, err
	}

	model.UpdatedAt = s.now()
	var task *tasktemplatedomain.TaskTemplate
	task, err = s.repo.Update(ctx, &model)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update task template", err)
	}

	return task, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: invalid id", ErrInvalidInput)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%w: failed to delete task template", err)
	}

	return nil
}

func (s *Service) List(ctx context.Context) ([]*tasktemplatedomain.TaskTemplate, error) {
	templates, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get template list", err)
	}

	return templates, nil
}

func (s *Service) ListActive(ctx context.Context) ([]*tasktemplatedomain.TaskTemplate, error) {
	templates, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get active templates list", err)
	}

	return templates, nil
}

func (s *Service) GenerateTasksForDate(ctx context.Context, date time.Time) error {
	if s.generationRepo == nil {
		return errors.New("generation repository is required")
	}

	templates, err := s.ListActive(ctx)

	if err != nil {
		return fmt.Errorf("%w: failed to get active templates list", err)
	}

	for _, template := range templates {
		if !shouldGenerate(template, date) {
			continue
		}

		task := &taskdomain.Task{
			ID:          uuid.New(),
			Title:       template.Title,
			Description: template.Description,
			Status:      domain.StatusNew,
			CreatedAt:   s.now(),
			UpdatedAt:   s.now(),
		}

		_, err = s.generationRepo.CreateTaskFromTemplate(ctx, *template, date, "", *task)
		if err != nil {
			return fmt.Errorf("%w: failed to create task from template", err)
		}
	}

	return nil
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
		AssignedID:       input.AssignedID,
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

	switch recurrenceType {
	case tasktemplatedomain.Daily:
		var config tasktemplatedomain.DailyConfig
		if err := json.Unmarshal(raw, &config); err != nil {
			return fmt.Errorf("%w: daily recurrence_config must be an object", ErrInvalidInput)
		}
		if config.Interval < 1 {
			return fmt.Errorf("%w: daily interval must be greater than zero", ErrInvalidInput)
		}
		return validateTimes(config.Times)

	case tasktemplatedomain.Monthly:
		var config tasktemplatedomain.MonthlyConfig
		if err := json.Unmarshal(raw, &config); err != nil {
			return fmt.Errorf("%w: monthly recurrence_config must be an object", ErrInvalidInput)
		}
		if config.DayOfMonth < 1 || config.DayOfMonth > 31 {
			return fmt.Errorf("%w: monthly day_of_month must be between 1 and 31", ErrInvalidInput)
		}
		return validateTimes(config.Times)

	case tasktemplatedomain.SpecificDates:
		var config tasktemplatedomain.SpecificDatesConfig
		if err := json.Unmarshal(raw, &config); err != nil {
			return fmt.Errorf("%w: specific dates recurrence_config must be an object", ErrInvalidInput)
		}
		if strings.TrimSpace(config.Date) == "" {
			return fmt.Errorf("%w: specific dates config requires date", ErrInvalidInput)
		}
		if _, err := time.Parse(dateLayout, strings.TrimSpace(config.Date)); err != nil {
			return fmt.Errorf("%w: date must use YYYY-MM-DD format", ErrInvalidInput)
		}
		return validateTimes(config.Times)

	case tasktemplatedomain.DayParity:
		var config tasktemplatedomain.DayParityConfig
		if err := json.Unmarshal(raw, &config); err != nil {
			return fmt.Errorf("%w: day parity recurrence_config must be an object", ErrInvalidInput)
		}
		parity := strings.ToLower(strings.TrimSpace(config.Parity))
		if parity == "" {
			return fmt.Errorf("%w: day parity config requires parity", ErrInvalidInput)
		}
		if parity != "odd" && parity != "even" {
			return fmt.Errorf("%w: parity must be odd or even", ErrInvalidInput)
		}
		return validateTimes(config.Times)
	}

	return nil
}

func validateTimes(times []string) error {
	if len(times) == 0 {
		return fmt.Errorf("%w: recurrence_config requires times", ErrInvalidInput)
	}
	for _, value := range times {
		if _, err := time.Parse(timeLayout, strings.TrimSpace(value)); err != nil {
			return fmt.Errorf("%w: times must use HH:MM format", ErrInvalidInput)
		}
	}

	return nil
}

func shouldGenerate(template *tasktemplatedomain.TaskTemplate, date time.Time) bool {
	if template == nil {
		return false
	}
	if template.Status == domain.StatusDone {
		return false
	}

	targetDate := truncateDate(date)
	startDate := truncateDate(template.StartDate)
	if targetDate.Before(startDate) {
		return false
	}
	if template.EndDate != nil && targetDate.After(truncateDate(*template.EndDate)) {
		return false
	}

	switch template.RecurrenceType {
	case tasktemplatedomain.Daily:
		var config tasktemplatedomain.DailyConfig
		if err := json.Unmarshal(template.RecurrenceConfig, &config); err != nil || config.Interval < 1 {
			return false
		}

		daysFromStart := int(targetDate.Sub(startDate).Hours() / dayHours)
		return daysFromStart%config.Interval == 0

	case tasktemplatedomain.Monthly:
		var config tasktemplatedomain.MonthlyConfig
		if err := json.Unmarshal(template.RecurrenceConfig, &config); err != nil {
			return false
		}

		return targetDate.Day() == config.DayOfMonth

	case tasktemplatedomain.SpecificDates:
		var config tasktemplatedomain.SpecificDatesConfig
		if err := json.Unmarshal(template.RecurrenceConfig, &config); err != nil {
			return false
		}

		specificDate, err := time.Parse(dateLayout, strings.TrimSpace(config.Date))
		if err != nil {
			return false
		}

		return targetDate.Equal(truncateDate(specificDate))

	case tasktemplatedomain.DayParity:
		var config tasktemplatedomain.DayParityConfig
		if err := json.Unmarshal(template.RecurrenceConfig, &config); err != nil {
			return false
		}

		parity := strings.ToLower(strings.TrimSpace(config.Parity))
		isEvenDay := targetDate.Day()%even == 0
		return parity == "even" && isEvenDay || parity == "odd" && !isEvenDay

	default:
		return false
	}
}

func truncateDate(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
