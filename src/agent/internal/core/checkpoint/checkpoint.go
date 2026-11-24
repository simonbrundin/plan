package checkpoint

import "time"

// ComplexityLevel represents the complexity of a checkpoint
type ComplexityLevel string

const (
	ComplexityLow    ComplexityLevel = "low"
	ComplexityMedium ComplexityLevel = "medium"
	ComplexityHigh   ComplexityLevel = "high"
)

// CheckpointStatus represents the current status of a checkpoint
type CheckpointStatus string

const (
	CheckpointPending    CheckpointStatus = "pending"
	CheckpointInProgress CheckpointStatus = "in_progress"
	CheckpointPassed     CheckpointStatus = "passed"
	CheckpointFailed     CheckpointStatus = "failed"
	CheckpointSkipped    CheckpointStatus = "skipped"
)

// Checkpoint represents a single unit of work in a feature implementation
type Checkpoint struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Complexity  ComplexityLevel        `json:"complexity"`
	Gates       []string               `json:"gates"` // which gates this checkpoint needs
	Tasks       []string               `json:"tasks"`
	Status      CheckpointStatus       `json:"status"`
	GateResults map[string]*GateResult `json:"gateResults,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"` // Additional metadata (prototypePath, designFile, etc.)
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
}

// NewCheckpoint creates a new checkpoint with defaults
func NewCheckpoint(id, name, description string, complexity ComplexityLevel) *Checkpoint {
	now := time.Now()
	return &Checkpoint{
		ID:          id,
		Name:        name,
		Description: description,
		Complexity:  complexity,
		Tasks:       []string{},
		Status:      CheckpointPending,
		GateResults: make(map[string]*GateResult),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AddTask adds a task to the checkpoint
func (c *Checkpoint) AddTask(task string) {
	c.Tasks = append(c.Tasks, task)
	c.UpdatedAt = time.Now()
}

// AddGate adds a gate requirement to the checkpoint
func (c *Checkpoint) AddGate(gateType string) {
	c.Gates = append(c.Gates, gateType)
	c.UpdatedAt = time.Now()
}

// SetGateResult sets the result for a specific gate
func (c *Checkpoint) SetGateResult(gateType string, result *GateResult) {
	c.GateResults[gateType] = result
	c.UpdatedAt = time.Now()
}

// AllGatesPassed returns true if all required gates have passed
func (c *Checkpoint) AllGatesPassed() bool {
	for _, gateType := range c.Gates {
		result, ok := c.GateResults[gateType]
		if !ok || result.Status != GateStatusPassed {
			return false
		}
	}
	return len(c.Gates) > 0
}

// GetPendingGates returns a list of gates that haven't passed
func (c *Checkpoint) GetPendingGates() []string {
	var pending []string
	for _, gateType := range c.Gates {
		result, ok := c.GateResults[gateType]
		if !ok || result.Status != GateStatusPassed {
			pending = append(pending, gateType)
		}
	}
	return pending
}

// MarkInProgress marks the checkpoint as in progress
func (c *Checkpoint) MarkInProgress() {
	c.Status = CheckpointInProgress
	c.UpdatedAt = time.Now()
}

// MarkPassed marks the checkpoint as passed
func (c *Checkpoint) MarkPassed() {
	c.Status = CheckpointPassed
	now := time.Now()
	c.CompletedAt = &now
	c.UpdatedAt = now
}

// MarkFailed marks the checkpoint as failed
func (c *Checkpoint) MarkFailed() {
	c.Status = CheckpointFailed
	c.UpdatedAt = time.Now()
}

// GetTasks returns the tasks for this checkpoint
func (c *Checkpoint) GetTasks() []string {
	return c.Tasks
}

// GetMetadata returns a metadata value by key
func (c *Checkpoint) GetMetadata(key string) string {
	if c.Metadata == nil {
		return ""
	}
	return c.Metadata[key]
}

// SetMetadata sets a metadata value
func (c *Checkpoint) SetMetadata(key, value string) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]string)
	}
	c.Metadata[key] = value
	c.UpdatedAt = time.Now()
}
