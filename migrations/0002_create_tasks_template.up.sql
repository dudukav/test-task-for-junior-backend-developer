CREATE TABLE IF NOT EXISTS task_templates (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    assignee_id UUID NOT NULL,
    recurrence_type TEXT NOT NULL,
    recurrence_config JSONB NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    status TEXT NOT NULL DEFAULT 'new',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_templates_status ON task_templates (status);
CREATE INDEX IF NOT EXISTS idx_task_templates_assignee_id ON task_templates (assignee_id);
CREATE INDEX IF NOT EXISTS idx_task_templates_start_date ON task_templates (start_date);
