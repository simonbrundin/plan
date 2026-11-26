package orchestrator

import (
	"encoding/json"
	"fmt"
	"sync"

	"plan/agent/internal/core/checkpoint"
)

// StateStore persists checkpoint and plan state
type StateStore interface {
	SavePlan(plan *checkpoint.CheckpointPlan) error
	GetPlan(id string) (*checkpoint.CheckpointPlan, error)
	GetAllPlans() ([]*checkpoint.CheckpointPlan, error)
	DeletePlan(id string) error

	SaveCheckpoint(cp *checkpoint.Checkpoint) error
	GetCheckpoint(id string) (*checkpoint.Checkpoint, error)
	GetCheckpointsByPlan(planID string) ([]*checkpoint.Checkpoint, error)

	SaveGateResult(checkpointID string, result *checkpoint.GateResult) error
	GetGateResult(checkpointIDID, gateType string) (*checkpoint.GateResult, error)
}

// InMemoryStateStore is a simple in-memory implementation
type InMemoryStateStore struct {
	mu          sync.RWMutex
	plans       map[string]*checkpoint.CheckpointPlan
	checkpoints map[string]*checkpoint.Checkpoint
	gateResults map[string]map[string]*checkpoint.GateResult // checkpointID -> gateType -> result
}

// NewInMemoryStateStore creates a new in-memory state store
func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{
		plans:       make(map[string]*checkpoint.CheckpointPlan),
		checkpoints: make(map[string]*checkpoint.Checkpoint),
		gateResults: make(map[string]map[string]*checkpoint.GateResult),
	}
}

// SavePlan saves a checkpoint plan
func (s *InMemoryStateStore) SavePlan(plan *checkpoint.CheckpointPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deep copy to avoid concurrent modification
	planCopy := deepCopyPlan(plan)
	s.plans[plan.ID] = planCopy

	// Also save individual checkpoints
	for _, cp := range plan.Checkpoints {
		s.checkpoints[cp.ID] = cp
	}

	return nil
}

// GetPlan retrieves a plan by ID
func (s *InMemoryStateStore) GetPlan(id string) (*checkpoint.CheckpointPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plan, ok := s.plans[id]
	if !ok {
		return nil, fmt.Errorf("plan not found: %s", id)
	}

	return deepCopyPlan(plan), nil
}

// GetAllPlans returns all plans
func (s *InMemoryStateStore) GetAllPlans() ([]*checkpoint.CheckpointPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plans := make([]*checkpoint.CheckpointPlan, 0, len(s.plans))
	for _, plan := range s.plans {
		plans = append(plans, deepCopyPlan(plan))
	}

	return plans, nil
}

// DeletePlan removes a plan
func (s *InMemoryStateStore) DeletePlan(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.plans, id)
	return nil
}

// SaveCheckpoint saves a checkpoint
func (s *InMemoryStateStore) SaveCheckpoint(cp *checkpoint.Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.checkpoints[cp.ID] = deepCopyCheckpoint(cp)
	return nil
}

// GetCheckpoint retrieves a checkpoint by ID
func (s *InMemoryStateStore) GetCheckpoint(id string) (*checkpoint.Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cp, ok := s.checkpoints[id]
	if !ok {
		return nil, fmt.Errorf("checkpoint not found: %s", id)
	}

	return deepCopyCheckpoint(cp), nil
}

// GetCheckpointsByPlan returns all checkpoints for a plan
func (s *InMemoryStateStore) GetCheckpointsByPlan(planID string) ([]*checkpoint.Checkpoint, error) {
	plan, err := s.GetPlan(planID)
	if err != nil {
		return nil, err
	}

	return plan.Checkpoints, nil
}

// SaveGateResult saves a gate result for a checkpoint
func (s *InMemoryStateStore) SaveGateResult(checkpointID string, result *checkpoint.GateResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.gateResults[checkpointID] == nil {
		s.gateResults[checkpointID] = make(map[string]*checkpoint.GateResult)
	}

	s.gateResults[checkpointID][result.Type] = deepCopyGateResult(result)

	// Also update the checkpoint
	if cp, ok := s.checkpoints[checkpointID]; ok {
		cp.GateResults[result.Type] = result
	}

	return nil
}

// GetGateResult retrieves a gate result
func (s *InMemoryStateStore) GetGateResult(checkpointID, gateType string) (*checkpoint.GateResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if results, ok := s.gateResults[checkpointID]; ok {
		if result, ok := results[gateType]; ok {
			return deepCopyGateResult(result), nil
		}
	}

	return nil, fmt.Errorf("gate result not found: %s/%s", checkpointID, gateType)
}

// Deep copy helpers

func deepCopyPlan(plan *checkpoint.CheckpointPlan) *checkpoint.CheckpointPlan {
	jsonBytes, _ := json.Marshal(plan)
	var copy checkpoint.CheckpointPlan
	_ = json.Unmarshal(jsonBytes, &copy)
	return &copy
}

func deepCopyCheckpoint(cp *checkpoint.Checkpoint) *checkpoint.Checkpoint {
	jsonBytes, _ := json.Marshal(cp)
	var copy checkpoint.Checkpoint
	_ = json.Unmarshal(jsonBytes, &copy)
	return &copy
}

func deepCopyGateResult(r *checkpoint.GateResult) *checkpoint.GateResult {
	jsonBytes, _ := json.Marshal(r)
	var copy checkpoint.GateResult
	_ = json.Unmarshal(jsonBytes, &copy)
	return &copy
}

// FileStateStore persists state to a JSON file
type FileStateStore struct {
	mu       sync.Mutex
	filePath string
	InMemoryStateStore
}

// NewFileStateStore creates a state store that persists to a file
func NewFileStateStore(filePath string) (*FileStateStore, error) {
	store := &FileStateStore{
		filePath: filePath,
	}

	// Try to load existing data
	// In production, would load from file here

	return store, nil
}

// Save implements StateStore
func (s *FileStateStore) SavePlan(plan *checkpoint.CheckpointPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.InMemoryStateStore.SavePlan(plan); err != nil {
		return err
	}
	return s.persist()
}

// SaveCheckpoint implements StateStore
func (s *FileStateStore) SaveCheckpoint(cp *checkpoint.Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.InMemoryStateStore.SaveCheckpoint(cp); err != nil {
		return err
	}
	return s.persist()
}

// SaveGateResult implements StateStore
func (s *FileStateStore) SaveGateResult(checkpointID string, result *checkpoint.GateResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.InMemoryStateStore.SaveGateResult(checkpointID, result); err != nil {
		return err
	}
	return s.persist()
}

func (s *FileStateStore) persist() error {
	plans, _ := s.GetAllPlans()

	data, err := json.MarshalIndent(map[string]interface{}{
		"plans":       plans,
		"checkpoints": s.checkpoints,
		"gateResults": s.gateResults,
	}, "", "  ")
	if err != nil {
		return err
	}

	// In production, would write to s.filePath here
	_ = data
	return nil
}
