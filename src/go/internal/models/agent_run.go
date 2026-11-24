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
	AgentRunStatusQueued    = "queued"
	AgentRunStatusRunning   = "running"
	AgentRunStatusBlocked   = "blocked"
	AgentRunStatusDone      = "done"
	AgentRunStatusFailed    = "failed"
	AgentRunStatusCancelled = "cancelled"
)

// Agentprofiler - vilken typ av agent som utförde körningen.
const (
	AgentProfileCoding = "coding"
	AgentProfileLife   = "life"
)

// ErrInvalidAgentRun betyder att en agentkörning inte kunde användas.
var ErrInvalidAgentRun = errors.New("invalid agent run")

// AgentRun är en agentkörning kopplad till ett mål.
//
// Plan äger körningarna, så gränssnittet visar agentens arbete även när
// agent-tjänsten är nere. Flera körningar kan finnas för samma mål - en
// körning per försök.
type AgentRun struct {
	ID     int64  `json:"id" example:"1"`
	GoalID int64  `json:"goal_id" example:"449"`
	Phase  string `json:"phase" example:"define"`
	Status string `json:"status" example:"running"`
	// Trigger beskriver vad som startade körningen: user, schedule, webhook eller agent.
	Trigger   string `json:"trigger" example:"user"`
	Initiator string `json:"initiator,omitempty" example:"user:1"`
	// AgentProfile är vilken sorts agent som körde - coding eller life.
	AgentProfile string `json:"agent_profile" example:"coding"`
	Summary      string `json:"summary" example:"Agenten specificerar målet"`
	// Result beskriver vad körningen ledde till.
	Result string `json:"result,omitempty" example:"Målet definierat och flyttat till plan"`
	// ArtifactURL pekar på resultatet - en PR-länk eller ett bokningsutkast.
	ArtifactURL *string `json:"artifact_url"`
	// CostUSD är kostnaden för körningen hittills.
	CostUSD *float64 `json:"cost_usd"`
	// ConversationID knytar körningen till en konversation i agent-tjänsten.
	ConversationID *string `json:"conversation_id"`
	// StartedAt sätts när körningen börjar, FinishedAt när den är klar.
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// CreateAgentRunRequest skapar en ny agentkörning.
type CreateAgentRunRequest struct {
	GoalID         int64   `json:"goal_id" binding:"required"`
	Phase          string  `json:"phase"`
	Status         string  `json:"status"`
	Trigger        string  `json:"trigger"`
	Initiator      string  `json:"initiator"`
	AgentProfile   string  `json:"agent_profile"`
	Summary        string  `json:"summary"`
	Result         string  `json:"result"`
	ArtifactURL    *string `json:"artifact_url"`
	ConversationID *string `json:"conversation_id"`
	// StartedAt sätts när körningen börjar. Nil ger "nu".
	StartedAt *time.Time `json:"started_at"`
}

// UpdateAgentRunRequest uppdaterar en agentkörning.
//
// Alla fält är valfria - nil lämnar fältet oförändrat.
type UpdateAgentRunRequest struct {
	Phase          *string    `json:"phase"`
	Status         *string    `json:"status"`
	Trigger        *string    `json:"trigger"`
	Initiator      *string    `json:"initiator"`
	AgentProfile   *string    `json:"agent_profile"`
	Summary        *string    `json:"summary"`
	Result         *string    `json:"result"`
	ArtifactURL    *string    `json:"artifact_url"`
	CostUSD        *float64   `json:"cost_usd"`
	ConversationID *string    `json:"conversation_id"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
}

// ValidAgentRunTrigger kontrollerar hur en körning startades.
func ValidAgentRunTrigger(trigger string) bool {
	switch trigger {
	case "user", "schedule", "webhook", "agent":
		return true
	default:
		return false
	}
}

// ValidateAgentRunTrigger returnerar ett fel om triggern är okänd.
func ValidateAgentRunTrigger(trigger string) error {
	if !ValidAgentRunTrigger(trigger) {
		return fmt.Errorf("%w: okänd trigger %q", ErrInvalidAgentRun, trigger)
	}
	return nil
}

// ValidAgentProfile kontrollerar vilka agentprofiler som finns.
func ValidAgentProfile(profile string) bool {
	switch profile {
	case AgentProfileCoding, AgentProfileLife:
		return true
	default:
		return false
	}
}

// ValidateAgentProfile returnerar ett fel om profilen är okänd.
func ValidateAgentProfile(profile string) error {
	if !ValidAgentProfile(profile) {
		return fmt.Errorf("%w: okänd agentprofil %q", ErrInvalidAgentRun, profile)
	}
	return nil
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
		AgentRunStatusDone, AgentRunStatusFailed, AgentRunStatusCancelled:
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

// IsTerminalAgentRunStatus berättar om en körning är slut - då ska den
// inte kunna återupplivas och den får en sluttid.
func IsTerminalAgentRunStatus(status string) bool {
	return status == AgentRunStatusDone || status == AgentRunStatusFailed ||
		status == AgentRunStatusCancelled
}

// ValidateAgentRunTransition säkerställer att en körning går framåt i sin
// livscykel och inte återupplivas efter ett terminalt tillstånd.
func ValidateAgentRunTransition(current AgentRun, requested UpdateAgentRunRequest) error {
	if requested.Phase != nil {
		if err := ValidateAgentRunPhase(*requested.Phase); err != nil {
			return err
		}
		if agentRunPhaseRank(*requested.Phase) < agentRunPhaseRank(current.Phase) {
			return fmt.Errorf("%w: fasen kan inte gå bakåt från %q till %q", ErrInvalidAgentRun, current.Phase, *requested.Phase)
		}
	}

	if requested.Status != nil && !validAgentRunStatusTransition(current.Status, *requested.Status) {
		return fmt.Errorf("%w: tillståndet kan inte gå från %q till %q", ErrInvalidAgentRun, current.Status, *requested.Status)
	}
	return nil
}

func agentRunPhaseRank(phase string) int {
	switch phase {
	case AgentRunPhaseDefine:
		return 0
	case AgentRunPhasePlan:
		return 1
	case AgentRunPhaseExecute:
		return 2
	case AgentRunPhaseVerify:
		return 3
	case AgentRunPhaseComplete:
		return 4
	default:
		return -1
	}
}

func validAgentRunStatusTransition(current, requested string) bool {
	if current == requested {
		return true
	}
	if IsTerminalAgentRunStatus(current) {
		return false
	}

	// En körning kan alltid avbrytas - av en människa, av en budget eller
	// av en policy. Det är ett sätt att stanna, inte ett sätt att gå vidare.
	if requested == AgentRunStatusCancelled {
		return true
	}

	switch current {
	case AgentRunStatusQueued:
		return requested == AgentRunStatusRunning || requested == AgentRunStatusBlocked || requested == AgentRunStatusDone || requested == AgentRunStatusFailed
	case AgentRunStatusRunning:
		return requested == AgentRunStatusBlocked || requested == AgentRunStatusDone || requested == AgentRunStatusFailed
	case AgentRunStatusBlocked:
		return requested == AgentRunStatusQueued || requested == AgentRunStatusRunning || requested == AgentRunStatusDone || requested == AgentRunStatusFailed
	default:
		return false
	}
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
