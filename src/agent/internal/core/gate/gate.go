package gate

import (
	"context"

	"plan-agent/internal/core/checkpoint"
)

// GateType represents the type of gate
type GateType string

const (
	GateTypeBehavior    GateType = "behavior"
	GateTypeUI          GateType = "ui"
	GateTypeAdversarial GateType = "adversarial"
	GateTypeHuman       GateType = "human"
)

// Gate defines the interface for all gates
type Gate interface {
	// Type returns the gate type
	Type() GateType

	// Name returns a human-readable name
	Name() string

	// Description returns a description of what this gate checks
	Description() string

	// Required returns true if this gate is required for all checkpoints
	Required() bool

	// Run executes the gate check
	Run(ctx context.Context, cp *checkpoint.Checkpoint) (*checkpoint.GateResult, error)
}

// GateRegistry holds all registered gates
type GateRegistry struct {
	gates map[GateType]Gate
}

// NewGateRegistry creates a new gate registry
func NewGateRegistry() *GateRegistry {
	return &GateRegistry{
		gates: make(map[GateType]Gate),
	}
}

// Register adds a gate to the registry
func (r *GateRegistry) Register(g Gate) {
	r.gates[g.Type()] = g
}

// Get returns a gate by type
func (r *GateRegistry) Get(gateType GateType) Gate {
	return r.gates[gateType]
}

// GetAll returns all registered gates
func (r *GateRegistry) GetAll() []Gate {
	gates := make([]Gate, 0, len(r.gates))
	for _, g := range r.gates {
		gates = append(gates, g)
	}
	return gates
}

// GetByTypes returns gates matching the given types
func (r *GateRegistry) GetByTypes(types []string) []Gate {
	var result []Gate
	for _, t := range types {
		if g, ok := r.gates[GateType(t)]; ok {
			result = append(result, g)
		}
	}
	return result
}

// IsBlocking returns true if the gate should block on failure
// Human gates are typically not blocking in the same way
func (g GateType) IsBlocking() bool {
	return g != GateTypeHuman
}

// IsAutomated returns true if the gate runs automatically
func (g GateType) IsAutomated() bool {
	return g != GateTypeHuman
}
