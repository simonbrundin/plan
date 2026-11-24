package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"plan-agent/internal/core/checkpoint"
)

// CheckpointStore implements StateStore using PostgreSQL
type CheckpointStore struct {
	pool *pgxpool.Pool
}

// NewCheckpointStore creates a new CheckpointStore
func NewCheckpointStore(pool *pgxpool.Pool) *CheckpointStore {
	return &CheckpointStore{pool: pool}
}

// SavePlan saves a checkpoint plan (goal)
func (s *CheckpointStore) SavePlan(plan *checkpoint.CheckpointPlan) error {
	ctx := context.Background()

	// Upsert goal
	_, err := s.pool.Exec(ctx, `
		INSERT INTO goals (id, feature, description, status, approved_by, approved_at, rejected_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			feature = EXCLUDED.feature,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			approved_by = EXCLUDED.approved_by,
			approved_at = EXCLUDED.approved_at,
			rejected_at = EXCLUDED.rejected_at,
			updated_at = EXCLUDED.updated_at
	`, plan.ID, plan.Feature, plan.Description, string(plan.Status),
		plan.ApprovedBy, plan.ApprovedAt, plan.RejectedAt,
		plan.CreatedAt, plan.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save goal: %w", err)
	}

	// Save individual checkpoints
	for _, cp := range plan.Checkpoints {
		if err := s.SaveCheckpoint(cp); err != nil {
			return err
		}
	}

	return nil
}

// GetPlan retrieves a plan by ID
func (s *CheckpointStore) GetPlan(id string) (*checkpoint.CheckpointPlan, error) {
	ctx := context.Background()

	var plan checkpoint.CheckpointPlan
	var approvedAt, rejectedAt *time.Time

	err := s.pool.QueryRow(ctx, `
		SELECT id, feature, description, status, approved_by, approved_at, rejected_at, created_at, updated_at
		FROM goals WHERE id = $1
	`, id).Scan(&plan.ID, &plan.Feature, &plan.Description, &plan.Status,
		&plan.ApprovedBy, &approvedAt, &rejectedAt, &plan.CreatedAt, &plan.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %s", id)
	}

	if approvedAt != nil {
		plan.ApprovedAt = approvedAt
	}
	if rejectedAt != nil {
		plan.RejectedAt = rejectedAt
	}

	// Load checkpoints
	checkpoints, err := s.GetCheckpointsByPlan(id)
	if err != nil {
		return nil, err
	}
	plan.Checkpoints = checkpoints

	return &plan, nil
}

// GetAllPlans returns all plans
func (s *CheckpointStore) GetAllPlans() ([]*checkpoint.CheckpointPlan, error) {
	ctx := context.Background()

	rows, err := s.pool.Query(ctx, `
		SELECT id, feature, description, status, approved_by, approved_at, rejected_at, created_at, updated_at
		FROM goals ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*checkpoint.CheckpointPlan
	for rows.Next() {
		var plan checkpoint.CheckpointPlan
		var approvedAt, rejectedAt *time.Time

		if err := rows.Scan(&plan.ID, &plan.Feature, &plan.Description, &plan.Status,
			&plan.ApprovedBy, &approvedAt, &rejectedAt, &plan.CreatedAt, &plan.UpdatedAt); err != nil {
			return nil, err
		}

		if approvedAt != nil {
			plan.ApprovedAt = approvedAt
		}
		if rejectedAt != nil {
			plan.RejectedAt = rejectedAt
		}

		plans = append(plans, &plan)
	}

	// Load checkpoints for each plan
	for _, plan := range plans {
		checkpoints, err := s.GetCheckpointsByPlan(plan.ID)
		if err != nil {
			return nil, err
		}
		plan.Checkpoints = checkpoints
	}

	return plans, nil
}

// DeletePlan removes a plan
func (s *CheckpointStore) DeletePlan(id string) error {
	ctx := context.Background()
	_, err := s.pool.Exec(ctx, `DELETE FROM goals WHERE id = $1`, id)
	return err
}

// SaveCheckpoint saves a checkpoint
func (s *CheckpointStore) SaveCheckpoint(cp *checkpoint.Checkpoint) error {
	ctx := context.Background()

	// Determine goal_id from the checkpoint or plan context
	// For now, we'll store it as part of the checkpoint data
	gatesJSON, _ := json.Marshal(cp.Gates)
	tasksJSON, _ := json.Marshal(cp.Tasks)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO checkpoints (id, goal_id, name, description, complexity, gates, tasks, status, position, created_at, updated_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			complexity = EXCLUDED.complexity,
			gates = EXCLUDED.gates,
			tasks = EXCLUDED.tasks,
			status = EXCLUDED.status,
			position = EXCLUDED.position,
			updated_at = EXCLUDED.updated_at,
			completed_at = EXCLUDED.completed_at
	`, cp.ID, "", cp.Name, cp.Description, string(cp.Complexity),
		string(gatesJSON), string(tasksJSON), string(cp.Status), 0,
		cp.CreatedAt, cp.UpdatedAt, cp.CompletedAt)
	if err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}

	// Save gate results
	for gateType, result := range cp.GateResults {
		if err := s.SaveGateResult(cp.ID, result); err != nil {
			return err
		}
		_ = gateType // suppress unused warning
	}

	return nil
}

