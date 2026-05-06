CREATE TABLE IF NOT EXISTS task_template_generations (
    template_id UUID NOT NULL,
    generated_for_date DATE NOT NULL,
    recurrence_slot TEXT NOT NULL DEFAULT '',
    task_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (template_id, generated_for_date, recurrence_slot)
);
