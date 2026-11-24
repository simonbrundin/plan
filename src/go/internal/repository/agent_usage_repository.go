package repository

import (
	"context"
	"fmt"
	"time"

	"plan-api/internal/database"
	"plan-api/internal/models"
	"plan-api/internal/pricing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AgentUsageRepository sparar och samlar användning och kostnad för
// agentkörningar.
//
// Kostnaden räknas fram i Plan, inte i agenten. Då sparas samma prislista
// för alla körningar, och kostnaden kan reproduceras senare.
type AgentUsageRepository struct {
	pool *pgxpool.Pool
	// prices är prislistan som används för att räkna fram kostnader.
	prices pricing.PriceList
}

// NewAgentUsageRepository skapar ett repository för användning.
func NewAgentUsageRepository() *AgentUsageRepository {
	return &AgentUsageRepository{
		pool:   database.GetPool(),
		prices: pricing.Current(),
	}
}

const agentUsageColumns = `u.id, u.agent_run_id, u.goal_id, u.agent_id, u.provider, u.model, u.model_version,
	u.input_tokens, u.output_tokens, u.cache_read_tokens, u.tool_calls, u.tool_names,
	u.cost_usd, u.currency, u.price_list_version, u.price_source, u.status,
	u.started_at, u.finished_at, u.created_at`

// Record sparar en användningsrad för en körning.
//
// Kostnaden räknas fram ur den aktuella prislistan, och listans version
// sparas med raden. Om modellen saknas i listan används schablonen, och
// källan markeras som "fallback" - kostnaden ska inte läs som exakt.
func (r *AgentUsageRepository) Record(ctx context.Context, userID, runID int64, req models.CreateAgentRunUsageRequest) (*models.AgentRunUsage, error) {
	run, err := r.runInfo(ctx, userID, runID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, nil
	}

	tokens := pricing.Tokens{
		Input:     req.InputTokens,
		Output:    req.OutputTokens,
		CacheRead: req.CacheReadTokens,
	}
	result := r.prices.Cost(run.Provider, run.Model, tokens)

	query := `INSERT INTO agent_run_usage (agent_run_id, goal_id, agent_id, provider, model, model_version,
		input_tokens, output_tokens, cache_read_tokens, tool_calls, tool_names,
		cost_usd, currency, price_list_version, price_source, status, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (agent_run_id) DO UPDATE SET
			agent_id = EXCLUDED.agent_id,
			provider = EXCLUDED.provider,
			model = EXCLUDED.model,
			model_version = EXCLUDED.model_version,
			input_tokens = EXCLUDED.input_tokens,
			output_tokens = EXCLUDED.output_tokens,
			cache_read_tokens = EXCLUDED.cache_read_tokens,
			tool_calls = EXCLUDED.tool_calls,
			tool_names = EXCLUDED.tool_names,
			cost_usd = EXCLUDED.cost_usd,
			currency = EXCLUDED.currency,
			price_list_version = EXCLUDED.price_list_version,
			price_source = EXCLUDED.price_source,
			status = EXCLUDED.status,
			started_at = EXCLUDED.started_at,
			finished_at = EXCLUDED.finished_at
		RETURNING id, agent_run_id, goal_id, agent_id, provider, model, model_version,
			input_tokens, output_tokens, cache_read_tokens, tool_calls, tool_names,
			cost_usd, currency, price_list_version, price_source, status,
			started_at, finished_at, created_at`

	var usage models.AgentRunUsage
	err = r.pool.QueryRow(ctx, query,
		runID, run.GoalID, req.AgentID, run.Provider, run.Model, req.ModelVersion,
		req.InputTokens, req.OutputTokens, req.CacheReadTokens, req.ToolCalls, req.ToolNames,
		result.CostUSD, result.Currency, result.PriceListVersion, result.Source,
		models.AgentUsageStatusReported, req.StartedAt, req.FinishedAt,
	).Scan(&usage.ID, &usage.AgentRunID, &usage.GoalID, &usage.AgentID, &usage.Provider, &usage.Model,
		&usage.ModelVersion, &usage.InputTokens, &usage.OutputTokens, &usage.CacheReadTokens,
		&usage.ToolCalls, &usage.ToolNames, &usage.CostUSD, &usage.Currency, &usage.PriceListVersion,
		&usage.PriceSource, &usage.Status, &usage.StartedAt, &usage.FinishedAt, &usage.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to record agent usage: %w", err)
	}

	// Körningens kostnad speglas i agent_runs, så att listan snabbt kan
	// visa vad en körning kostade.
	if _, err := r.pool.Exec(ctx,
		`UPDATE agent_runs SET cost_usd = $1, updated_at = NOW() WHERE id = $2`,
		result.CostUSD, runID); err != nil {
		return nil, fmt.Errorf("failed to update run cost: %w", err)
	}

	return &usage, nil
}

// EnsureRow skapar en ofullständig rad om körningen saknar en.
//
// En körning som avbryts innan agenten hann rapportera ska ändå ha en
// kostnadsrad, så att det går att se att något saknas i stället för att
// raden tyst försvinner. Raden markeras som "incomplete".
func (r *AgentUsageRepository) EnsureRow(ctx context.Context, userID, runID int64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO agent_run_usage (agent_run_id, goal_id, status, currency, price_list_version, price_source)
		SELECT ar.id, ar.goal_id, $3, $4, $5, $6
		FROM agent_runs ar
		WHERE ar.id = $1
			AND ar.goal_id IN (SELECT goal_id FROM user_goals WHERE user_id = $2)
		ON CONFLICT (agent_run_id) DO NOTHING`,
		runID, userID, models.AgentUsageStatusIncomplete, r.prices.Currency, r.prices.Version, pricing.SourcePriceList)
	if err != nil {
		return fmt.Errorf("failed to ensure usage row: %w", err)
	}
	return nil
}

// GetByRun returnerar användningen för en körning.
func (r *AgentUsageRepository) GetByRun(ctx context.Context, userID, runID int64) (*models.AgentRunUsage, error) {
	query := `SELECT ` + agentUsageColumns + `
		FROM agent_run_usage u
		INNER JOIN agent_runs ar ON ar.id = u.agent_run_id
		WHERE u.agent_run_id = $1
			AND ar.goal_id IN (SELECT goal_id FROM user_goals WHERE user_id = $2)`

	var usage models.AgentRunUsage
	err := r.pool.QueryRow(ctx, query, runID, userID).Scan(&usage.ID, &usage.AgentRunID, &usage.GoalID,
		&usage.AgentID, &usage.Provider, &usage.Model, &usage.ModelVersion,
		&usage.InputTokens, &usage.OutputTokens, &usage.CacheReadTokens, &usage.ToolCalls, &usage.ToolNames,
		&usage.CostUSD, &usage.Currency, &usage.PriceListVersion, &usage.PriceSource, &usage.Status,
		&usage.StartedAt, &usage.FinishedAt, &usage.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent usage: %w", err)
	}
	return &usage, nil
}

// List returnerar användningsrader för användaren, med filter.
func (r *AgentUsageRepository) List(ctx context.Context, userID int64, filter models.AgentUsageFilter) ([]models.AgentRunUsage, error) {
	query := `SELECT ` + agentUsageColumns + `
		FROM agent_run_usage u
		INNER JOIN user_goals ug ON ug.goal_id = u.goal_id
		WHERE ug.user_id = $1`
	args := []interface{}{userID}

	if filter.AgentID != "" {
		args = append(args, filter.AgentID)
		query += fmt.Sprintf(" AND u.agent_id = $%d", len(args))
	}
	if filter.Model != "" {
		args = append(args, filter.Model)
		query += fmt.Sprintf(" AND u.model = $%d", len(args))
	}
	if filter.Tool != "" {
		args = append(args, filter.Tool)
		query += fmt.Sprintf(" AND $%d = ANY(u.tool_names)", len(args))
	}
	if filter.From != nil {
		args = append(args, *filter.From)
		query += fmt.Sprintf(" AND u.created_at >= $%d", len(args))
	}
	if filter.To != nil {
		args = append(args, *filter.To)
		query += fmt.Sprintf(" AND u.created_at <= $%d", len(args))
	}
	query += " ORDER BY u.created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list agent usage: %w", err)
	}
	defer rows.Close()

	usage := []models.AgentRunUsage{}
	for rows.Next() {
		var u models.AgentRunUsage
		if err := rows.Scan(&u.ID, &u.AgentRunID, &u.GoalID, &u.AgentID, &u.Provider, &u.Model,
			&u.ModelVersion, &u.InputTokens, &u.OutputTokens, &u.CacheReadTokens, &u.ToolCalls, &u.ToolNames,
			&u.CostUSD, &u.Currency, &u.PriceListVersion, &u.PriceSource, &u.Status,
			&u.StartedAt, &u.FinishedAt, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan agent usage: %w", err)
		}
		usage = append(usage, u)
	}
	return usage, rows.Err()
}

// Summary summerar användningen för ett filter.
func (r *AgentUsageRepository) Summary(ctx context.Context, userID int64, filter models.AgentUsageFilter) (*models.AgentUsageSummary, error) {
	rows, err := r.List(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	summary := &models.AgentUsageSummary{
		Runs:     len(rows),
		Currency: r.prices.Currency,
	}

	byModel := map[string]*models.AgentUsageByModel{}
	for _, u := range rows {
		summary.TotalCostUSD += u.CostUSD
		summary.InputTokens += u.InputTokens
		summary.OutputTokens += u.OutputTokens
		summary.CacheReadTokens += u.CacheReadTokens
		summary.ToolCalls += u.ToolCalls

		key := u.Provider + "/" + u.Model
		entry, ok := byModel[key]
		if !ok {
			entry = &models.AgentUsageByModel{Provider: u.Provider, Model: u.Model}
			byModel[key] = entry
		}
		entry.Runs++
		entry.CostUSD += u.CostUSD
	}

	for _, entry := range byModel {
		summary.ByModel = append(summary.ByModel, *entry)
	}
	return summary, nil
}

// runInfo hämtar de uppgifter om körningen som behövs för att prissätta.
func (r *AgentUsageRepository) runInfo(ctx context.Context, userID, runID int64) (*models.AgentRun, error) {
	query := `SELECT ar.id, ar.goal_id, ar.phase, ar.status, COALESCE(ar.trigger, 'user'),
		COALESCE(ar.initiator, ''), COALESCE(ar.agent_profile, 'coding'),
		COALESCE(ar.summary, ''), COALESCE(ar.result, ''), ar.artifact_url, ar.cost_usd,
		ar.conversation_id, ar.started_at, ar.finished_at, ar.created_at, ar.updated_at
		FROM agent_runs ar
		WHERE ar.id = $1
			AND ar.goal_id IN (SELECT goal_id FROM user_goals WHERE user_id = $2)`

	var run models.AgentRun
	err := r.pool.QueryRow(ctx, query, runID, userID).Scan(&run.ID, &run.GoalID, &run.Phase, &run.Status,
		&run.Trigger, &run.Initiator, &run.AgentProfile, &run.Summary, &run.Result, &run.ArtifactURL,
		&run.CostUSD, &run.ConversationID, &run.StartedAt, &run.FinishedAt, &run.CreatedAt, &run.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent run: %w", err)
	}
	return &run, nil
}

// periodStart räknar ut periodens början för en budget.
func periodStart(period string, now time.Time) time.Time {
	year, month, day := now.UTC().Date()
	switch period {
	case models.BudgetPeriodDay:
		return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	case models.BudgetPeriodWeek:
		// Veckans början är måndagen.
		offset := (int(now.UTC().Weekday()) + 6) % 7
		start := time.Date(year, month, day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -offset)
		return start
	default:
		return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	}
}
