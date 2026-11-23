package repository

import (
	"context"
	"fmt"

	"plan-api/internal/database"
	"plan-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentRunRepository hanterar agentkörningar - kopplingen mellan ett mål
// och det arbete en agent gör med det.
//
// En körning tillhör en användare via sitt mål. Agenten autentiserar med
// en API-nyckel som är scopad till samma användare, så agenten ser bara
// sina egna körningar.
type AgentRunRepository struct {
	pool *pgxpool.Pool
}

func NewAgentRunRepository() *AgentRunRepository {
	return &AgentRunRepository{
		pool: database.GetPool(),
	}
}

const agentRunColumns = `ar.id, ar.goal_id, ar.phase, ar.status, COALESCE(ar.summary, ''),
	ar.artifact_url, ar.cost_usd, ar.conversation_id, ar.created_at, ar.updated_at`

// List returnerar användarens agentkörningar, nyast först.
//
// Om goalID anges returneras bara körningar för det målet.
func (r *AgentRunRepository) List(ctx context.Context, userID int64, goalID *int64) ([]models.AgentRun, error) {
	query := `SELECT ` + agentRunColumns + `
		FROM agent_runs ar
		INNER JOIN user_goals ug ON ug.goal_id = ar.goal_id
		WHERE ug.user_id = $1`
	args := []interface{}{userID}

	if goalID != nil {
		query += ` AND ar.goal_id = $2`
		args = append(args, *goalID)
	}
	query += ` ORDER BY ar.created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent runs: %w", err)
	}
	defer rows.Close()

	runs := []models.AgentRun{}
	for rows.Next() {
		var run models.AgentRun
		if err := rows.Scan(&run.ID, &run.GoalID, &run.Phase, &run.Status, &run.Summary,
			&run.ArtifactURL, &run.CostUSD, &run.ConversationID, &run.CreatedAt, &run.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan agent run: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, rows.Err()
}

// Get returnerar en agentkörning som tillhör användaren.
func (r *AgentRunRepository) Get(ctx context.Context, userID, runID int64) (*models.AgentRun, error) {
	query := `SELECT ` + agentRunColumns + `
		FROM agent_runs ar
		INNER JOIN user_goals ug ON ug.goal_id = ar.goal_id
		WHERE ug.user_id = $1 AND ar.id = $2`

	var run models.AgentRun
	err := r.pool.QueryRow(ctx, query, userID, runID).Scan(
		&run.ID, &run.GoalID, &run.Phase, &run.Status, &run.Summary,
		&run.ArtifactURL, &run.CostUSD, &run.ConversationID, &run.CreatedAt, &run.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent run: %w", err)
	}

	return &run, nil
}

// Create sparar en ny agentkörning för ett mål.
//
// Målet måste tillhöra användaren - annars avbryts skapandet.
func (r *AgentRunRepository) Create(ctx context.Context, userID int64, req models.CreateAgentRunRequest) (*models.AgentRun, error) {
	hasAccess, err := r.goalAccessible(ctx, userID, req.GoalID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, fmt.Errorf("goal %d does not belong to user", req.GoalID)
	}

	phase := req.Phase
	if phase == "" {
		phase = models.AgentRunPhaseDefine
	}
	status := req.Status
	if status == "" {
		status = models.AgentRunStatusQueued
	}

	query := `INSERT INTO agent_runs (goal_id, phase, status, summary, artifact_url, conversation_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, goal_id, phase, status, COALESCE(summary, ''), artifact_url, cost_usd, conversation_id, created_at, updated_at`

	var run models.AgentRun
	err = r.pool.QueryRow(ctx, query, req.GoalID, phase, status, req.Summary,
		req.ArtifactURL, req.ConversationID).Scan(
		&run.ID, &run.GoalID, &run.Phase, &run.Status, &run.Summary,
		&run.ArtifactURL, &run.CostUSD, &run.ConversationID, &run.CreatedAt, &run.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent run: %w", err)
	}

	return &run, nil
}

// Update uppdaterar en agentkörning.
//
// Endast de fält som sätts i req ändras - resten lämnas som de var.
func (r *AgentRunRepository) Update(ctx context.Context, userID, runID int64, req models.UpdateAgentRunRequest) (*models.AgentRun, error) {
	query := `UPDATE agent_runs SET `
	args := []interface{}{}
	argNum := 1

	if req.Phase != nil {
		query += fmt.Sprintf("phase = $%d, ", argNum)
		args = append(args, *req.Phase)
		argNum++
	}
	if req.Status != nil {
		query += fmt.Sprintf("status = $%d, ", argNum)
		args = append(args, *req.Status)
		argNum++
	}
	if req.Summary != nil {
		query += fmt.Sprintf("summary = $%d, ", argNum)
		args = append(args, *req.Summary)
		argNum++
	}
	if req.ArtifactURL != nil {
		query += fmt.Sprintf("artifact_url = $%d, ", argNum)
		args = append(args, *req.ArtifactURL)
		argNum++
	}
	if req.CostUSD != nil {
		query += fmt.Sprintf("cost_usd = $%d, ", argNum)
		args = append(args, *req.CostUSD)
		argNum++
	}
	if req.ConversationID != nil {
		query += fmt.Sprintf("conversation_id = $%d, ", argNum)
		args = append(args, *req.ConversationID)
		argNum++
	}

	if len(args) == 0 {
		return r.Get(ctx, userID, runID)
	}

	query += fmt.Sprintf("updated_at = NOW() WHERE id = $%d AND goal_id IN (SELECT goal_id FROM user_goals WHERE user_id = $%d)",
		argNum, argNum+1)
	query += " RETURNING id, goal_id, phase, status, COALESCE(summary, ''), artifact_url, cost_usd, conversation_id, created_at, updated_at"
	args = append(args, runID, userID)

	var run models.AgentRun
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&run.ID, &run.GoalID, &run.Phase, &run.Status, &run.Summary,
		&run.ArtifactURL, &run.CostUSD, &run.ConversationID, &run.CreatedAt, &run.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update agent run: %w", err)
	}

	return &run, nil
}

// goalAccessible berättar om ett mål tillhör användaren.
func (r *AgentRunRepository) goalAccessible(ctx context.Context, userID, goalID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_goals WHERE user_id = $1 AND goal_id = $2)`,
		userID, goalID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check goal access: %w", err)
	}
	return exists, nil
}
