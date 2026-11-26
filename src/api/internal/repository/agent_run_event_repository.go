package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"plan/api/internal/database"
	"plan/api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentRunEventRepository hanterar händelser som tillhör en AgentRun.
type AgentRunEventRepository struct {
	pool *pgxpool.Pool
}

func NewAgentRunEventRepository() *AgentRunEventRepository {
	return &AgentRunEventRepository{pool: database.GetPool()}
}

const agentRunEventColumns = `id, agent_run_id, type, phase, payload, created_at`

// List returnerar händelser i kronologisk ordning efter åtkomstkontroll.
func (r *AgentRunEventRepository) List(ctx context.Context, userID, runID int64) ([]models.AgentRunEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+agentRunEventColumns+`
		FROM agent_run_events
		WHERE agent_run_id = $1
		  AND agent_run_id IN (
			SELECT ar.id FROM agent_runs ar
			JOIN user_goals ug ON ug.goal_id = ar.goal_id
			WHERE ug.user_id = $2
		  )
		ORDER BY created_at ASC, id ASC`, runID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent run events: %w", err)
	}
	defer rows.Close()

	events := []models.AgentRunEvent{}
	for rows.Next() {
		var event models.AgentRunEvent
		var payload []byte
		if err := rows.Scan(&event.ID, &event.AgentRunID, &event.Type, &event.Phase, &payload, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan agent run event: %w", err)
		}
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &event.Payload); err != nil {
				return nil, fmt.Errorf("failed to decode agent run event payload: %w", err)
			}
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

// Create sparar en händelse om AgentRun tillhör användaren.
func (r *AgentRunEventRepository) Create(ctx context.Context, userID, runID int64, req models.CreateAgentRunEventRequest) (*models.AgentRunEvent, error) {
	var event models.AgentRunEvent
	payload, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode event payload: %w", err)
	}

	err = r.pool.QueryRow(ctx, `
		INSERT INTO agent_run_events (agent_run_id, type, phase, payload)
		SELECT $1, $2, $3, $4
		WHERE EXISTS (
			SELECT 1 FROM agent_runs ar
			JOIN user_goals ug ON ug.goal_id = ar.goal_id
			WHERE ar.id = $1 AND ug.user_id = $5
		)
		RETURNING `+agentRunEventColumns, runID, req.Type, req.Phase, payload, userID).
		Scan(&event.ID, &event.AgentRunID, &event.Type, &event.Phase, &payload, &event.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to create agent run event: %w", err)
	}
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &event.Payload); err != nil {
			return nil, fmt.Errorf("failed to decode created event payload: %w", err)
		}
	}
	return &event, nil
}
