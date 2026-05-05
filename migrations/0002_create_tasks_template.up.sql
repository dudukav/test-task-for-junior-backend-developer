CREATE TABLE IF NOT EXISTS task_templates (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    assignee_id UUID NOT NULL,

    recurrence_type TEXT NOT NULL,
    recurrence_config JSONB NOT NULL,

    start_date DATE NOT NULL,
    end_date DATE,

    is_active BOOLEAN NOT NULL DEFAULT TRUE
);