package tasktemplate

import (
	"context"
	"errors"
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	tasktemplatedomain "example.com/taskservice/internal/domain/task_template"
	tasktemplate "example.com/taskservice/internal/usecase/task_template"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GenerationRepository struct {
	pool *pgxpool.Pool
}

func NewGeneration(pool *pgxpool.Pool) tasktemplate.GenerationRepository {
	return &GenerationRepository{pool: pool}
}

func (r *GenerationRepository) CreateTaskFromTemplate(
	ctx context.Context,
	template tasktemplatedomain.TaskTemplate,
	date time.Time,
	slot string,
	task taskdomain.Task,
) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("%w: failed to begin transaction", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	reserved, err := reserveGeneration(ctx, tx, template.ID, date, slot)
	if err != nil {
		return false, err
	}
	if !reserved {
		return false, nil
	}

	if err = createGeneratedTask(ctx, tx, task); err != nil {
		return false, err
	}

	if err = completeGeneration(ctx, tx, template.ID, date, slot, task.ID); err != nil {
		return false, err
	}

	if err = tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("%w: failed to commit transaction", err)
	}

	return true, nil
}

func reserveGeneration(ctx context.Context, tx pgx.Tx, templateID uuid.UUID, date time.Time, slot string) (bool, error) {
	const query = `
		INSERT INTO task_template_generations (
			template_id,
			generated_for_date,
			recurrence_slot
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (template_id, generated_for_date, recurrence_slot) DO NOTHING
		RETURNING template_id
	`

	var returnedID uuid.UUID
	if err := tx.QueryRow(ctx, query, templateID, date, slot).Scan(&returnedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("%w: failed to reserve generation", err)
	}

	return true, nil
}

func createGeneratedTask(ctx context.Context, tx pgx.Tx, task taskdomain.Task) error {
	const query = `
		INSERT INTO tasks (id, title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := tx.Exec(ctx, query, task.ID, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt); err != nil {
		return fmt.Errorf("%w: failed to create generated task", err)
	}

	return nil
}

func completeGeneration(
	ctx context.Context,
	tx pgx.Tx,
	templateID uuid.UUID,
	date time.Time,
	slot string,
	taskID uuid.UUID,
) error {
	const query = `
		UPDATE task_template_generations
		SET task_id = $1
		WHERE template_id = $2
			AND generated_for_date = $3
			AND recurrence_slot = $4
	`

	result, err := tx.Exec(ctx, query, taskID, templateID, date, slot)
	if err != nil {
		return fmt.Errorf("%w: failed to complete generation", err)
	}
	if result.RowsAffected() == 0 {
		return tasktemplatedomain.ErrNotFound
	}

	return nil
}
