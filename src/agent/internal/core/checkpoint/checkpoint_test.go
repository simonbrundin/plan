package checkpoint

import (
	"testing"
)

func TestNewCheckpoint(t *testing.T) {
	cp := NewCheckpoint("cp-1", "Database Schema", "Set up schema", ComplexityLow)

	if cp.ID != "cp-1" {
		t.Errorf("Expected ID 'cp-1', got '%s'", cp.ID)
	}
	if cp.Name != "Database Schema" {
		t.Errorf("Expected Name 'Database Schema', got '%s'", cp.Name)
	}
	if cp.Status != CheckpointPending {
		t.Errorf("Expected Status 'pending', got '%s'", cp.Status)
	}
	if cp.Complexity != ComplexityLow {
		t.Errorf("Expected Complexity 'low', got '%s'", cp.Complexity)
	}
}

func TestCheckpointTasks(t *testing.T) {
	cp := NewCheckpoint("cp-1", "Test", "Test", ComplexityLow)

	cp.AddTask("Task 1")
	cp.AddTask("Task 2")

	if len(cp.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(cp.Tasks))
	}

	if cp.Tasks[0] != "Task 1" {
		t.Errorf("Expected first task 'Task 1', got '%s'", cp.Tasks[0])
	}
}

func TestCheckpointGates(t *testing.T) {
	cp := NewCheckpoint("cp-1", "Test", "Test", ComplexityLow)

	cp.AddGate("behavior")
	cp.AddGate("adversarial")

	if len(cp.Gates) != 2 {
		t.Errorf("Expected 2 gates, got %d", len(cp.Gates))
	}
}

func TestAllGatesPassed(t *testing.T) {
	cp := NewCheckpoint("cp-1", "Test", "Test", ComplexityLow)
	cp.AddGate("behavior")
	cp.AddGate("adversarial")

	// No gates passed yet
	if cp.AllGatesPassed() {
		t.Error("Expected AllGatesPassed() to return false with no gates passed")
	}

	// Pass one gate
	cp.GateResults["behavior"] = &GateResult{
		Status: GateStatusPassed,
	}

	if cp.AllGatesPassed() {
		t.Error("Expected AllGatesPassed() to return false with one gate passed")
	}

	// Pass all gates
	cp.GateResults["adversarial"] = &GateResult{
		Status: GateStatusPassed,
	}

	if !cp.AllGatesPassed() {
		t.Error("Expected AllGatesPassed() to return true with all gates passed")
	}
}

func TestGetPendingGates(t *testing.T) {
	cp := NewCheckpoint("cp-1", "Test", "Test", ComplexityLow)
	cp.AddGate("behavior")
	cp.AddGate("adversarial")
	cp.AddGate("human")

	pending := cp.GetPendingGates()
	if len(pending) != 3 {
		t.Errorf("Expected 3 pending gates, got %d", len(pending))
	}

	// Pass behavior
	cp.GateResults["behavior"] = &GateResult{Status: GateStatusPassed}

	pending = cp.GetPendingGates()
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending gates after passing behavior, got %d", len(pending))
	}
}

