package orchestrator

import (
	"context"

	"plan/agent/internal/core/gate"
)

// NewDefaultGateRegistry creates a registry with all default gates
func NewDefaultGateRegistry() *gate.GateRegistry {
	registry := gate.NewGateRegistry()

	// Register behavior gate (always needed)
	registry.Register(gate.NewBehaviorGate())

	// UI gate is optional (only for UI checkpoints)
	registry.Register(gate.NewUIGate("", ""))

	// Adversarial gate for code review
	registry.Register(gate.NewAdversarialGate())

	// Human gate for manual approval
	registry.Register(gate.NewSimpleHumanGate())

	return registry
}

// GateAgentFactory wraps GateAgent creation for gates
type GateAgentFactory struct {
	model string
}

// ProviderWrapper wraps the provider manager for gate use
type ProviderWrapper struct{}

// NewGateAgentFactory creates a factory that produces gate-compatible agents
func NewGateAgentFactory(provider interface{}, model string) *GateAgentFactory {
	return &GateAgentFactory{
		model: model,
	}
}

// CreateAgent creates a gate.Agent from the factory
func (f *GateAgentFactory) CreateAgent(config gate.AgentConfig) gate.Agent {
	return &gateAgentAdapter{
		name:         config.Name,
		systemPrompt: config.SystemPrompt,
		model:        f.model,
	}
}

// gateAgentAdapter implements gate.Agent for use in gates
type gateAgentAdapter struct {
	name         string
	systemPrompt string
	model        string
}

// Run executes the agent (placeholder - needs provider)
func (a *gateAgentAdapter) Run(ctx context.Context, prompt string) (string, error) {
	// This is a placeholder - actual implementation requires provider access
	// The orchestrator should set up the factory with proper provider access
	return "", nil
}

// NewUIGateWithAgentFactory creates a UIGate with agent factory
func NewUIGateWithAgentFactory(prototypePath, designFile string, factory gate.AgentFactory) *gate.UIGate {
	return gate.NewUIGateWithFactory(prototypePath, designFile, factory)
}

// NewGateRegistryWithAgentFactory creates a registry with agent factory for UI gate
func NewGateRegistryWithAgentFactory(factory gate.AgentFactory) *gate.GateRegistry {
	registry := gate.NewGateRegistry()

	// Register behavior gate (always needed)
	registry.Register(gate.NewBehaviorGate())

	// UI gate with agent factory for sub-agent spawning
	registry.Register(gate.NewUIGateWithFactory("", "", factory))

	// Adversarial gate for code review
	registry.Register(gate.NewAdversarialGate())

	// Human gate for manual approval
	registry.Register(gate.NewSimpleHumanGate())

	return registry
}