// GetCheckpoint retrieves a checkpoint by ID
func (s *CheckpointStore) GetCheckpoint(id string) (*checkpoint.Checkpoint, error) {
	ctx := context.Background()

	var cp checkpoint.Checkpoint
	var gatesJSON, tasksJSON []byte
	var completedAt *time.Time

	err := s.pool.QueryRow(ctx, `
		SELECT id, goal_id, name, description, complexity, gates, tasks, status, position, created_at, updated_at, completed_at
		FROM checkpoints WHERE id = $1
	`, id).Scan(&cp.ID, &uuid.Nil, &cp.Name, &cp.Description, &cp.Complexity,
		&gatesJSON, &tasksJSON, &cp.Status, new(int), &cp.CreatedAt, &cp.UpdatedAt, &completedAt)
	if err != nil {
		return nil, fmt.Errorf("checkpoint not found: %s", id)
	}

	json.Unmarshal(gatesJSON, &cp.Gates)
	json.Unmarshal(tasksJSON, &cp.Tasks)

	if completedAt != nil {
		cp.CompletedAt = completedAt
	}

	cp.GateResults = make(map[string]*checkpoint.GateResult)

	return &cp, nil
}

// GetCheckpointsByPlan returns all checkpoints for a plan
func (s *CheckpointStore) GetCheckpointsByPlan(planID string) ([]*checkpoint.Checkpoint, error) {
	ctx := context.Background()

	rows, err := s.pool.Query(ctx, `
		SELECT id, goal_id, name, description, complexity, gates, tasks, status, position, created_at, updated_at, completed_at
		FROM checkpoints WHERE goal_id = $1
		ORDER BY position ASC, created_at ASC
	`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checkpoints []*checkpoint.Checkpoint
	for rows.Next() {
		var cp checkpoint.Checkpoint
		var gatesJSON, tasksJSON []byte
		var completedAt *time.Time

		if err := rows.Scan(&cp.ID, &uuid.Nil, &cp.Name, &cp.Description, &cp.Complexity,
			&gatesJSON, &tasksJSON, &cp.Status, new(int), &cp.CreatedAt, &cp.UpdatedAt, &completedAt); err != nil {
			return nil, err
		}

		json.Unmarshal(gatesJSON, &cp.Gates)
		json.Unmarshal(tasksJSON, &cp.Tasks)

		if completedAt != nil {
			cp.CompletedAt = completedAt
		}

		cp.GateResults = make(map[string]*checkpoint.GateResult)

		// Load gate results for this checkpoint
		results, err := s.getGateResultsByCheckpoint(cp.ID)
		if err != nil {
			return nil, err
		}
		for _, result := range results {
			cp.GateResults[result.Type] = result
		}

		checkpoints = append(checkpoints, &cp)
	}

	return checkpoints, nil
}

// SaveGateResult saves a gate result for a checkpoint
func (s *CheckpointStore) SaveGateResult(checkpointID string, result *checkpoint.GateResult) error {
	ctx := context.Background()

	issuesJSON, _ := json.Marshal(result.Issues)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO checkpoint_results (id, checkpoint_id, gate_type, status, started_at, ended_at, duration, feedback, issues, summary)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			started_at = EXCLUDED.started_at,
			ended_at = EXCLUDED.ended_at,
			duration = EXCLUDED.duration,
			feedback = EXCLUDED.feedback,
			issues = EXCLUDED.issues,
			summary = EXCLUDED.summary
	`, uuid.New().String(), checkpointID, result.Type, result.Status,
		result.StartedAt, result.EndedAt, result.Duration, result.Feedback,
		string(issuesJSON), result.Summary)
	if err != nil {
		return fmt.Errorf("failed to save gate result: %w", err)
	}

	return nil
}

// GetGateResult retrieves a gate result
func (s *CheckpointStore) GetGateResult(checkpointID, gateType string) (*checkpoint.GateResult, error) {
	ctx := context.Background()

	var result checkpoint.GateResult
	var issuesJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT gate_type, status, started_at, ended_at, duration, feedback, issues, summary
		FROM checkpoint_results WHERE checkpoint_id = $1 AND gate_type = $2
	`, checkpointID, gateType).Scan(&result.Type, &result.Status,
		&result.StartedAt, &result.EndedAt, &result.Duration, &result.Feedback,
		&issuesJSON, &result.Summary)
	if err != nil {
		return nil, fmt.Errorf("gate result not found: %s/%s", checkpointID, gateType)
	}

	json.Unmarshal(issuesJSON, &result.Issues)

	return &result, nil
}

// getGateResultsByCheckpoint loads all gate results for a checkpoint
func (s *CheckpointStore) getGateResultsByCheckpoint(checkpointID string) ([]*checkpoint.GateResult, error) {
	ctx := context.Background()

	rows, err := s.pool.Query(ctx, `
		SELECT id, checkpoint_id, gate_type, status, started_at, ended_at, duration, feedback, issues, summary
		FROM checkpoint_results WHERE checkpoint_id = $1
	`, checkpointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*checkpoint.GateResult
	for rows.Next() {
		var result checkpoint.GateResult
		var issuesJSON []byte

		if err := rows.Scan(&result.Type, &result.Status,
			&result.StartedAt, &result.EndedAt, &result.Duration, &result.Feedback,
			&issuesJSON, &result.Summary); err != nil {
			return nil, err
		}

		json.Unmarshal(issuesJSON, &result.Issues)
		results = append(results, &result)
	}

	return results, nil
}
