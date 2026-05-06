package tasktemplate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/taskservice/internal/domain"
	taskdomain "example.com/taskservice/internal/domain/task"
	tasktemplatedomain "example.com/taskservice/internal/domain/task_template"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	integrationDateYear  = 2026
	integrationDateDay   = 6
	integrationDateHour  = 9
	integrationDateMonth = time.May

	expectedNoRows = 0
	expectedOneRow = 1
)

func TestGenerationRepositoryCreateTaskFromTemplateIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	ctx := context.Background()
	adminPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer adminPool.Close()

	schema := "generation_test_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	quotedSchema := pgx.Identifier{schema}.Sanitize()

	if _, err = adminPool.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	defer func() {
		_, _ = adminPool.Exec(context.Background(), "DROP SCHEMA "+quotedSchema+" CASCADE")
	}()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect test schema: %v", err)
	}
	defer pool.Close()

	applyMigrations(ctx, t, pool)

	repo := NewGeneration(pool)
	template := tasktemplatedomain.TaskTemplate{
		ID:          uuid.New(),
		Title:       "Check reports",
		Description: "Generated from template",
		Status:      domain.StatusNew,
	}
	generationDate := time.Date(
		integrationDateYear,
		integrationDateMonth,
		integrationDateDay,
		integrationDateHour,
		expectedNoRows,
		expectedNoRows,
		expectedNoRows,
		time.UTC,
	)
	slot := "09:00"

	firstTask := newGeneratedTask(template)
	created, err := repo.CreateTaskFromTemplate(ctx, template, generationDate, slot, firstTask)
	if err != nil {
		t.Fatalf("create task from template: %v", err)
	}
	if !created {
		t.Fatal("expected first generation to create a task")
	}

	assertRowCount(ctx, t, pool, "tasks", "id = $1", expectedOneRow, firstTask.ID)
	assertRowCount(ctx, t, pool, "task_template_generations", "template_id = $1 AND task_id = $2", expectedOneRow, template.ID, firstTask.ID)

	secondTask := newGeneratedTask(template)
	created, err = repo.CreateTaskFromTemplate(ctx, template, generationDate, slot, secondTask)
	if err != nil {
		t.Fatalf("create duplicate task from template: %v", err)
	}
	if created {
		t.Fatal("expected duplicate generation to be skipped")
	}

	assertRowCount(ctx, t, pool, "tasks", "id = $1", expectedNoRows, secondTask.ID)
	assertRowCount(ctx, t, pool, "task_template_generations", "template_id = $1 AND generated_for_date = $2", expectedOneRow, template.ID, generationDate)

	rollbackDate := generationDate.AddDate(expectedNoRows, expectedNoRows, expectedOneRow)
	_, err = repo.CreateTaskFromTemplate(ctx, template, rollbackDate, slot, firstTask)
	if err == nil {
		t.Fatal("expected duplicate task id to fail")
	}

	assertRowCount(ctx, t, pool, "task_template_generations", "template_id = $1 AND generated_for_date = $2", expectedNoRows, template.ID, rollbackDate)
}

func applyMigrations(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	migrationPaths := []string{
		filepath.Join("..", "..", "..", "..", "migrations", "0001_create_tasks.up.sql"),
		filepath.Join("..", "..", "..", "..", "migrations", "0002_create_tasks_template.up.sql"),
		filepath.Join("..", "..", "..", "..", "migrations", "0003_task_template_generations.up.sql"),
	}

	for _, path := range migrationPaths {
		query, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", path, err)
		}

		if _, err = pool.Exec(ctx, string(query)); err != nil {
			t.Fatalf("apply migration %s: %v", path, err)
		}
	}
}

func newGeneratedTask(template tasktemplatedomain.TaskTemplate) taskdomain.Task {
	now := time.Date(
		integrationDateYear,
		integrationDateMonth,
		integrationDateDay,
		integrationDateHour,
		expectedNoRows,
		expectedNoRows,
		expectedNoRows,
		time.UTC,
	)

	return taskdomain.Task{
		ID:          uuid.New(),
		Title:       template.Title,
		Description: template.Description,
		Status:      domain.StatusNew,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func assertRowCount(
	ctx context.Context,
	t *testing.T,
	pool *pgxpool.Pool,
	table string,
	where string,
	expected int,
	args ...any,
) {
	t.Helper()

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where)
	var count int
	if err := pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		t.Fatalf("count rows in %s: %v", table, err)
	}
	if count != expected {
		t.Fatalf("expected %d rows in %s, got %d", expected, table, count)
	}
}
