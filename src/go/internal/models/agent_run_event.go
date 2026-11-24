package models

import (
	"errors"
	"fmt"
	"time"
)

// Händelsetyper i en AgentRun.
//
// Typerna är ett slutet set - det gör att historiken blir läsbar och att
// en trasig agent inte kan skriva godtycklig text i historiken.
const (
	// AgentRunEventRunStarted - körningen har startat.
	AgentRunEventRunStarted = "run_started"
	// AgentRunEventDefinitionCreated - agenten har definierat målet.
	AgentRunEventDefinitionCreated = "definition_created"
	// AgentRunEventPhaseChanged - körningen har bytt fas.
	AgentRunEventPhaseChanged = "phase_changed"
	// AgentRunEventToolCall - ett verktyg anropades.
	AgentRunEventToolCall = "tool_call"
	// AgentRunEventToolResult - ett verktyg gav ett resultat.
	AgentRunEventToolResult = "tool_result"
	// AgentRunEventSubgoalCreated - agenten skapade ett undermål.
	AgentRunEventSubgoalCreated = "subgoal_created"
	// AgentRunEventApprovalRequested - agenten begär ett mänskligt beslut.
	AgentRunEventApprovalRequested = "approval_requested"
	// AgentRunEventApprovalDecided - ett mänskligt beslut fattades.
	AgentRunEventApprovalDecided = "approval_decided"
	// AgentRunEventVerification - en verifiering utfördes.
	AgentRunEventVerification = "verification"
	// AgentRunEventArtifact - en artefakt skapades, till exempel en diff.
	AgentRunEventArtifact = "artifact"
	// AgentRunEventError - något gick fel.
	AgentRunEventError = "error"
	// AgentRunEventRunFailed - körningen misslyckades.
	AgentRunEventRunFailed = "run_failed"
	// AgentRunEventRunCompleted - körningen var klar.
	AgentRunEventRunCompleted = "run_completed"
	// AgentRunEventVerificationCheckpoint - en manuell kontrollpunkt.
	AgentRunEventVerificationCheckpoint = "verification_checkpoint"
)

// ErrInvalidAgentRunEvent betyder att en händelse inte kunde användas.
var ErrInvalidAgentRunEvent = errors.New("invalid agent run event")

// AgentRunEvent är en spårbar händelse i en AgentRun.
type AgentRunEvent struct {
	ID         int64          `json:"id" example:"1"`
	AgentRunID int64          `json:"agent_run_id" example:"12"`
	Type       string         `json:"type" example:"tool_call"`
	Phase      string         `json:"phase,omitempty" example:"execute"`
	Payload    map[string]any `json:"payload,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// CreateAgentRunEventRequest skapar en händelse för en körning.
type CreateAgentRunEventRequest struct {
	Type    string         `json:"type" binding:"required"`
	Phase   string         `json:"phase,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

// ValidAgentRunEventType kontrollerar att händelsetypen är en känd typ.
func ValidAgentRunEventType(eventType string) bool {
	switch eventType {
	case AgentRunEventRunStarted, AgentRunEventDefinitionCreated,
		AgentRunEventPhaseChanged, AgentRunEventToolCall, AgentRunEventToolResult,
		AgentRunEventSubgoalCreated, AgentRunEventApprovalRequested,
		AgentRunEventApprovalDecided, AgentRunEventVerification,
		AgentRunEventArtifact, AgentRunEventError, AgentRunEventRunFailed,
		AgentRunEventRunCompleted, AgentRunEventVerificationCheckpoint:
		return true
	default:
		return false
	}
}

// ValidateAgentRunEventType returnerar ett fel om händelsetypen är okänd.
func ValidateAgentRunEventType(eventType string) error {
	if !ValidAgentRunEventType(eventType) {
		return fmt.Errorf("%w: okänd händelsetyp %q", ErrInvalidAgentRunEvent, eventType)
	}
	return nil
}
