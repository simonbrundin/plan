package models

import (
	"slices"
	"testing"
)

// TestAgentRunPhasesÄrKanoniska kontrollerar att faslistan är den enda
// definitionen - en fas som inte står i listan är inte giltig.
func TestAgentRunPhasesÄrKanoniska(t *testing.T) {
	if len(AgentRunPhases) == 0 {
		t.Fatal("faslistan får inte vara tom")
	}

	seen := map[string]bool{}
	for _, phase := range AgentRunPhases {
		if seen[phase] {
			t.Errorf("fasen %q finns två gånger i listan", phase)
		}
		seen[phase] = true

		if !ValidAgentRunPhase(phase) {
			t.Errorf("fasen %q i listan ska vara giltig", phase)
		}
	}
}

// TestAgentRunStatusesÄrKanoniska kontrollerar samma sak för tillstånden.
func TestAgentRunStatusesÄrKanoniska(t *testing.T) {
	if len(AgentRunStatuses) == 0 {
		t.Fatal("tillståndslistan får inte vara tom")
	}

	seen := map[string]bool{}
	for _, status := range AgentRunStatuses {
		if seen[status] {
			t.Errorf("tillståndet %q finns två gånger i listan", status)
		}
		seen[status] = true

		if !ValidAgentRunStatus(status) {
			t.Errorf("tillståndet %q i listan ska vara giltigt", status)
		}
	}
}

// TestÖvergångstabellenÄrSammansatt kontrollerar att övergångstabellen bara
// pekar på kända tillstånd och att terminala tillstånd inte går vidare.
func TestÖvergångstabellenÄrSammansatt(t *testing.T) {
	transitions := AgentRunStatusTransitions()

	for från, till := range transitions {
		if !ValidAgentRunStatus(från) {
			t.Errorf("övergångstabellen har okänt tillstånd %q", från)
		}

		if IsTerminalAgentRunStatus(från) && len(till) > 0 {
			t.Errorf("terminalt tillstånd %q får inte ha övergångar, fick %v", från, till)
		}

		for _, status := range till {
			if !ValidAgentRunStatus(status) {
				t.Errorf("övergången %q -> %q pekar på ett okänt tillstånd", från, status)
			}
		}
	}

	for _, status := range AgentRunStatuses {
		if _, ok := transitions[status]; !ok {
			t.Errorf("tillståndet %q saknas i övergångstabellen", status)
		}
	}
}

// TestVarjeKörningKanAvbrytas kontrollerar att en körning alltid kan stanna
// - av en människa, av en budget eller av en policy.
func TestVarjeKörningKanAvbrytas(t *testing.T) {
	cancelled := AgentRunStatusCancelled

	for _, current := range []string{
		AgentRunStatusQueued,
		AgentRunStatusRunning,
		AgentRunStatusBlocked,
	} {
		run := AgentRun{Phase: AgentRunPhasePlan, Status: current}
		if err := ValidateAgentRunTransition(run, UpdateAgentRunRequest{Status: &cancelled}); err != nil {
			t.Errorf("körning i %q ska kunna avbrytas: %v", current, err)
		}
	}
}

// TestBlockedVäntarPåMänniska kontrollerar skillnaden mellan "blocked" och
// "running" - blocked är det tillstånd där en människa måste agera.
func TestBlockedVäntarPåMänniska(t *testing.T) {
	if !AgentRunStatusNeedsHuman(AgentRunStatusBlocked) {
		t.Error("blocked ska kräva en människa")
	}

	for _, status := range []string{
		AgentRunStatusQueued,
		AgentRunStatusRunning,
		AgentRunStatusDone,
		AgentRunStatusFailed,
		AgentRunStatusCancelled,
	} {
		if AgentRunStatusNeedsHuman(status) {
			t.Errorf("tillståndet %q ska inte kräva en människa", status)
		}
	}

	if IsTerminalAgentRunStatus(AgentRunStatusBlocked) {
		t.Error("blocked är inte terminalt - en körning kan återupptas")
	}
}

// TestAgentRunStatusTransitionsReturnerarKopia kontrollerar att en caller
// inte kan ändra den kanoniska tabellen genom att ändra det returnerade
// värdet.
func TestAgentRunStatusTransitionsReturnerarKopia(t *testing.T) {
	transitions := AgentRunStatusTransitions()
	transitions[AgentRunStatusQueued] = []string{AgentRunStatusDone}

	if slices.Contains(AgentRunStatusTransitions()[AgentRunStatusQueued], AgentRunStatusDone) &&
		!slices.Contains(agentRunStatusTransitions[AgentRunStatusQueued], AgentRunStatusDone) {
		t.Fatal("den kanoniska tabellen ska inte kunna ändras utifrån")
	}
}

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
		AgentRunStatusCancelled,
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

// TestValidateAgentRunTransitionTillaterAvbrytning kontrollerar att en
// körning kan avbrytas från varje tillstånd - det är ett sätt att stanna,
// inte ett sätt att gå vidare.
func TestValidateAgentRunTransitionTillaterAvbrytning(t *testing.T) {
	cancelled := AgentRunStatusCancelled

	for _, current := range []string{
		AgentRunStatusQueued,
		AgentRunStatusRunning,
		AgentRunStatusBlocked,
	} {
		run := AgentRun{Phase: AgentRunPhasePlan, Status: current}
		if err := ValidateAgentRunTransition(run, UpdateAgentRunRequest{Status: &cancelled}); err != nil {
			t.Errorf("körning i %q ska kunna avbrytas: %v", current, err)
		}
	}
}

// TestValidateAgentRunTransitionStopparAvbrutenRun kontrollerar att en
// avbruten körning inte kan återupplivas.
func TestValidateAgentRunTransitionStopparAvbrutenRun(t *testing.T) {
	running := AgentRunStatusRunning
	current := AgentRun{Phase: AgentRunPhasePlan, Status: AgentRunStatusCancelled}

	if err := ValidateAgentRunTransition(current, UpdateAgentRunRequest{Status: &running}); err == nil {
		t.Fatal("en avbruten körning ska inte kunna återupplivas")
	}
}
