package models

import "testing"

// TestValidAgentRunPhase kontrollerar vilka faser som accepteras.
func TestValidAgentRunPhase(t *testing.T) {
	valid := []string{
		AgentRunPhaseDefine,
		AgentRunPhasePlan,
		AgentRunPhaseExecute,
		AgentRunPhaseVerify,
		AgentRunPhaseComplete,
	}
	for _, phase := range valid {
		if !ValidAgentRunPhase(phase) {
			t.Errorf("fasen %q ska vara giltig", phase)
		}
	}

	invalid := []string{"", "kokar", "Define", "done"}
	for _, phase := range invalid {
		if ValidAgentRunPhase(phase) {
			t.Errorf("fasen %q ska inte vara giltig", phase)
		}
	}
}

// TestValidAgentRunStatus kontrollerar vilka tillstånd som accepteras.
func TestValidAgentRunStatus(t *testing.T) {
	valid := []string{
		AgentRunStatusQueued,
		AgentRunStatusRunning,
		AgentRunStatusBlocked,
		AgentRunStatusDone,
		AgentRunStatusFailed,
	}
	for _, status := range valid {
		if !ValidAgentRunStatus(status) {
			t.Errorf("tillståndet %q ska vara giltigt", status)
		}
	}

	invalid := []string{"", "running_now", "Done"}
	for _, status := range invalid {
		if ValidAgentRunStatus(status) {
			t.Errorf("tillståndet %q ska inte vara giltigt", status)
		}
	}
}

// TestValidateAgentRunPhaseReturnerarFel kontrollerar att en okänd fas ger
// ett fel - API:t ska svara 400 på sådana.
func TestValidateAgentRunPhaseReturnerarFel(t *testing.T) {
	if err := ValidateAgentRunPhase(AgentRunPhaseExecute); err != nil {
		t.Errorf("giltig fas ska inte ge fel: %v", err)
	}
	if err := ValidateAgentRunPhase("kokar"); err == nil {
		t.Error("okänd fas ska ge fel")
	}
}

// TestValidateAgentRunStatusReturnerarFel kontrollerar att ett okänt
// tillstånd ger ett fel.
func TestValidateAgentRunStatusReturnerarFel(t *testing.T) {
	if err := ValidateAgentRunStatus(AgentRunStatusBlocked); err != nil {
		t.Errorf("giltigt tillstånd ska inte ge fel: %v", err)
	}
	if err := ValidateAgentRunStatus("unknown"); err == nil {
		t.Error("okänt tillstånd ska ge fel")
	}
}
