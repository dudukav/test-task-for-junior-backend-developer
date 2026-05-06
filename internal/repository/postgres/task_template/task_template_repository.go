package tasktemplate

import (
	"context"
	"errors"
	"fmt"

	"example.com/taskservice/internal/domain"
	tasktemplatedomain "example.com/taskservice/internal/domain/task_template"
	tasktemplate "example.com/taskservice/internal/usecase/task_template"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) tasktemplate.Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, template tasktemplatedomain.TaskTemplate) (*tasktemplatedomain.TaskTemplate, error) {
	const query = `
		INSERT INTO task_templates (
			id, title, description, assignee_id,
			recurrence_type, recurrence_config,
			start_date, end_date, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, title, description, assignee_id,
			recurrence_type, recurrence_config,
			start_date, end_date, status,
			created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx, query,
		template.ID,
		template.Title,
		template.Description,
		template.AssignedID,
		template.RecurrenceType,
		template.RecurrenceConfig,
		template.StartDate,
		template.EndDate,
		template.Status,
		template.CreatedAt,
		template.UpdatedAt,
	)
	created, err := scanTaskTemplate(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*tasktemplatedomain.TaskTemplate, error) {
	const query = `
		SELECT id, title, description, assignee_id,
			recurrence_type, recurrence_config,
			start_date, end_date, status,
			created_at, updated_at
		FROM task_templates
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTaskTemplate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, tasktemplatedomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, template *tasktemplatedomain.TaskTemplate) (*tasktemplatedomain.TaskTemplate, error) {
	const query = `
		UPDATE task_templates
		SET title = $1,
			description = $2,
			assignee_id = $3,
			recurrence_type = $4,
			recurrence_config = $5,
			start_date = $6,
			end_date = $7,
			status = $8,
			updated_at = $9
		WHERE id = $10
		RETURNING id, title, description, assignee_id,
			recurrence_type, recurrence_config,
			start_date, end_date, status,
			created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx, query,
		template.Title,
		template.Description,
		template.AssignedID,
		template.RecurrenceType,
		template.RecurrenceConfig,
		template.StartDate,
		template.EndDate,
		template.Status,
		template.UpdatedAt,
		template.ID,
	)
	updated, err := scanTaskTemplate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, tasktemplatedomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM task_templates
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%w: failed to exec task", err)
	}

	if result.RowsAffected() == 0 {
		return tasktemplatedomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]*tasktemplatedomain.TaskTemplate, error) {
	const query = `
		SELECT id, title, description, assignee_id,
			recurrence_type, recurrence_config,
			start_date, end_date, status,
			created_at, updated_at
		FROM task_templates
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read rows", err)
	}
	defer rows.Close()

	tasks := make([]*tasktemplatedomain.TaskTemplate, 0)
	for rows.Next() {
		var task *tasktemplatedomain.TaskTemplate
		task, err = scanTaskTemplate(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: failed to read rows", err)
	}

	return tasks, nil
}

func (r *Repository) ListActive(ctx context.Context) ([]*tasktemplatedomain.TaskTemplate, error) {
	const query = `
		SELECT id, title, description, assignee_id,
			recurrence_type, recurrence_config,
			start_date, end_date, status,
			created_at, updated_at
		FROM task_templates
		WHERE status != $1
		AND start_date <= CURRENT_DATE
		AND (end_date IS NULL OR end_date >= CURRENT_DATE)
		ORDER BY start_date ASC, created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, domain.StatusDone)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read rows", err)
	}
	defer rows.Close()

	tasks := make([]*tasktemplatedomain.TaskTemplate, 0)
	for rows.Next() {
		var task *tasktemplatedomain.TaskTemplate
		task, err = scanTaskTemplate(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: failed to read rows", err)
	}

	return tasks, nil
}

type taskTemplateScanner interface {
	Scan(dest ...any) error
}

func scanTaskTemplate(scanner taskTemplateScanner) (*tasktemplatedomain.TaskTemplate, error) {
	var (
		template       tasktemplatedomain.TaskTemplate
		recurrenceType string
		status         string
	)

	if err := scanner.Scan(
		&template.ID,
		&template.Title,
		&template.Description,
		&template.AssignedID,
		&recurrenceType,
		&template.RecurrenceConfig,
		&template.StartDate,
		&template.EndDate,
		&status,
		&template.CreatedAt,
		&template.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("%w: failed to scan task template", err)
	}

	template.RecurrenceType = tasktemplatedomain.RecurrenceType(recurrenceType)
	template.Status = domain.Status(status)

	return &template, nil
}
