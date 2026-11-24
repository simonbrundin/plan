package usage

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TokenUsage represents token usage for a single request
type TokenUsage struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversationId"`
	Provider       string    `json:"provider"`
	Model          string    `json:"model"`
	InputTokens    int       `json:"inputTokens"`
	OutputTokens   int       `json:"outputTokens"`
	TotalTokens    int       `json:"totalTokens"`
	Cost           float64   `json:"cost"` // Estimated cost in USD
	LatencyMs      int       `json:"latencyMs"`
	Timestamp      time.Time `json:"timestamp"`
}

// CostConfig defines cost per 1M tokens for each provider/model
type CostConfig struct {
	Provider        string
	Model           string
	InputCostPer1M  float64 // Cost per 1M input tokens
	OutputCostPer1M float64 // Cost per 1M output tokens
}

// DefaultCosts returns the default cost configuration
func DefaultCosts() []CostConfig {
	return []CostConfig{
		// MiniMax (approximate, check actual pricing)
		{"minimax", "MiniMax-M2.7", 0.20, 0.40},

		// OpenAI
		{"openai", "gpt-4o-mini", 0.15, 0.60},
		{"openai", "gpt-4o", 2.50, 10.00},
		{"openai", "gpt-4-turbo", 10.00, 30.00},
		{"openai", "gpt-3.5-turbo", 0.50, 1.50},

		// Anthropic
		{"anthropic", "claude-sonnet-4-7-25", 3.00, 15.00},
		{"anthropic", "claude-opus-4-5", 15.00, 75.00},
		{"anthropic", "claude-haiku-4-7", 0.80, 4.00},

		// AgentGateway
		{"agentgateway", "gpt-4o-mini", 0.10, 0.40},
	}
}

// CalculateCost calculates the cost based on tokens and provider
func CalculateCost(provider, model string, inputTokens, outputTokens int, costs []CostConfig) float64 {
	// Find matching cost config
	for _, c := range costs {
		if c.Provider == provider && (c.Model == model || c.Model == "*") {
			inputCost := float64(inputTokens) / 1_000_000 * c.InputCostPer1M
			outputCost := float64(outputTokens) / 1_000_000 * c.OutputCostPer1M
			return inputCost + outputCost
		}
	}

	// Default fallback: assume cheap
	return float64(inputTokens+outputTokens) / 1_000_000 * 0.50
}

// Tracker tracks token usage
type Tracker struct {
	mu    sync.RWMutex
	usage []TokenUsage
	pool  *pgxpool.Pool
	costs []CostConfig
}

// NewTracker creates a new usage tracker
func NewTracker(pool *pgxpool.Pool) *Tracker {
	return &Tracker{
		pool:  pool,
		costs: DefaultCosts(),
	}
}

// Record records a token usage event
func (t *Tracker) Record(ctx context.Context, usage TokenUsage) error {
	// Calculate cost
	usage.Cost = CalculateCost(usage.Provider, usage.Model, usage.InputTokens, usage.OutputTokens, t.costs)
	usage.ID = uuid.New().String()
	usage.Timestamp = time.Now()

	t.mu.Lock()
	t.usage = append(t.usage, usage)
	t.mu.Unlock()

	// Persist to database
	if t.pool != nil {
		return t.persist(ctx, usage)
	}

	return nil
}

// persist saves usage to database
func (t *Tracker) persist(ctx context.Context, usage TokenUsage) error {
	query := `
		INSERT INTO token_usage (
			id, conversation_id, provider, model,
			input_tokens, output_tokens, total_tokens,
			cost, latency_ms, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := t.pool.Exec(ctx, query,
		usage.ID,
		usage.ConversationID,
		usage.Provider,
		usage.Model,
		usage.InputTokens,
		usage.OutputTokens,
		usage.TotalTokens,
		usage.Cost,
		usage.LatencyMs,
		usage.Timestamp,
	)

	return err
}

// GetByConversation returns usage for a conversation
func (t *Tracker) GetByConversation(ctx context.Context, conversationID string) ([]TokenUsage, error) {
	if t.pool == nil {
		t.mu.RLock()
		defer t.mu.RUnlock()

		var result []TokenUsage
		for _, u := range t.usage {
			if u.ConversationID == conversationID {
				result = append(result, u)
			}
		}
		return result, nil
	}

	query := `
		SELECT id, conversation_id, provider, model,
			   input_tokens, output_tokens, total_tokens,
			   cost, latency_ms, timestamp
		FROM token_usage
		WHERE conversation_id = $1
		ORDER BY timestamp DESC
	`

	rows, err := t.pool.Query(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TokenUsage
	for rows.Next() {
		var u TokenUsage
		err := rows.Scan(
			&u.ID, &u.ConversationID, &u.Provider, &u.Model,
			&u.InputTokens, &u.OutputTokens, &u.TotalTokens,
			&u.Cost, &u.LatencyMs, &u.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}

// GetStats returns aggregate statistics
func (t *Tracker) GetStats(ctx context.Context) (*Stats, error) {
	stats := &Stats{}

	if t.pool == nil {
		t.mu.RLock()
		defer t.mu.RUnlock()

		for _, u := range t.usage {
			stats.TotalRequests++
			stats.TotalInputTokens += int64(u.InputTokens)
			stats.TotalOutputTokens += int64(u.OutputTokens)
			stats.TotalCost += u.Cost
			stats.TotalLatencyMs += int64(u.LatencyMs)

			if u.Timestamp.After(stats.FirstRequest) || stats.FirstRequest.IsZero() {
				stats.FirstRequest = u.Timestamp
			}
			if u.Timestamp.Before(stats.LastRequest) || stats.LastRequest.IsZero() {
				stats.LastRequest = u.Timestamp
			}
		}
	} else {
		query := `
			SELECT
				COUNT(*) as total_requests,
				COALESCE(SUM(input_tokens), 0) as total_input_tokens,
				COALESCE(SUM(output_tokens), 0) as total_output_tokens,
				COALESCE(SUM(cost), 0) as total_cost,
				COALESCE(SUM(latency_ms), 0) as total_latency_ms,
				MIN(timestamp) as first_request,
				MAX(timestamp) as last_request
			FROM token_usage
		`

		row := t.pool.QueryRow(ctx, query)
		err := row.Scan(
			&stats.TotalRequests,
			&stats.TotalInputTokens,
			&stats.TotalOutputTokens,
			&stats.TotalCost,
			&stats.TotalLatencyMs,
			&stats.FirstRequest,
			&stats.LastRequest,
		)
		if err != nil {
			return nil, err
		}
	}

	if stats.TotalRequests > 0 {
		stats.AverageLatencyMs = stats.TotalLatencyMs / int64(stats.TotalRequests)
	}

	return stats, nil
}

// Stats represents aggregate usage statistics
type Stats struct {
	TotalRequests     int64     `json:"totalRequests"`
	TotalInputTokens  int64     `json:"totalInputTokens"`
	TotalOutputTokens int64     `json:"totalOutputTokens"`
	TotalCost         float64   `json:"totalCost"`
	TotalLatencyMs    int64     `json:"totalLatencyMs"`
	AverageLatencyMs  int64     `json:"averageLatencyMs"`
	FirstRequest      time.Time `json:"firstRequest"`
	LastRequest       time.Time `json:"lastRequest"`
}

// InMemoryTracker creates a tracker that only stores in memory (for testing)
func InMemoryTracker() *Tracker {
	return &Tracker{
		usage: make([]TokenUsage, 0),
		costs: DefaultCosts(),
	}
}
