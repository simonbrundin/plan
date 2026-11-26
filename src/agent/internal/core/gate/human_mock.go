package gate

import (
	"context"
	"sync"

	"plan/agent/internal/core/checkpoint"
)

// MockHumanGate is for testing purposes
type MockHumanGate struct {
	mu        sync.Mutex
	decisions map[string]bool
}

func NewMockHumanGate() *MockHumanGate {
	return &MockHumanGate{decisions: make(map[string]bool)}
}

func (m *MockHumanGate) Approve(checkpointID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.decisions[checkpointID] = true
}

func (m *MockHumanGate) Reject(checkpointID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.decisions[checkpointID] = false
}

func (m *MockHumanGate) Type() GateType      { return GateTypeHuman }
func (m *MockHumanGate) Name() string        { return "Mock Human Gate" }
func (m *MockHumanGate) Description() string { return "Mock human gate for testing" }
func (m *MockHumanGate) Required() bool      { return true }

func (m *MockHumanGate) Run(ctx context.Context, cp *checkpoint.Checkpoint) (*checkpoint.GateResult, error) {
	result := checkpoint.NewGateResult(string(GateTypeHuman))

	m.mu.Lock()
	approved, exists := m.decisions[cp.ID]
	m.mu.Unlock()

	if !exists {
		result.Status = checkpoint.GateStatusPending
		return result, nil
	}

	if approved {
		result.Pass("Mock approved")
	} else {
		result.Fail("Mock rejected", nil)
	}

	return result, nil
}
