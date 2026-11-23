package models

import (
	"errors"
	"fmt"
	"time"
)

// Agentkörningens faser. Faserna följer agentens livscykel: målet
// specificeras, planeras, utförs, verifieras och är sedan klart.
const (
	AgentRunPhaseDefine   = "define"
	AgentRunPhasePlan     = "plan"
	AgentRunPhaseExecute  = "execute"
	AgentRunPhaseVerify   = "verify"
	AgentRunPhaseComplete = "completed"
)

// Agentkörningens tillstånd. "blocked" betyder att agenten väntar på
// människan - till exempel ett godkännande.
const (
	AgentRunStatusQueued  = "queued"
	AgentRunStatusRunning = "running"
	AgentRunStatusBlocked = "blocked"
	AgentRunStatusDone    = "done"
	AgentRunStatusFailed  = "failed"
)

// ErrInvalidAgentRun betyder att en agentkörning inte kunde användas.
var ErrInvalidAgentRun = errors.New("invalid agent run")

// AgentRun är en agentkörning kopplad till ett mål.
//
// Plan äger körningarna, så gränssnittet visar agentens arbete även när
// agent-tjänsten är nere. Flera körningar kan finnas för samma mål - en
// körning per försök.
type AgentRun struct {
	ID      int64  `json:"id" example:"1"`
	GoalID  int64  `json:"goal_id" example:"449"`
	Phase   string `json:"phase" example:"define"`
	Status  string `json:"status" example:"running"`
	Summary string `json:"summary" example:"Agenten specificerar målet"`
	// ArtifactURL pekar på resultatet - en PR-länk eller ett bokningsutkast.
	ArtifactURL *string `json:"artifact_url"`
	// CostUSD är kostnaden för körningen hittills.
	CostUSD *float64 `json:"cost_usd"`
	// ConversationID knytar körningen till en konversation i agent-tjänsten.
	ConversationID *string   `json:"conversation_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateAgentRunRequest skapar en ny agentkörning.
type CreateAgentRunRequest struct {
	GoalID         int64   `json:"goal_id" binding:"required"`
	Phase          string  `json:"phase"`
	Status         string  `json:"status"`
	Summary        string  `json:"summary"`
	ArtifactURL    *string `json:"artifact_url"`
	ConversationID *string `json:"conversation_id"`
}

// UpdateAgentRunRequest uppdaterar en agentkörning.
//
// Alla fält är valfria - nil lämnar fältet oförändrat.
type UpdateAgentRunRequest struct {
	Phase          *string  `json:"phase"`
	Status         *string  `json:"status"`
	Summary        *string  `json:"summary"`
	ArtifactURL    *string  `json:"artifact_url"`
	CostUSD        *float64 `json:"cost_usd"`
	ConversationID *string  `json:"conversation_id"`
}

// ValidAgentRunPhase kontrollerar att en fas är en känd agentfas.
func ValidAgentRunPhase(phase string) bool {
	switch phase {
	case AgentRunPhaseDefine, AgentRunPhasePlan, AgentRunPhaseExecute,
		AgentRunPhaseVerify, AgentRunPhaseComplete:
		return true
	}
	return false
}

// ValidAgentRunStatus kontrollerar att ett tillstånd är ett känt tillstånd.
func ValidAgentRunStatus(status string) bool {
	switch status {
	case AgentRunStatusQueued, AgentRunStatusRunning, AgentRunStatusBlocked,
		AgentRunStatusDone, AgentRunStatusFailed:
		return true
	}
	return false
}

// ValidateAgentRunPhase returnerar ett fel om fasen är okänd.
func ValidateAgentRunPhase(phase string) error {
	if !ValidAgentRunPhase(phase) {
		return fmt.Errorf("%w: okänd fas %q", ErrInvalidAgentRun, phase)
	}
	return nil
}

// ValidateAgentRunStatus returnerar ett fel om tillståndet är okänt.
func ValidateAgentRunStatus(status string) error {
	if !ValidAgentRunStatus(status) {
		return fmt.Errorf("%w: okänt tillstånd %q", ErrInvalidAgentRun, status)
	}
	return nil
}

// GoalFilter filtrerar en mållista.
//
// Nil-fält betyder "ingen filtrering" - till skillnad från false, som
// filtrerar bort mål där agent-flaggan inte är satt.
type GoalFilter struct {
	// Agent filtrerar på agent-flaggan.
	Agent *bool
	// StatusID filtrerar på målets status.
	StatusID *int
}
