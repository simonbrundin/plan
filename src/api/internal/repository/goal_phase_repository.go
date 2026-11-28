package repository

import (
	"context"
	"fmt"

	"plan/api/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GoalPhaseRepository hämtar de fakta fasmaskinen behöver.
//
// Faktorad är enkla och frågade fram vid varje övergång - en maskin som
// fattar beslut utifrån färsk data blir inte fel av gammal cache.
type GoalPhaseRepository struct {
	pool *pgxpool.Pool
}

// NewGoalPhaseRepository ger en repository med databasen.
func NewGoalPhaseRepository() *GoalPhaseRepository {
	return &GoalPhaseRepository{pool: database.GetPool()}
}

// StatusIDByName ger status-id:t för en fas.
func (r *GoalPhaseRepository) StatusIDByName(ctx context.Context, name string) (int64, error) {
	query := `SELECT id FROM statuses WHERE name = $1`

	var id int64
	err := r.pool.QueryRow(ctx, query, name).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("okänd fas %q", name)
		}
		return 0, fmt.Errorf("failed to get status id: %w", err)
	}

	return id, nil
}

// HasApprovedDefinition svarar på om en människa har godkänt en
// agentdefinition för målet.
//
// Godkännandet ligger i agent_approvals, med typen goal_definition i
// payloaden. Ett godkänt kodförslag räknas inte - det säger ingenting om
// målets definition.
func (r *GoalPhaseRepository) HasApprovedDefinition(ctx context.Context, goalID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM agent_approvals
			WHERE goal_id = $1
				AND status = 'approved'
				AND payload->>'kind' = 'goal_definition'
		)`

	var ok bool
	if err := r.pool.QueryRow(ctx, query, goalID).Scan(&ok); err != nil {
		return false, fmt.Errorf("failed to check approved definition: %w", err)
	}

	return ok, nil
}

// ChildFacts svarar på två saker om målets delmål:
//
// hasPlanned - finns det minst ett delmål med datum eller uppskattad tid?
// allResolved - har varje delmål ett resultat?
//
// Ett delmål räknas som löst när det är avslutat. Ett mål utan delmål
// räknas som löst - då är det inget som saknas.
func (r *GoalPhaseRepository) ChildFacts(ctx context.Context, goalID int64) (hasPlanned, allResolved bool, err error) {
	query := `
		SELECT
			COALESCE(bool_or(g.due IS NOT NULL OR g.start_date IS NOT NULL OR g.estimated_minutes IS NOT NULL), FALSE),
			COALESCE(bool_and(s.name = 'completed'), TRUE)
		FROM goal_relations r
		JOIN goals g ON g.id = r.child_id
		JOIN statuses s ON s.id = g.status_id
		WHERE r.parent_id = $1`

	if err := r.pool.QueryRow(ctx, query, goalID).Scan(&hasPlanned, &allResolved); err != nil {
		return false, false, fmt.Errorf("failed to get child facts: %w", err)
	}

	return hasPlanned, allResolved, nil
}

// SetCompletionKind sparar hur ett mål avslutades.
func (r *GoalPhaseRepository) SetCompletionKind(ctx context.Context, goalID int64, kind string) error {
	query := `UPDATE goals SET completion_kind = $1 WHERE id = $2`

	if _, err := r.pool.Exec(ctx, query, kind, goalID); err != nil {
		return fmt.Errorf("failed to set completion kind: %w", err)
	}

	return nil
}

// RecordEvent loggar en fasövergång i målets historik.
//
// Loggen är grunden för "varför gjorde agenten det här?" - en tvingad
// övergång (bakåt eller ett fashopp) loggas med orsaken.
func (r *GoalPhaseRepository) RecordEvent(
	ctx context.Context, goalID int64, from, to, actor string, forced bool, reason string,
) error {
	query := `
		INSERT INTO goal_phase_events (goal_id, from_phase, to_phase, actor, forced, reason)
		VALUES ($1, $2, $3, $4, $5, $6)`

	if _, err := r.pool.Exec(ctx, query, goalID, from, to, actor, forced, reason); err != nil {
		return fmt.Errorf("failed to record phase event: %w", err)
	}

	return nil
}

// CompletionKind ger hur ett mål avslutades, eller tom sträng.
func (r *GoalPhaseRepository) CompletionKind(ctx context.Context, goalID int64) (string, error) {
	query := `SELECT COALESCE(completion_kind, '') FROM goals WHERE id = $1`

	var kind string
	err := r.pool.QueryRow(ctx, query, goalID).Scan(&kind)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get completion kind: %w", err)
	}

	return kind, nil
}
