package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"plan-api/internal/database"
	"plan-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentApprovalRepository hanterar mänskliga beslut i agentkörningar.
//
// Plan äger godkännandena så att besluten finns kvar även när agenttjänsten
// är nere - och så att agenten aldrig kan godkänna sig själv.
type AgentApprovalRepository struct {
	pool *pgxpool.Pool
}

func NewAgentApprovalRepository() *AgentApprovalRepository {
	return &AgentApprovalRepository{pool: database.GetPool()}
}

const agentApprovalColumns = `aa.id, aa.agent_run_id, aa.goal_id, aa.title, aa.description, aa.status,
	aa.decided_by, aa.feedback, aa.consequence, aa.verification_plan, aa.payload, aa.created_at, aa.decided_at`

// List returnerar godkännanden för en körning, nyast först.
func (r *AgentApprovalRepository) List(ctx context.Context, userID, runID int64) ([]models.AgentApproval, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+agentApprovalColumns+`
		FROM agent_approvals aa
		JOIN agent_runs ar ON ar.id = aa.agent_run_id
		JOIN user_goals ug ON ug.goal_id = ar.goal_id
		WHERE aa.agent_run_id = $1 AND ug.user_id = $2
		ORDER BY aa.created_at DESC, aa.id DESC`, runID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent approvals: %w", err)
	}
	defer rows.Close()

	return scanAgentApprovals(rows)
}

// ListByGoal returnerar godkännanden för ett mål, nyast först.
//
// Ett mål kan ha flera godkännanden över tid - varje körning har sitt.
func (r *AgentApprovalRepository) ListByGoal(ctx context.Context, userID, goalID int64) ([]models.AgentApproval, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+agentApprovalColumns+`
		FROM agent_approvals aa
		JOIN agent_runs ar ON ar.id = aa.agent_run_id
		JOIN user_goals ug ON ug.goal_id = ar.goal_id
		WHERE ar.goal_id = $1 AND ug.user_id = $2
		ORDER BY aa.created_at DESC, aa.id DESC`, goalID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent approvals: %w", err)
	}
	defer rows.Close()

	return scanAgentApprovals(rows)
}

// ListPending returnerar de godkännanden som väntar på ett beslut.
//
// Det är listan en människa behöver se - alla beslut som väntar, oavsett
// vilket mål de gäller.
func (r *AgentApprovalRepository) ListPending(ctx context.Context, userID int64) ([]models.AgentApproval, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+agentApprovalColumns+`
		FROM agent_approvals aa
		JOIN agent_runs ar ON ar.id = aa.agent_run_id
		JOIN user_goals ug ON ug.goal_id = ar.goal_id
		WHERE ug.user_id = $1 AND aa.status = $2
		ORDER BY aa.created_at ASC, aa.id ASC`, userID, models.AgentApprovalPending)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending agent approvals: %w", err)
	}
	defer rows.Close()

	return scanAgentApprovals(rows)
}

// scanAgentApprovals läser en uppsättning godkännanden.
func scanAgentApprovals(rows pgx.Rows) ([]models.AgentApproval, error) {
	approvals := []models.AgentApproval{}
	for rows.Next() {
		approval, err := scanAgentApproval(rows)
		if err != nil {
			return nil, err
		}
		approvals = append(approvals, *approval)
	}
	return approvals, rows.Err()
}

// Get returnerar ett godkännande som tillhör användaren.
func (r *AgentApprovalRepository) Get(ctx context.Context, userID, approvalID int64) (*models.AgentApproval, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+agentApprovalColumns+`
		FROM agent_approvals aa
		JOIN agent_runs ar ON ar.id = aa.agent_run_id
		JOIN user_goals ug ON ug.goal_id = ar.goal_id
		WHERE aa.id = $1 AND ug.user_id = $2`, approvalID, userID)

	approval, err := scanAgentApproval(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return approval, nil
}

// Create sparar ett godkännande som väntar på ett mänskligt beslut.
func (r *AgentApprovalRepository) Create(ctx context.Context, userID int64, req models.CreateAgentApprovalRequest) (*models.AgentApproval, error) {
	payload, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode approval payload: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO agent_approvals (agent_run_id, goal_id, title, description, status, consequence, verification_plan, payload)
		SELECT ar.id, ar.goal_id, $2, $3, $4, $5, $6, $7
		FROM agent_runs ar
		JOIN user_goals ug ON ug.goal_id = ar.goal_id
		WHERE ar.id = $1 AND ug.user_id = $8
		RETURNING id, agent_run_id, goal_id, title, description, status,
			decided_by, feedback, consequence, verification_plan, payload,
			created_at, decided_at`,
		req.AgentRunID, req.Title, req.Description, models.AgentApprovalPending,
		req.Consequence, req.VerificationPlan, payload, userID)

	approval, err := scanAgentApproval(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to create agent approval: %w", err)
	}
	return approval, nil
}

// Decide registrerar ett mänskligt beslut.
//
// Ett redan avgjort godkännande kan inte ändras - beslutet ska vara
// entydigt i historiken.
func (r *AgentApprovalRepository) Decide(ctx context.Context, userID, approvalID int64, decision, feedback string) (*models.AgentApproval, error) {
	status, ok := approvalStatusForDecision(decision)
	if !ok {
		return nil, fmt.Errorf("%w: okänt beslut %q", models.ErrInvalidApproval, decision)
	}

	row := r.pool.QueryRow(ctx, `
		UPDATE agent_approvals aa
		SET status = $3, feedback = $4, decided_by = $5, decided_at = NOW()
		FROM agent_runs ar, user_goals ug
		WHERE aa.id = $1 AND aa.agent_run_id = ar.id AND ar.goal_id = ug.goal_id
		  AND ug.user_id = $5 AND aa.status = $2
		RETURNING aa.id, aa.agent_run_id, aa.goal_id, aa.title, aa.description, aa.status,
			aa.decided_by, aa.feedback, aa.consequence, aa.verification_plan, aa.payload,
			aa.created_at, aa.decided_at`,
		approvalID, models.AgentApprovalPending, status, feedback, userID)

	approval, err := scanAgentApproval(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to decide agent approval: %w", err)
	}
	return approval, nil
}

// PendingCount returnerar hur många beslut som väntar på en människa.
func (r *AgentApprovalRepository) PendingCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM agent_approvals aa
		JOIN agent_runs ar ON ar.id = aa.agent_run_id
		JOIN user_goals ug ON ug.goal_id = ar.goal_id
		WHERE ug.user_id = $1 AND aa.status = $2`, userID, models.AgentApprovalPending).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending approvals: %w", err)
	}
	return count, nil
}

// approvalStatusForDecision översätter ett beslut till en status.
func approvalStatusForDecision(decision string) (string, bool) {
	switch decision {
	case "approve":
		return models.AgentApprovalApproved, true
	case "reject":
		return models.AgentApprovalRejected, true
	default:
		return "", false
	}
}

// rowScanner läser en rad - antingen från en query eller från en tx.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanAgentApproval(row rowScanner) (*models.AgentApproval, error) {
	var approval models.AgentApproval
	var payload []byte
	err := row.Scan(&approval.ID, &approval.AgentRunID, &approval.GoalID, &approval.Title,
		&approval.Description, &approval.Status, &approval.DecidedBy, &approval.Feedback,
		&approval.Consequence, &approval.VerificationPlan, &payload,
		&approval.CreatedAt, &approval.DecidedAt)
	if err != nil {
		return nil, err
	}
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &approval.Payload); err != nil {
			return nil, fmt.Errorf("failed to decode approval payload: %w", err)
		}
	}
	return &approval, nil
}
