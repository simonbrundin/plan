package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"plan-api/internal/pricing"
)

// Statusar för en användningsrad.
const (
	// AgentUsageStatusReported betyder att agenten har rapporterat
	// användningen.
	AgentUsageStatusReported = "reported"
	// AgentUsageStatusIncomplete betyder att körningen avslutades utan att
	// agenten hann rapportera - till exempel för att jobbet avbröts.
	AgentUsageStatusIncomplete = "incomplete"
)

// Budgetperioder.
const (
	// BudgetPeriodDay täcker ett kalenderdygn.
	BudgetPeriodDay = "day"
	// BudgetPeriodWeek täcker sju dagar.
	BudgetPeriodWeek = "week"
	// BudgetPeriodMonth täcker en kalendermånad.
	BudgetPeriodMonth = "month"
)

// ErrBudgetExceeded betyder att en budget är slut och att en ny körning
// inte får starta.
var ErrBudgetExceeded = errors.New("budget exceeded")

// AgentRunUsage är kostnaden för en agentkörning.
//
// Raden sparas när körningen avslutas. Den innehåller allt som behövs
// för att reproducera kostnaden: användningen, modellen, och vilken
// prislista som räknades fram den.
type AgentRunUsage struct {
	ID         int64 `json:"id" example:"1"`
	AgentRunID int64 `json:"agent_run_id" example:"12"`
	GoalID     int64 `json:"goal_id" example:"42"`
	// AgentID identifierar agenten som gjorde körningen.
	AgentID string `json:"agent_id" example:"agent-worker-run-72"`
	// Provider och Model visar vilken modell som användes.
	Provider string `json:"provider" example:"minimax"`
	Model    string `json:"model" example:"MiniMax-M2.7"`
	// ModelVersion är modellens version.
	ModelVersion string `json:"model_version,omitempty"`
	// Tokenantal för de olika sorterna.
	InputTokens     int `json:"input_tokens" example:"1200"`
	OutputTokens    int `json:"output_tokens" example:"800"`
	CacheReadTokens int `json:"cache_read_tokens" example:"0"`
	// ToolCalls är antalet verktygsanrop och ToolNames de anropade
	// verktygens namn.
	ToolCalls int      `json:"tool_calls" example:"4"`
	ToolNames []string `json:"tool_names,omitempty"`
	// CostUSD är den beräknade kostnaden.
	CostUSD float64 `json:"cost_usd" example:"0.0042"`
	// Currency är valutan kostnaden anges i.
	Currency string `json:"currency" example:"USD"`
	// PriceListVersion visar vilken prislista som användes.
	PriceListVersion string `json:"price_list_version" example:"2026-11-24.1"`
	// PriceSource visar om priset kom från prislistan eller från en
	// schablon.
	PriceSource string `json:"price_source" example:"price_list"`
	// Status är "reported" eller "incomplete".
	Status string `json:"status" example:"reported"`
	// StartedAt och FinishedAt avgränsar körningen.
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreateAgentRunUsageRequest är det agenten rapporterar för en körning.
type CreateAgentRunUsageRequest struct {
	AgentID string `json:"agent_id"`
	// Provider och Model används för att hitta priset i prislistan.
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	ModelVersion string `json:"model_version,omitempty"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	// CacheReadTokens är tokens som lästs från modellens cache.
	CacheReadTokens int        `json:"cache_read_tokens"`
	ToolCalls       int        `json:"tool_calls"`
	ToolNames       []string   `json:"tool_names,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

// AgentBudget är en gräns för agentkostnad under en period.
//
// En budget gäller antingen en viss agent eller alla agenter. AgentID är
// tom när budgeten gäller alla.
type AgentBudget struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"Mars"`
	// Period är day, week eller month.
	Period string `json:"period" example:"month"`
	// HardLimitUSD stoppar nya körningar när gränsen nås.
	HardLimitUSD float64 `json:"hard_limit_usd" example:"10"`
	// WarningLimitUSD varnar utan att blockera.
	WarningLimitUSD *float64 `json:"warning_limit_usd,omitempty"`
	// ReservationPerRunUSD är hur mycket en ny körning reserverar direkt.
	// Det gör att samtidiga körningar inte tillsammans överskrider
	// gränsen.
	ReservationPerRunUSD float64 `json:"reservation_per_run_usd" example:"1"`
	// AgentID gäller en enskild agent. Tom gäller alla.
	AgentID string `json:"agent_id,omitempty"`
	// Active anger om budgeten gäller.
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateAgentBudgetRequest skapar en budget.
type CreateAgentBudgetRequest struct {
	Name                 string   `json:"name"`
	Period               string   `json:"period"`
	HardLimitUSD         float64  `json:"hard_limit_usd"`
	WarningLimitUSD      *float64 `json:"warning_limit_usd,omitempty"`
	ReservationPerRunUSD float64  `json:"reservation_per_run_usd,omitempty"`
	AgentID              string   `json:"agent_id,omitempty"`
}

// UpdateAgentBudgetRequest ändrar en befintlig budget.
//
// Varje fält är valfritt - det som inte skickas lämnas som det är.
type UpdateAgentBudgetRequest struct {
	Name                 *string  `json:"name,omitempty"`
	Period               *string  `json:"period,omitempty"`
	HardLimitUSD         *float64 `json:"hard_limit_usd,omitempty"`
	WarningLimitUSD      *float64 `json:"warning_limit_usd,omitempty"`
	ReservationPerRunUSD *float64 `json:"reservation_per_run_usd,omitempty"`
	Active               *bool    `json:"active,omitempty"`
}

// ValidateAgentBudgetPeriod kontrollerar att perioden är känd.
func ValidateAgentBudgetPeriod(period string) error {
	switch period {
	case BudgetPeriodDay, BudgetPeriodWeek, BudgetPeriodMonth:
		return nil
	default:
		return fmt.Errorf("ogiltig budgetperiod %q", period)
	}
}

// BudgetStatus beskriver en budgets läge just nu.
//
// Det är detta som visas när en körning stoppas, så att en människa kan
// se varför den inte startade.
type BudgetStatus struct {
	BudgetID int64  `json:"budget_id" example:"1"`
	Name     string `json:"name" example:"Mars"`
	Period   string `json:"period" example:"month"`
	// HardLimitUSD, SpentUSD och ReservedUSD beskriver gränsen och vad
	// som redan är använd eller reserverat.
	HardLimitUSD float64 `json:"hard_limit_usd" example:"10"`
	SpentUSD     float64 `json:"spent_usd" example:"8.5"`
	ReservedUSD  float64 `json:"reserved_usd" example:"1"`
	// Exceeded är true om gränsen är nådd.
	Exceeded bool `json:"exceeded"`
	// Warning är true om varningsgränsen är nådd.
	Warning bool `json:"warning"`
	// Reason förklarar varför körningen stoppades: "limit_reached" eller
	// "unknown_price".
	Reason string `json:"reason,omitempty" example:"limit_reached"`
}

// Currency är valutan budgeten räknas i.
func (s BudgetStatus) Currency() string {
	return pricing.Currency
}

// BudgetExceededError säger att en budget stoppade en körning.
type BudgetExceededError struct {
	Status BudgetStatus
}

func (e *BudgetExceededError) Error() string {
	// Skälet styr budskapet - en okänd prisbild är inte samma sak som
	// en budget som är slut.
	if e.Status.Reason == "unknown_price" {
		return fmt.Sprintf("budget %q kräver ett känt modellpris, men priset för modellen är okänt",
			e.Status.Name)
	}
	return fmt.Sprintf("budget %q är slut: %.4f %s använd eller reserverad, gräns %.4f %s",
		e.Status.Name, e.Status.SpentUSD+e.Status.ReservedUSD, e.Status.Currency(),
		e.Status.HardLimitUSD, e.Status.Currency())
}

// AgentUsageFilter filtrerar kostnadsdata.
type AgentUsageFilter struct {
	// AgentID gäller en enskild agent.
	AgentID string
	// Model gäller en enskild modell.
	Model string
	// Tool gäller körningar som använt ett verktyg.
	Tool string
	// From och To avgränsar perioden.
	From *time.Time
	To   *time.Time
}

// AgentUsageSummary är en aggregerad kostnad.
type AgentUsageSummary struct {
	// Runs är antalet körningar i urvalet.
	Runs int `json:"runs" example:"4"`
	// TotalCostUSD är summan av körningarnas kostnad.
	TotalCostUSD float64 `json:"total_cost_usd" example:"0.42"`
	// InputTokens, OutputTokens och CacheReadTokens är summan av tokens.
	InputTokens     int `json:"input_tokens" example:"12000"`
	OutputTokens    int `json:"output_tokens" example:"8000"`
	CacheReadTokens int `json:"cache_read_tokens" example:"0"`
	// ToolCalls är summan av verktygsanrop.
	ToolCalls int `json:"tool_calls" example:"17"`
	// Currency är valutan kostnaderna anges i.
	Currency string `json:"currency" example:"USD"`
	// ByModel bryter ner kostnaden per modell.
	ByModel []AgentUsageByModel `json:"by_model,omitempty"`
}

// AgentUsageByModel är kostnaden för en modell.
type AgentUsageByModel struct {
	Provider string  `json:"provider" example:"minimax"`
	Model    string  `json:"model" example:"MiniMax-M2.7"`
	Runs     int     `json:"runs" example:"2"`
	CostUSD  float64 `json:"cost_usd" example:"0.12"`
}

// NormalizeAgentID tar bort mellanslag runt ett agent-id.
func NormalizeAgentID(agentID string) string {
	return strings.TrimSpace(agentID)
}
