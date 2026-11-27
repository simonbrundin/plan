package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// Connect creates a connection pool to PostgreSQL
func Connect() (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@postgres:5432/software_agent?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Auto-migrate: create tables and add missing columns
	if err := runMigrations(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return pool, nil
}

// GetPool returns the database pool
func GetPool() *pgxpool.Pool {
	return pool
}

// SetPool sets the database pool (for use by main)
func SetPool(p *pgxpool.Pool) {
	pool = p
}

// Close closes the database pool
func Close() {
	if pool != nil {
		pool.Close()
	}
}

// runMigrations creates tables and adds missing columns
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Create conversations table
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS conversations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			title VARCHAR(255) DEFAULT 'New conversation',
			model VARCHAR(100) DEFAULT 'MiniMax-M2.7',
			provider VARCHAR(50) DEFAULT 'minimax'
		)
	`); err != nil {
		return fmt.Errorf("failed to create conversations table: %w", err)
	}

	// Migration: conversations.goal_id - kopplingen till Plan-målet.
	// Kolumnen läggs till separat eftersom tabellen redan finns i
	// databaser som kör agenten sedan tidigare.
	if _, err := pool.Exec(ctx, `
		ALTER TABLE conversations ADD COLUMN IF NOT EXISTS goal_id BIGINT
	`); err != nil {
		return fmt.Errorf("failed to add conversations.goal_id: %w", err)
	}

	// Create messages table
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
			role VARCHAR(20) NOT NULL,
			content TEXT NOT NULL,
			tool_name VARCHAR(100) DEFAULT '',
			tool_args JSONB DEFAULT '{}',
			tool_result TEXT DEFAULT '',
			is_error BOOLEAN DEFAULT FALSE,
			thinking TEXT DEFAULT '',
			tokens_used INT DEFAULT 0,
			latency_ms INT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	// Create api_keys table
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS api_keys (
			id SERIAL PRIMARY KEY,
			provider VARCHAR(50) NOT NULL UNIQUE,
			api_key TEXT NOT NULL,
			model VARCHAR(100) DEFAULT '',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("failed to create api_keys table: %w", err)
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_created_at ON conversations(created_at)`,
		// Samtalen hittas tillbaka via målet - därför indexeras goal_id.
		`CREATE INDEX IF NOT EXISTS idx_conversations_goal_id ON conversations(goal_id)`,
	}

	for _, idx := range indexes {
		if _, err := pool.Exec(ctx, idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Migration: goals table (checkpoint plans)
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS goals (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			feature VARCHAR(500) NOT NULL,
			description TEXT DEFAULT '',
			status VARCHAR(50) DEFAULT 'pending',
			approved_by VARCHAR(255) DEFAULT '',
			approved_at TIMESTAMPTZ,
			rejected_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("failed to create goals table: %w", err)
	}

	// Migration: checkpoints table
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS checkpoints (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT DEFAULT '',
			complexity VARCHAR(20) DEFAULT 'medium',
			gates JSONB DEFAULT '[]',
			tasks JSONB DEFAULT '[]',
			status VARCHAR(50) DEFAULT 'pending',
			position INT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			completed_at TIMESTAMPTZ
		)
	`); err != nil {
		return fmt.Errorf("failed to create checkpoints table: %w", err)
	}

	// Migration: checkpoint_results table (gate results)
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS checkpoint_results (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			checkpoint_id UUID NOT NULL REFERENCES checkpoints(id) ON DELETE CASCADE,
			gate_type VARCHAR(50) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			started_at TIMESTAMPTZ DEFAULT NOW(),
			ended_at TIMESTAMPTZ,
			duration VARCHAR(100) DEFAULT '',
			feedback TEXT DEFAULT '',
			issues JSONB DEFAULT '[]',
			summary TEXT DEFAULT '',
			created_at TIMESTAMPTZ DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("failed to create checkpoint_results table: %w", err)
	}

	// Create indexes for new tables
	newIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_checkpoints_goal_id ON checkpoints(goal_id)`,
		`CREATE INDEX IF NOT EXISTS idx_checkpoints_status ON checkpoints(status)`,
		`CREATE INDEX IF NOT EXISTS idx_checkpoint_results_checkpoint_id ON checkpoint_results(checkpoint_id)`,
		`CREATE INDEX IF NOT EXISTS idx_checkpoint_results_gate_type ON checkpoint_results(gate_type)`,
	}

	for _, idx := range newIndexes {
		if _, err := pool.Exec(ctx, idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