func TestMarkPassed(t *testing.T) {
	cp := NewCheckpoint("cp-1", "Test", "Test", ComplexityLow)

	cp.MarkPassed()

	if cp.Status != CheckpointPassed {
		t.Errorf("Expected Status 'passed', got '%s'", cp.Status)
	}
	if cp.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestNewPlan(t *testing.T) {
	plan := NewPlan("plan-1", "Add Auth", "Add authentication")

	if plan.ID != "plan-1" {
		t.Errorf("Expected ID 'plan-1', got '%s'", plan.ID)
	}
	if plan.Feature != "Add Auth" {
		t.Errorf("Expected Feature 'Add Auth', got '%s'", plan.Feature)
	}
	if plan.Status != PlanPending {
		t.Errorf("Expected Status 'pending', got '%s'", plan.Status)
	}
}

func TestPlanCheckpoints(t *testing.T) {
	plan := NewPlan("plan-1", "Test", "Test")
	cp1 := NewCheckpoint("cp-1", "Schema", "Schema", ComplexityLow)
	cp2 := NewCheckpoint("cp-2", "API", "API", ComplexityMedium)

	plan.AddCheckpoint(cp1)
	plan.AddCheckpoint(cp2)

	if len(plan.Checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints, got %d", len(plan.Checkpoints))
	}

	cp := plan.GetCheckpoint("cp-1")
	if cp == nil || cp.Name != "Schema" {
		t.Error("GetCheckpoint failed to find cp-1")
	}

	missing := plan.GetCheckpoint("cp-999")
	if missing != nil {
		t.Error("GetCheckpoint should return nil for non-existent checkpoint")
	}
}

func TestPlanProgress(t *testing.T) {
	plan := NewPlan("plan-1", "Test", "Test")
	plan.AddCheckpoint(NewCheckpoint("cp-1", "1", "1", ComplexityLow))
	plan.AddCheckpoint(NewCheckpoint("cp-2", "2", "2", ComplexityLow))
	plan.AddCheckpoint(NewCheckpoint("cp-3", "3", "3", ComplexityLow))

	passed, total := plan.Progress()
	if passed != 0 || total != 3 {
		t.Errorf("Expected progress 0/3, got %d/%d", passed, total)
	}

	plan.Checkpoints[0].MarkPassed()
	passed, total = plan.Progress()
	if passed != 1 || total != 3 {
		t.Errorf("Expected progress 1/3, got %d/%d", passed, total)
	}

	plan.Checkpoints[1].MarkPassed()
	plan.Checkpoints[2].MarkPassed()
	passed, total = plan.Progress()
	if passed != 3 || total != 3 {
		t.Errorf("Expected progress 3/3, got %d/%d", passed, total)
	}
	if !plan.AllPassed() {
		t.Error("Expected AllPassed() to return true")
	}
}

func TestPlanApproval(t *testing.T) {
	plan := NewPlan("plan-1", "Test", "Test")

	plan.Approve("user@example.com")

	if plan.Status != PlanApproved {
		t.Errorf("Expected Status 'approved', got '%s'", plan.Status)
	}
	if plan.ApprovedBy != "user@example.com" {
		t.Errorf("Expected ApprovedBy 'user@example.com', got '%s'", plan.ApprovedBy)
	}
	if plan.ApprovedAt == nil {
		t.Error("Expected ApprovedAt to be set")
	}
}

func TestGetNextPending(t *testing.T) {
	plan := NewPlan("plan-1", "Test", "Test")
	cp1 := NewCheckpoint("cp-1", "1", "1", ComplexityLow)
	cp2 := NewCheckpoint("cp-2", "2", "2", ComplexityLow)
	cp3 := NewCheckpoint("cp-3", "3", "3", ComplexityLow)

	plan.AddCheckpoint(cp1)
	plan.AddCheckpoint(cp2)
	plan.AddCheckpoint(cp3)

	// First should be cp-1
	next := plan.GetNextPending()
	if next.ID != "cp-1" {
		t.Errorf("Expected GetNextPending to return cp-1, got %s", next.ID)
	}

	// Pass cp-1, next should be cp-2
	cp1.MarkPassed()
	next = plan.GetNextPending()
	if next.ID != "cp-2" {
		t.Errorf("Expected GetNextPending to return cp-2, got %s", next.ID)
	}

	// Pass all, next should be nil
	cp2.MarkPassed()
	cp3.MarkPassed()
	next = plan.GetNextPending()
	if next != nil {
		t.Error("Expected GetNextPending to return nil when all passed")
	}
}

func TestGateResult(t *testing.T) {
	result := NewGateResult("behavior")

	if result.Type != "behavior" {
		t.Errorf("Expected Type 'behavior', got '%s'", result.Type)
	}
	if result.Status != GateStatusPending {
		t.Errorf("Expected Status 'pending', got '%s'", result.Status)
	}

	result.Pass("All tests passed")

	if result.Status != GateStatusPassed {
		t.Errorf("Expected Status 'passed', got '%s'", result.Status)
	}
	if result.Feedback != "All tests passed" {
		t.Errorf("Expected Feedback 'All tests passed', got '%s'", result.Feedback)
	}
	if result.EndedAt.IsZero() {
		t.Error("Expected EndedAt to be set")
	}
}

func TestGateResultFail(t *testing.T) {
	result := NewGateResult("adversarial")

	issues := []Issue{
		{Severity: SeverityHigh, Title: "Memory leak", Description: "Resource not released"},
	}

	result.Fail("Found critical issues", issues)

	if result.Status != GateStatusFailed {
		t.Errorf("Expected Status 'failed', got '%s'", result.Status)
	}
	if len(result.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(result.Issues))
	}
	if result.Issues[0].Title != "Memory leak" {
		t.Errorf("Expected issue title 'Memory leak', got '%s'", result.Issues[0].Title)
	}
}
