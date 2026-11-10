package database

import (
	"context"
	"fmt"
)

// EnsureSeedData makes sure the rows the application depends on exist.
//
// The schema is applied by the Atlas operator in declarative mode, which
// only handles DDL - the INSERT statements in the schema file are never
// executed. The application therefore seeds the data itself, which also
// keeps a freshly created database usable without manual steps.
//
// The function is idempotent and safe to run on every start.
func EnsureSeedData(ctx context.Context) error {
	pool := GetPool()
	if pool == nil {
		return fmt.Errorf("no database connection")
	}

	// The status workflow the goals use. The order matters: the goals table
	// defaults to status_id 1 (define).
	_, err := pool.Exec(ctx, `
		INSERT INTO statuses (name, label) VALUES
			('define', 'Defining'),
			('plan', 'Planning'),
			('execute', 'Executing'),
			('verify', 'Verifying'),
			('completed', 'Completed')
		ON CONFLICT (name) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed statuses: %w", err)
	}

	// The Root goal is the parent of all top level goals.
	_, err = pool.Exec(ctx, `
		INSERT INTO goals (id, title, icon) VALUES (1, 'Root', 'heroicons:home')
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("seed root goal: %w", err)
	}

	// The Root goal is inserted with an explicit id, which leaves the
	// identity sequence behind - without this the first goal created in the
	// app collides with the Root goal.
	_, err = pool.Exec(ctx, `
		SELECT setval(
			pg_get_serial_sequence('goals', 'id'),
			GREATEST((SELECT COALESCE(MAX(id), 0) + 1 FROM goals), 1)
		)
	`)
	if err != nil {
		return fmt.Errorf("reset goals sequence: %w", err)
	}

	return nil
}
