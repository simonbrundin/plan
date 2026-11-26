package models

import "testing"

// TestValidAgentRunEventType kontrollerar vilka händelsetyper som accepteras.
func TestValidAgentRunEventType(t *testing.T) {
	valid := []string{
		AgentRunEventRunStarted,
		AgentRunEventDefinitionCreated,
		AgentRunEventPhaseChanged,
		AgentRunEventToolCall,
		AgentRunEventToolResult,
		AgentRunEventSubgoalCreated,
		AgentRunEventApprovalRequested,
		AgentRunEventApprovalDecided,
		AgentRunEventVerification,
		AgentRunEventArtifact,
		AgentRunEventError,
		AgentRunEventRunFailed,
		AgentRunEventRunCompleted,
		AgentRunEventVerificationCheckpoint,
	}
	for _, eventType := range valid {
		if !ValidAgentRunEventType(eventType) {
			t.Errorf("händelsetypen %q ska vara giltig", eventType)
		}
	}

	invalid := []string{"", "happened", "run_started ", "TOOL_CALL"}
	for _, eventType := range invalid {
		if ValidAgentRunEventType(eventType) {
			t.Errorf("händelsetypen %q ska inte vara giltig", eventType)
		}
	}
}

// TestValidateAgentRunEventTypeReturnerarFel kontrollerar att en okänd
// händelsetyp ger ett fel - API:t ska svara 400 på sådana.
func TestValidateAgentRunEventTypeReturnerarFel(t *testing.T) {
	if err := ValidateAgentRunEventType(AgentRunEventToolCall); err != nil {
		t.Errorf("giltig händelsetyp ska inte ge fel: %v", err)
	}
	if err := ValidateAgentRunEventType("happened"); err == nil {
		t.Error("okänd händelsetyp ska ge fel")
	}
}

// TestIsTerminalAgentRunStatus kontrollerar vilka tillstånd som avslutar
// en körning.
func TestIsTerminalAgentRunStatus(t *testing.T) {
	terminal := []string{AgentRunStatusSucceeded, AgentRunStatusFailed}
	for _, status := range terminal {
		if !IsTerminalAgentRunStatus(status) {
			t.Errorf("tillståndet %q ska vara terminalt", status)
		}
	}

	ongoing := []string{AgentRunStatusQueued, AgentRunStatusRunning, AgentRunStatusWaitingApproval}
	for _, status := range ongoing {
		if IsTerminalAgentRunStatus(status) {
			t.Errorf("tillståndet %q ska inte vara terminalt", status)
		}
	}
}

// TestValidAgentProfile kontrollerar vilka agentprofiler som finns.
func TestValidAgentProfile(t *testing.T) {
	for _, profile := range []string{AgentProfileCoding, AgentProfileLife} {
		if !ValidAgentProfile(profile) {
			t.Errorf("agentprofilen %q ska vara giltig", profile)
		}
	}

	for _, profile := range []string{"", "fysgodis", "Coding"} {
		if ValidAgentProfile(profile) {
			t.Errorf("agentprofilen %q ska inte vara giltig", profile)
		}
	}
}
