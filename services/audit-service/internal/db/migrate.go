package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS audit_logs (
			id SERIAL PRIMARY KEY,
			task_id UUID,
			action_string TEXT NOT NULL,
			payload JSONB,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	return err
}
