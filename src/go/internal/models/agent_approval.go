package models

import (
	"errors"
	"fmt"
	"time"
)

// Approvalstatusar för ett human gate.
const (
	// AgentApprovalPending - beslutet väntar på en människa.
	AgentApprovalPending = "pending"
	// AgentApprovalApproved - en människa godkände.
	AgentApprovalApproved = "approved"
	// AgentApprovalRejected - en människa avvisade.
	AgentApprovalRejected = "rejected"
)

// ErrInvalidApproval betyder att ett godkännande inte kunde användas.
var ErrInvalidApproval = errors.New("invalid approval")

// AgentApproval är ett mänskligt beslut i en agentkörning.
//
// Plan äger godkännandena. Agenten får föreslå, men bara en människa kan
// godkänna - därför är beslutet en egen entitet med sin aktör.
type AgentApproval struct {
	ID          int64  `json:"id" example:"1"`
	AgentRunID  int64  `json:"agent_run_id" example:"12"`
	GoalID      int64  `json:"goal_id" example:"449"`
	Title       string `json:"title" example:"Förslag: lägg till sak i fil"`
	Description string `json:"description" example:"Varför ändringen behövs"`
	// Status är pending, approved eller rejected.
	Status string `json:"status" example:"pending"`
	// DecidedBy är användaren som fattade beslutet.
	DecidedBy *int64 `json:"decided_by"`
	// Feedback är människans motivering.
	Feedback string `json:"feedback,omitempty"`
	// Consequence beskriver vad som händer om beslutet tas.
	Consequence string `json:"consequence,omitempty"`
	// VerificationPlan beskriver hur resultatet ska kontrolleras.
	VerificationPlan string `json:"verification_plan,omitempty"`
	// Payload innehåller exakt det som föreslogs - de ändringar och det
	// testkommando som ska köras. Utan det skulle agenten behöva generera
	// om förslaget, och det som godkänns skulle inte vara det som appliceras.
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	DecidedAt *time.Time     `json:"decided_at"`
}

// CreateAgentApprovalRequest skapar ett godkännande som väntar på beslut.
//
// AgentRunID sätts från URL:en - det ska inte gå att skicka en
// godkännande till en annan körning än den man står i.
type CreateAgentApprovalRequest struct {
	AgentRunID int64  `json:"-"`
	Title      string `json:"title" binding:"required"`
	// Description förklarar varför förslaget finns.
	Description string `json:"description"`
	// Consequence beskriver vad som händer om beslutet tas.
	Consequence string `json:"consequence"`
	// VerificationPlan beskriver hur resultatet ska kontrolleras.
	VerificationPlan string `json:"verification_plan"`
	// Payload innehåller exakt det som föreslogs.
	Payload map[string]any `json:"payload,omitempty"`
}

// DecideAgentApprovalRequest fattar beslutet om ett godkännande.
type DecideAgentApprovalRequest struct {
	// Decision är "approve" eller "reject".
	Decision string `json:"decision" binding:"required"`
	// Feedback är människans motivering.
	Feedback string `json:"feedback,omitempty"`
}

// ValidAgentApprovalDecision kontrollerar att beslutet är känt.
func ValidAgentApprovalDecision(decision string) bool {
	switch decision {
	case "approve", "reject":
		return true
	default:
		return false
	}
}

// ValidateAgentApprovalDecision returnerar ett fel om beslutet är okänt.
func ValidateAgentApprovalDecision(decision string) error {
	if !ValidAgentApprovalDecision(decision) {
		return fmt.Errorf("%w: okänt beslut %q", ErrInvalidApproval, decision)
	}
	return nil
}
