package repository

import (
	"context"
	"fmt"
	"time"

	"plan-api/internal/database"
	"plan-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentBudgetRepository hanterar budgetgränser för agentkostnad.
//
// En budget har en hård gräns som stoppar nya körningar och en
// varningsgräns som bara varnar. Kontrollen och reservationen sker i
// samma transaktion som körningens skapande, så att samtidiga körningar
// inte tillsammans kan överskrida gränsen.
type AgentBudgetRepository struct {
	pool *pgxpool.Pool
}

// NewAgentBudgetRepository skapar ett repository för budgetar.
func NewAgentBudgetRepository() *AgentBudgetRepository {
	return &AgentBudgetRepository{
		pool: database.GetPool(),
	}
}

// List returnerar användarens budgetar.
func (r *AgentBudgetRepository) List(ctx context.Context, userID int64) ([]models.AgentBudget, error) {
	query := `SELECT id, name, period, hard_limit_usd, warning_limit_usd,
		reservation_per_run_usd, agent_id, active, created_at
		FROM agent_budgets WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent budgets: %w", err)
	}
	defer rows.Close()

	budgets := []models.AgentBudget{}
	for rows.Next() {
		var b models.AgentBudget
		if err := rows.Scan(&b.ID, &b.Name, &b.Period, &b.HardLimitUSD, &b.WarningLimitUSD,
			&b.ReservationPerRunUSD, &b.AgentID, &b.Active, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan agent budget: %w", err)
		}
		budgets = append(budgets, b)
	}
	return budgets, rows.Err()
}

// Create sparar en ny budget.
func (r *AgentBudgetRepository) Create(ctx context.Context, userID int64, req models.CreateAgentBudgetRequest) (*models.AgentBudget, error) {
	if err := models.ValidateAgentBudgetPeriod(req.Period); err != nil {
		return nil, err
	}

	reservation := req.ReservationPerRunUSD
	if reservation <= 0 {
		reservation = 1
	}

	query := `INSERT INTO agent_budgets (user_id, name, period, hard_limit_usd, warning_limit_usd,
		reservation_per_run_usd, agent_id, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
		RETURNING id, name, period, hard_limit_usd, warning_limit_usd,
			reservation_per_run_usd, agent_id, active, created_at`

	var budget models.AgentBudget
	err := r.pool.QueryRow(ctx, query, userID, req.Name, req.Period, req.HardLimitUSD,
		req.WarningLimitUSD, reservation, models.NormalizeAgentID(req.AgentID)).Scan(
		&budget.ID, &budget.Name, &budget.Period, &budget.HardLimitUSD, &budget.WarningLimitUSD,
		&budget.ReservationPerRunUSD, &budget.AgentID, &budget.Active, &budget.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent budget: %w", err)
	}
	return &budget, nil
}

// Update ändrar en budget. Bara de fält som skickas ändras.
//
// En budget som nått sin gräns måste kunna pausas, annars kan agenten
// inte köra igen förrän perioden går ut.
func (r *AgentBudgetRepository) Update(ctx context.Context, userID, budgetID int64, req models.UpdateAgentBudgetRequest) (*models.AgentBudget, error) {
	if req.Period != nil {
		if err := models.ValidateAgentBudgetPeriod(*req.Period); err != nil {
			return nil, err
		}
	}

	query := `UPDATE agent_budgets SET
		name = COALESCE($3, name),
		period = COALESCE($4, period),
		hard_limit_usd = COALESCE($5, hard_limit_usd),
		warning_limit_usd = CASE WHEN $6::boolean THEN $7 ELSE warning_limit_usd END,
		reservation_per_run_usd = COALESCE($8, reservation_per_run_usd),
		active = COALESCE($9, active)
		WHERE id = $1 AND user_id = $2
		RETURNING id, name, period, hard_limit_usd, warning_limit_usd,
			reservation_per_run_usd, agent_id, active, created_at`

	var budget models.AgentBudget
	err := r.pool.QueryRow(ctx, query, budgetID, userID, req.Name, req.Period,
		req.HardLimitUSD, req.WarningLimitUSD != nil, req.WarningLimitUSD,
		req.ReservationPerRunUSD, req.Active).Scan(
		&budget.ID, &budget.Name, &budget.Period, &budget.HardLimitUSD, &budget.WarningLimitUSD,
		&budget.ReservationPerRunUSD, &budget.AgentID, &budget.Active, &budget.CreatedAt)
	if err != nil {
		if pgx.ErrNoRows == err {
			return nil, fmt.Errorf("budgeten hittades inte")
		}
		return nil, fmt.Errorf("failed to update agent budget: %w", err)
	}
	return &budget, nil
}

// Delete tar bort en budget.
func (r *AgentBudgetRepository) Delete(ctx context.Context, userID, budgetID int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM agent_budgets WHERE id = $1 AND user_id = $2`, budgetID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete agent budget: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("budgeten hittades inte")
	}
	return nil
}

// Status berättar hur en budget ligger, utan att låsa något.
//
// Den används för att visa budgetens läge i gränssnittet.
func (r *AgentBudgetRepository) Status(ctx context.Context, userID int64, budget models.AgentBudget) (models.BudgetStatus, error) {
	return budgetStatus(ctx, r.pool, userID, budget, time.Now())
}

// budgetStatus räknar ut en budgets läge.
//
// Spent är den kostnad som redan bokförts i perioden. Reserved är den
// kostnad som reserverats för körningar som fortfarande pågår.
func budgetStatus(ctx context.Context, q querier, userID int64, budget models.AgentBudget, now time.Time) (models.BudgetStatus, error) {
	start := periodStart(budget.Period, now)

	var spent float64
	err := q.QueryRow(ctx, `
		SELECT COALESCE(SUM(u.cost_usd), 0)
		FROM agent_run_usage u
		INNER JOIN user_goals ug ON ug.goal_id = u.goal_id
		WHERE ug.user_id = $1 AND u.created_at >= $2 AND u.created_at <= $3`,
		userID, start, now).Scan(&spent)
	if err != nil {
		return models.BudgetStatus{}, fmt.Errorf("failed to sum spent cost: %w", err)
	}

	var reserved float64
	err = q.QueryRow(ctx, `
		SELECT COALESCE(SUM(ar.reserved_cost_usd), 0)
		FROM agent_runs ar
		INNER JOIN user_goals ug ON ug.goal_id = ar.goal_id
		WHERE ug.user_id = $1
			AND ar.status IN ('queued', 'running', 'blocked')
			AND ar.reserved_cost_usd > 0
			AND ($2 = '' OR ar.agent_id = $2 OR ar.agent_id = '')`,
		userID, budget.AgentID).Scan(&reserved)
	if err != nil {
		return models.BudgetStatus{}, fmt.Errorf("failed to sum reserved cost: %w", err)
	}

	status := models.BudgetStatus{
		BudgetID:     budget.ID,
		Name:         budget.Name,
		Period:       budget.Period,
		HardLimitUSD: budget.HardLimitUSD,
		SpentUSD:     spent,
		ReservedUSD:  reserved,
	}
	status.Exceeded = spent+reserved >= budget.HardLimitUSD
	if budget.WarningLimitUSD != nil {
		status.Warning = spent+reserved >= *budget.WarningLimitUSD
	}
	return status, nil
}

// querier är det som budgetStatus behöver - antingen en pool eller en
// transaktion.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
