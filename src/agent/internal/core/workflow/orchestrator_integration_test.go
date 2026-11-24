package workflow

import (
	"testing"

	"plan-agent/internal/core/gate"
	"plan-agent/internal/core/orchestrator"
)

// Regressionstest: InitOrchestrator får inte returnera fel för en tom Config.
//
// Förr slog init-felet tillbaka med "StateStore is required", vilket gjorde att
// server.go fortsatte med en nil-orchestrator. Då registrerades inga
// workflow-endpoints och inga gates kördes - trots att hela checkpoint-flödet
// var byggt för det.
func TestInitOrchestratorWithoutStateStore(t *testing.T) {
	ResetOrchestrator()
	t.Cleanup(ResetOrchestrator)

	o, err := InitOrchestrator(Config{})
	if err != nil {
		t.Fatalf("InitOrchestrator(Config{}) returned error: %v", err)
	}
	if o == nil {
		t.Fatal("expected an orchestrator, got nil")
	}
	if o.StateStore() == nil {
		t.Fatal("expected a state store to be set")
	}
	if GetOrchestrator() == nil {
		t.Fatal("expected the global orchestrator to be set")
	}
}

// En caller som skickar in en egen StateStore ska fortfarande få den.
func TestInitOrchestratorKeepsProvidedStateStore(t *testing.T) {
	ResetOrchestrator()
	t.Cleanup(ResetOrchestrator)

	store := orchestrator.NewInMemoryStateStore()
	o, err := InitOrchestrator(Config{StateStore: store})
	if err != nil {
		t.Fatalf("InitOrchestrator returned error: %v", err)
	}
	if o.StateStore() != store {
		t.Fatal("expected the provided state store to be kept")
	}
}

// Alla fyra gate-typerna ska finnas i default-registret - det är dem som
// avgör om ett checkpoint får passera.
func TestDefaultGateRegistryHasAllGates(t *testing.T) {
	registry := createGateRegistry()

	for _, gateType := range []gate.GateType{
		gate.GateTypeBehavior,
		gate.GateTypeUI,
		gate.GateTypeAdversarial,
		gate.GateTypeHuman,
	} {
		if registry.Get(gateType) == nil {
			t.Errorf("gate %q is not registered", gateType)
		}
	}
}
