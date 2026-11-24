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

// TestValidAgentRunTrigger kontrollerar vilka triggrar som accepteras.
func TestValidAgentRunTrigger(t *testing.T) {
	valid := []string{"user", "schedule", "webhook", "agent"}
	for _, trigger := range valid {
		if !ValidAgentRunTrigger(trigger) {
			t.Errorf("triggern %q ska vara giltig", trigger)
		}
	}

	invalid := []string{"", "kokar", "User", "cron"}
	for _, trigger := range invalid {
		if ValidAgentRunTrigger(trigger) {
			t.Errorf("triggern %q ska inte vara giltig", trigger)
		}
	}
}

// TestValidateAgentRunTriggerReturnerarFel kontrollerar att en okänd trigger
// ger ett fel - API:t ska svara 400 på sådana.
func TestValidateAgentRunTriggerReturnerarFel(t *testing.T) {
	if err := ValidateAgentRunTrigger("agent"); err != nil {
		t.Errorf("giltig trigger ska inte ge fel: %v", err)
	}
	if err := ValidateAgentRunTrigger("fysgodis"); err == nil {
		t.Error("okänd trigger ska ge fel")
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

func TestValidateAgentRunTransitionTillaterLivscykel(t *testing.T) {
	phase := AgentRunPhasePlan
	status := AgentRunStatusDone
	current := AgentRun{Phase: AgentRunPhaseDefine, Status: AgentRunStatusRunning}

	if err := ValidateAgentRunTransition(current, UpdateAgentRunRequest{Phase: &phase, Status: &status}); err != nil {
		t.Fatalf("giltig övergång ska tillåtas: %v", err)
	}
}

func TestValidateAgentRunTransitionStopparBakåtgång(t *testing.T) {
	phase := AgentRunPhasePlan
	current := AgentRun{Phase: AgentRunPhaseExecute, Status: AgentRunStatusRunning}

	if err := ValidateAgentRunTransition(current, UpdateAgentRunRequest{Phase: &phase}); err == nil {
		t.Fatal("fasen ska inte kunna gå bakåt")
	}
}

func TestValidateAgentRunTransitionStopparTerminalRun(t *testing.T) {
	status := AgentRunStatusRunning
	current := AgentRun{Phase: AgentRunPhasePlan, Status: AgentRunStatusDone}

	if err := ValidateAgentRunTransition(current, UpdateAgentRunRequest{Status: &status}); err == nil {
		t.Fatal("en avslutad körning ska inte kunna återupplivas")
	}
}
