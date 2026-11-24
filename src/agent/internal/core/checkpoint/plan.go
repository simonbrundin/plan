package checkpoint

import "time"

// PlanStatus represents the status of a checkpoint plan
type PlanStatus string

const (
	PlanPending   PlanStatus = "pending"
	PlanApproved  PlanStatus = "approved"
	PlanRejected  PlanStatus = "rejected"
	PlanCompleted PlanStatus = "completed"
)

// CheckpointPlan represents a complete plan for implementing a feature
type CheckpointPlan struct {
	ID          string        `json:"id"`
	Feature     string        `json:"feature"`
	Description string        `json:"description"`
	Checkpoints []*Checkpoint `json:"checkpoints"`
	Status      PlanStatus    `json:"status"`
	ApprovedBy  string        `json:"approvedBy,omitempty"`
	ApprovedAt  *time.Time    `json:"approvedAt,omitempty"`
	RejectedAt  *time.Time    `json:"rejectedAt,omitempty"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

// NewPlan creates a new checkpoint plan
func NewPlan(id, feature, description string) *CheckpointPlan {
	now := time.Now()
	return &CheckpointPlan{
		ID:          id,
		Feature:     feature,
		Description: description,
		Checkpoints: []*Checkpoint{},
		Status:      PlanPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AddCheckpoint adds a checkpoint to the plan in order
func (p *CheckpointPlan) AddCheckpoint(cp *Checkpoint) {
	p.Checkpoints = append(p.Checkpoints, cp)
	p.UpdatedAt = time.Now()
}

// GetCheckpoint returns a checkpoint by ID
func (p *CheckpointPlan) GetCheckpoint(id string) *Checkpoint {
	for _, cp := range p.Checkpoints {
		if cp.ID == id {
			return cp
		}
	}
	return nil
}

// GetNextPending returns the next pending checkpoint
func (p *CheckpointPlan) GetNextPending() *Checkpoint {
	for _, cp := range p.Checkpoints {
		if cp.Status == CheckpointPending || cp.Status == CheckpointInProgress {
			return cp
		}
	}
	return nil
}

// GetCurrent returns the currently executing checkpoint (if any)
func (p *CheckpointPlan) GetCurrent() *Checkpoint {
	for _, cp := range p.Checkpoints {
		if cp.Status == CheckpointInProgress {
			return cp
		}
	}
	return nil
}

// AllPassed returns true if all checkpoints have passed
func (p *CheckpointPlan) AllPassed() bool {
	for _, cp := range p.Checkpoints {
		if cp.Status != CheckpointPassed && cp.Status != CheckpointSkipped {
			return false
		}
	}
	return len(p.Checkpoints) > 0
}

// Progress returns the progress as a fraction (e.g., 2/5)
func (p *CheckpointPlan) Progress() (passed, total int) {
	for _, cp := range p.Checkpoints {
		total++
		if cp.Status == CheckpointPassed || cp.Status == CheckpointSkipped {
			passed++
		}
	}
	return
}

// Approve marks the plan as approved
func (p *CheckpointPlan) Approve(approvedBy string) {
	p.Status = PlanApproved
	p.ApprovedBy = approvedBy
	now := time.Now()
	p.ApprovedAt = &now
	p.UpdatedAt = now
}

// Reject marks the plan as rejected
func (p *CheckpointPlan) Reject() {
	p.Status = PlanRejected
	now := time.Now()
	p.RejectedAt = &now
	p.UpdatedAt = now
}

// Complete marks the plan as completed
func (p *CheckpointPlan) Complete() {
	p.Status = PlanCompleted
	p.UpdatedAt = time.Now()
}

// GateResult represents the result of a gate check
type GateResult struct {
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	Feedback  string    `json:"feedback,omitempty"`
	Issues    []Issue   `json:"issues,omitempty"`
	Summary   string    `json:"summary,omitempty"`
}

// NewGateResult creates a new gate result
func NewGateResult(gateType string) *GateResult {
	return &GateResult{
		Type:      gateType,
		Status:    GateStatusPending,
		StartedAt: time.Now(),
		Issues:    []Issue{},
	}
}

// Pass marks the gate as passed
func (g *GateResult) Pass(feedback string) {
	g.Status = GateStatusPassed
	g.EndedAt = time.Now()
	g.Duration = g.EndedAt.Sub(g.StartedAt).String()
	g.Feedback = feedback
}

// Fail marks the gate as failed
func (g *GateResult) Fail(feedback string, issues []Issue) {
	g.Status = GateStatusFailed
	g.EndedAt = time.Now()
	g.Duration = g.EndedAt.Sub(g.StartedAt).String()
	g.Feedback = feedback
	g.Issues = issues
}

// Skip marks the gate as skipped
func (g *GateResult) Skip(reason string) {
	g.Status = GateStatusSkipped
	g.EndedAt = time.Now()
	g.Duration = g.EndedAt.Sub(g.StartedAt).String()
	g.Feedback = reason
}

// Issue represents a problem found during gate review
type Issue struct {
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location,omitempty"` // file:line
	Suggestion  string `json:"suggestion,omitempty"`
}

// Gate statuses
const (
	GateStatusPending = "pending"
	GateStatusRunning = "running"
	GateStatusPassed  = "passed"
	GateStatusFailed  = "failed"
	GateStatusSkipped = "skipped"
)

// Issue severities
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)
