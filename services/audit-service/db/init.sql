CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    task_id UUID,
    action_string TEXT NOT NULL,
    payload JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
