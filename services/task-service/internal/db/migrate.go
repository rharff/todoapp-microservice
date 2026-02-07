package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title TEXT NOT NULL,
			stage TEXT NOT NULL DEFAULT 'todo',
			position INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `ALTER TABLE tasks ADD COLUMN IF NOT EXISTS stage TEXT NOT NULL DEFAULT 'todo'`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `ALTER TABLE tasks ADD COLUMN IF NOT EXISTS position INT NOT NULL DEFAULT 0`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `DO $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'tasks' AND column_name = 'status'
		) THEN
			UPDATE tasks
			SET stage = CASE WHEN status = 'done' THEN 'done' ELSE 'todo' END
			WHERE stage IS NULL OR stage = '';
			ALTER TABLE tasks DROP COLUMN IF EXISTS status;
		END IF;
	END $$;`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_tasks_stage_position ON tasks (stage, position)`)
	return err
}
