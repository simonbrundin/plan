package orchestrator

import (
	"fmt"
	"strings"

	"plan-agent/internal/core"
	"plan-agent/internal/core/checkpoint"
)

// ShouldStopHook implements the gate-checking logic for agent stopping
type ShouldStopHook struct {
	orchestrator *Orchestrator
	// Message builder for feedback to agent
	messageBuilder strings.Builder
}

// NewShouldStopHook creates a new ShouldStop hook
func NewShouldStopHook(o *Orchestrator) ShouldStopHook {
	return ShouldStopHook{
		orchestrator: o,
	}
}

// ShouldStop returns true if the agent should stop (all gates passed)
// or false if the agent should continue (gates not passed)
//
// This is the "exit code 2" mechanism from Shopify's workflow:
// - Agent tries to stop
// - Hook intercepts and says "no, keep working"
// - Only when all gates pass does the hook allow stopping
func (h *ShouldStopHook) ShouldStop(turnCtx *core.TurnContext) bool {
	plan := h.orchestrator.CurrentPlan()
	if plan == nil {
		// No active plan, allow stop
		return true
	}

	current := plan.GetCurrent()
	if current == nil {
		// No active checkpoint, allow stop
		return true
	}

	// Check all gates for this checkpoint
	pendingGates := current.GetPendingGates()
	if len(pendingGates) == 0 {
		// All gates passed
		return true
	}

	// Gates not passed - agent must continue
	// Build feedback message
	h.messageBuilder.Reset()
	h.messageBuilder.WriteString(fmt.Sprintf(
		"\n⚠️  Cannot stop - checkpoint '%s' has pending gates:\n",
		current.Name,
	))

	for _, gateType := range pendingGates {
		result := current.GateResults[gateType]
		status := string(checkpoint.GateStatusPending)
		if result != nil {
			status = string(result.Status)
		}
		h.messageBuilder.WriteString(fmt.Sprintf("  • %s: %s\n", gateType, status))
	}

	h.messageBuilder.WriteString("\n📋 Continue working to pass these gates before stopping.\n")

	// In a real implementation, this message would be injected into the agent's context
	// For now, this hook prevents stopping but the message needs to be conveyed separately

	return false
}

// GetBlockedMessage returns the message explaining why stop was blocked
func (h *ShouldStopHook) GetBlockedMessage() string {
	return h.messageBuilder.String()
}

// CheckGateStatus checks if a specific gate has passed for the current checkpoint
func (h *ShouldStopHook) CheckGateStatus(gateType string) (passed bool, reason string) {
	plan := h.orchestrator.CurrentPlan()
	if plan == nil {
		return true, "no active plan"
	}

	current := plan.GetCurrent()
	if current == nil {
		return true, "no active checkpoint"
	}

	result, exists := current.GateResults[gateType]
	if !exists {
		return false, fmt.Sprintf("gate %s not yet executed", gateType)
	}

	switch result.Status {
	case checkpoint.GateStatusPassed:
		return true, "gate passed"
	case checkpoint.GateStatusFailed:
		return false, fmt.Sprintf("gate failed: %s", result.Feedback)
	case checkpoint.GateStatusRunning:
		return false, "gate still running"
	case checkpoint.GateStatusSkipped:
		return true, "gate skipped"
	default:
		return false, fmt.Sprintf("gate status: %s", result.Status)
	}
}

// GetPendingGatesForCurrentCheckpoint returns all pending gates
func (h *ShouldStopHook) GetPendingGatesForCurrentCheckpoint() []string {
	plan := h.orchestrator.CurrentPlan()
	if plan == nil {
		return nil
	}

	current := plan.GetCurrent()
	if current == nil {
		return nil
	}

	return current.GetPendingGates()
}

// IsCheckpointComplete returns true if the current checkpoint can be marked complete
func (h *ShouldStopHook) IsCheckpointComplete() bool {
	plan := h.orchestrator.CurrentPlan()
	if plan == nil {
		return false
	}

	current := plan.GetCurrent()
	if current == nil {
		return false
	}

	return current.AllGatesPassed()
}

// GetCheckpointProgress returns progress info for the current checkpoint
func (h *ShouldStopHook) GetCheckpointProgress() (passed, total int) {
	plan := h.orchestrator.CurrentPlan()
	if plan == nil {
		return 0, 0
	}

	current := plan.GetCurrent()
	if current == nil {
		return 0, 0
	}

	total = len(current.Gates)
	for _, gateType := range current.Gates {
		result, exists := current.GateResults[gateType]
		if exists && result.Status == checkpoint.GateStatusPassed {
			passed++
		}
	}

	return passed, total
}
