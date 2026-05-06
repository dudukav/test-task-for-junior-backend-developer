package transporthttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"example.com/taskservice/internal/domain"
	tasktemplatedomain "example.com/taskservice/internal/domain/task_template"
	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgrestask "example.com/taskservice/internal/repository/postgres/task"
	postgrestemplate "example.com/taskservice/internal/repository/postgres/task_template"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httptaskhandlers "example.com/taskservice/internal/transport/http/handlers/task"
	httptemplatehandlers "example.com/taskservice/internal/transport/http/handlers/task_template"
	"example.com/taskservice/internal/usecase/task"
	tasktemplate "example.com/taskservice/internal/usecase/task_template"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresImage = "postgres:16-alpine"
	dbName        = "taskservice"
	dbUser        = "postgres"
	dbPassword    = "postgres"
)

func TestTaskTemplateServiceIntegration(t *testing.T) {
	ctx := context.Background()
	pool := startPostgres(t, ctx)
	defer pool.Close()

	applyMigrations(t, ctx, pool)

	templateUsecase := buildTaskTemplateUsecase(pool)
	server := httptest.NewServer(buildRouter(pool, templateUsecase))
	defer server.Close()

	assigneeID := uuid.New()
	createTemplateBody := map[string]any{
		"title":           "Daily report",
		"description":     "Check generated report",
		"assignee_id":     assigneeID.String(),
		"recurrence_type": string(tasktemplatedomain.Daily),
		"recurrence_config": map[string]any{
			"interval": 1,
			"times":    []string{"09:00"},
		},
		"start_date": time.Now().UTC().AddDate(0, 0, -1).Format(time.RFC3339),
		"status":     string(domain.StatusNew),
	}

	createTemplateResp := doJSON[taskTemplateResponse](t, server.URL+"/api/v1/task-templates", http.MethodPost, createTemplateBody, http.StatusCreated)
	if createTemplateResp.ID == uuid.Nil {
		t.Fatal("expected created task template id")
	}
	if createTemplateResp.AssignedID != assigneeID {
		t.Fatalf("expected assignee %s, got %s", assigneeID, createTemplateResp.AssignedID)
	}

	if err := templateUsecase.GenerateTasksForDate(ctx, time.Now().UTC()); err != nil {
		t.Fatalf("generate tasks for date: %v", err)
	}
	if err := templateUsecase.GenerateTasksForDate(ctx, time.Now().UTC()); err != nil {
		t.Fatalf("generate duplicate tasks for date: %v", err)
	}

	tasks := doJSON[[]taskResponse](t, server.URL+"/api/v1/tasks", http.MethodGet, nil, http.StatusOK)
	if len(tasks) != 1 {
		t.Fatalf("expected one generated task, got %d", len(tasks))
	}
	if tasks[0].Title != createTemplateBody["title"] {
		t.Fatalf("expected generated task title %q, got %q", createTemplateBody["title"], tasks[0].Title)
	}
	if tasks[0].Status != domain.StatusNew {
		t.Fatalf("expected generated task status %q, got %q", domain.StatusNew, tasks[0].Status)
	}

	templates := doJSON[[]taskTemplateResponse](t, server.URL+"/api/v1/task-templates", http.MethodGet, nil, http.StatusOK)
	if len(templates) != 1 {
		t.Fatalf("expected one task template, got %d", len(templates))
	}

	generationCount := countRows(t, ctx, pool, "task_template_generations")
	if generationCount != 1 {
		t.Fatalf("expected one generation record, got %d", generationCount)
	}
}

type taskTemplateResponse struct {
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

type taskResponse struct {
	ID          uuid.UUID     `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Status      domain.Status `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func startPostgres(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	container, err := postgres.Run(
		ctx,
		postgresImage,
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if terminateErr := testcontainers.TerminateContainer(container); terminateErr != nil {
			t.Fatalf("terminate postgres container: %v", terminateErr)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get postgres connection string: %v", err)
	}

	pool, err := infrastructurepostgres.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open postgres pool: %v", err)
	}

	return pool
}

func applyMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	migrations := []string{
		filepath.Join("..", "..", "..", "migrations", "0001_create_tasks.up.sql"),
		filepath.Join("..", "..", "..", "migrations", "0002_create_tasks_template.up.sql"),
		filepath.Join("..", "..", "..", "migrations", "0003_task_template_generations.up.sql"),
	}

	for _, migration := range migrations {
		query, err := os.ReadFile(migration)
		if err != nil {
			t.Fatalf("read migration %s: %v", migration, err)
		}
		if _, err = pool.Exec(ctx, string(query)); err != nil {
			t.Fatalf("apply migration %s: %v", migration, err)
		}
	}
}

func buildTaskTemplateUsecase(pool *pgxpool.Pool) tasktemplate.Usecase {
	templateRepo := postgrestemplate.New(pool)
	generationRepo := postgrestemplate.NewGeneration(pool)

	return tasktemplate.NewWithGenerator(templateRepo, generationRepo)
}

func buildRouter(pool *pgxpool.Pool, templateUsecase tasktemplate.Usecase) http.Handler {
	taskRepo := postgrestask.New(pool)
	taskUsecase := task.NewService(taskRepo)
	taskHandler := httptaskhandlers.NewTaskHandler(taskUsecase)
	templateHandler := httptemplatehandlers.NewTaskTemplateHandler(templateUsecase)
	docsHandler := swaggerdocs.NewHandler()

	return transporthttp.NewRouter(taskHandler, templateHandler, docsHandler)
}

func doJSON[T any](t *testing.T, url string, method string, payload any, expectedStatus int) T {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, &body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var result T
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return result
}

func countRows(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) int {
	t.Helper()

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	var count int
	if err := pool.QueryRow(ctx, query).Scan(&count); err != nil {
		t.Fatalf("count rows in %s: %v", table, err)
	}

	return count
}
