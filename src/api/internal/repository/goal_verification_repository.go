package repository

import (
	"context"
	"fmt"

	"plan/api/internal/database"
	"plan/api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GoalVerificationRepository hanterar verifieringsplaner, bevis och
// verifieringsomdömen.
//
// Plan äger alltihop, så att bevisen finns kvar även när agenttjänsten är
// nere - och så att en människa alltid kan se varför ett mål räknas som
// verifierat.
type GoalVerificationRepository struct {
	pool *pgxpool.Pool
}

// NewGoalVerificationRepository ger en repository med databasen.
func NewGoalVerificationRepository() *GoalVerificationRepository {
	return &GoalVerificationRepository{pool: database.GetPool()}
}

const goalVerificationColumns = `id, goal_id, type, criterion, strength, created_at, updated_at`

// GetByGoal ger målets verifieringsplan, eller nil om målet inte har någon.
func (r *GoalVerificationRepository) GetByGoal(ctx context.Context, goalID int64) (*models.GoalVerification, error) {
	query := `SELECT ` + goalVerificationColumns + ` FROM goal_verifications WHERE goal_id = $1`

	var v models.GoalVerification
	err := r.pool.QueryRow(ctx, query, goalID).Scan(
		&v.ID, &v.GoalID, &v.Type, &v.Criterion, &v.Strength, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get goal verification: %w", err)
	}

	return &v, nil
}

// Upsert sätter målets verifieringsplan. Ett mål har högst en plan - en
// ny plan ersätter den gamla.
func (r *GoalVerificationRepository) Upsert(
	ctx context.Context, goalID int64, typ, criterion, strength string,
) (*models.GoalVerification, error) {
	query := `
		INSERT INTO goal_verifications (goal_id, type, criterion, strength)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (goal_id) DO UPDATE
			SET type = EXCLUDED.type,
				criterion = EXCLUDED.criterion,
				strength = EXCLUDED.strength,
				updated_at = NOW()
		RETURNING id, goal_id, type, criterion, strength, created_at, updated_at`

	var v models.GoalVerification
	err := r.pool.QueryRow(ctx, query, goalID, typ, criterion, strength).Scan(
		&v.ID, &v.GoalID, &v.Type, &v.Criterion, &v.Strength, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save goal verification: %w", err)
	}

	// En ändrad plan gör tidigare omdömen ogiltiga - de gäller ju den
	// gamla planen.
	if _, err := r.pool.Exec(ctx,
		`DELETE FROM goal_verification_results WHERE goal_id = $1`, goalID,
	); err != nil {
		return nil, fmt.Errorf("failed to clear verification results: %w", err)
	}

	return &v, nil
}

// Delete tar bort målets verifieringsplan.
func (r *GoalVerificationRepository) Delete(ctx context.Context, goalID int64) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM goal_verifications WHERE goal_id = $1`, goalID); err != nil {
		return fmt.Errorf("failed to delete goal verification: %w", err)
	}
	return nil
}

// AddEvidence sparar ett bevis på målet.
func (r *GoalVerificationRepository) AddEvidence(
	ctx context.Context, goalID int64, verificationID *int64, typ, value, source string,
) (*models.GoalEvidence, error) {
	query := `
		INSERT INTO goal_evidence (goal_id, verification_id, type, value, source)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, goal_id, verification_id, type, value, source, created_at`

	var e models.GoalEvidence
	err := r.pool.QueryRow(ctx, query, goalID, verificationID, typ, value, source).Scan(
		&e.ID, &e.GoalID, &e.VerificationID, &e.Type, &e.Value, &e.Source, &e.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add evidence: %w", err)
	}

	return &e, nil
}

// ListEvidence ger alla bevis för ett mål, äldsta först.
func (r *GoalVerificationRepository) ListEvidence(ctx context.Context, goalID int64) ([]models.GoalEvidence, error) {
	query := `
		SELECT id, goal_id, verification_id, type, value, source, created_at
		FROM goal_evidence
		WHERE goal_id = $1
		ORDER BY created_at, id`

	rows, err := r.pool.Query(ctx, query, goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to list evidence: %w", err)
	}
	defer rows.Close()

	var evidence []models.GoalEvidence
	for rows.Next() {
		var e models.GoalEvidence
		if err := rows.Scan(
			&e.ID, &e.GoalID, &e.VerificationID, &e.Type, &e.Value, &e.Source, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan evidence: %w", err)
		}
		evidence = append(evidence, e)
	}

	return evidence, nil
}

// CreateResult sparar en bedömning av en verifiering.
//
// Resultatet är inte godkänt än - det ska en människa bestämma.
func (r *GoalVerificationRepository) CreateResult(
	ctx context.Context, goalID int64, verificationID *int64, outcome, note string,
) (*models.GoalVerificationResult, error) {
	query := `
		INSERT INTO goal_verification_results (goal_id, verification_id, outcome, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id, goal_id, verification_id, outcome, note, decided_by, decided_at, created_at`

	var res models.GoalVerificationResult
	err := r.pool.QueryRow(ctx, query, goalID, verificationID, outcome, note).Scan(
		&res.ID, &res.GoalID, &res.VerificationID, &res.Outcome, &res.Note,
		&res.DecidedBy, &res.DecidedAt, &res.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create verification result: %w", err)
	}

	return &res, nil
}

// ApproveResult godkänner ett omdöme åt en människa.
func (r *GoalVerificationRepository) ApproveResult(
	ctx context.Context, resultID, userID int64,
) (*models.GoalVerificationResult, error) {
	query := `
		UPDATE goal_verification_results
		SET decided_by = $2, decided_at = NOW()
		WHERE id = $1
		RETURNING id, goal_id, verification_id, outcome, note, decided_by, decided_at, created_at`

	var res models.GoalVerificationResult
	err := r.pool.QueryRow(ctx, query, resultID, userID).Scan(
		&res.ID, &res.GoalID, &res.VerificationID, &res.Outcome, &res.Note,
		&res.DecidedBy, &res.DecidedAt, &res.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to approve verification result: %w", err)
	}

	return &res, nil
}

// LatestResult ger det senaste omdömet för ett mål, eller nil.
func (r *GoalVerificationRepository) LatestResult(ctx context.Context, goalID int64) (*models.GoalVerificationResult, error) {
	query := `
		SELECT id, goal_id, verification_id, outcome, note, decided_by, decided_at, created_at
		FROM goal_verification_results
		WHERE goal_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 1`

	var res models.GoalVerificationResult
	err := r.pool.QueryRow(ctx, query, goalID).Scan(
		&res.ID, &res.GoalID, &res.VerificationID, &res.Outcome, &res.Note,
		&res.DecidedBy, &res.DecidedAt, &res.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest verification result: %w", err)
	}

	return &res, nil
}

// HasPassedVerification svarar på om målet har en verifiering som gått
// igenom och godkänts av en människa.
//
// Ett omdöme som inte är godkänt räknas inte - en agent får inte godkänna
// sig själv.
func (r *GoalVerificationRepository) HasPassedVerification(ctx context.Context, goalID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM goal_verification_results
			WHERE goal_id = $1
				AND outcome = $2
				AND decided_by IS NOT NULL
		)`

	var ok bool
	if err := r.pool.QueryRow(ctx, query, goalID, models.VerificationOutcomePassed).Scan(&ok); err != nil {
		return false, fmt.Errorf("failed to check verification: %w", err)
	}

	return ok, nil
}
