package orchestrator

import (
	"context"
	"testing"

	"plan/agent/internal/core/checkpoint"
)

func TestNewOrchestrator(t *testing.T) {
	cfg := Config{
		StateStore: NewInMemoryStateStore(),
	}

	o := New(cfg)

	if o == nil {
		t.Fatal("Expected non-nil orchestrator")
	}

	if o.ID() == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestCreatePlanWithPlanner(t *testing.T) {
	cfg := Config{
		StateStore:        NewInMemoryStateStore(),
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}

	o := New(cfg)
	ctx := context.Background()

	// Use a feature that will generate checkpoints (has database/API keywords)
	plan, err := o.CreatePlan(ctx, "Add user database and API", "Create user table and REST endpoints")
	if err != nil {
		t.Fatalf("CreatePlan failed: %v", err)
	}

	if plan.Feature != "Add user database and API" {
		t.Errorf("Expected Feature 'Add user database and API', got '%s'", plan.Feature)
	}
	if plan.Status != "pending" {
		t.Errorf("Expected Status 'pending', got '%s'", plan.Status)
	}
	if len(plan.Checkpoints) == 0 {
		t.Error("Expected checkpoints to be generated")
	}
}

func TestApprovePlan(t *testing.T) {
	cfg := Config{
		StateStore: NewInMemoryStateStore(),
	}

	o := New(cfg)
	ctx := context.Background()

	plan, _ := o.CreatePlan(ctx, "Test Feature", "")

	err := o.ApprovePlan(plan.ID, "test-user")
	if err != nil {
		t.Fatalf("ApprovePlan failed: %v", err)
	}

	// Reload plan from store
	reloaded, _ := cfg.StateStore.GetPlan(plan.ID)
	if reloaded.Status != "approved" {
		t.Errorf("Expected Status 'approved', got '%s'", reloaded.Status)
	}
	if reloaded.ApprovedBy != "test-user" {
		t.Errorf("Expected ApprovedBy 'test-user', got '%s'", reloaded.ApprovedBy)
	}
}

func TestRejectPlan(t *testing.T) {
	cfg := Config{
		StateStore: NewInMemoryStateStore(),
	}

	o := New(cfg)
	ctx := context.Background()

	plan, _ := o.CreatePlan(ctx, "Test Feature", "")

	err := o.RejectPlan(plan.ID)
	if err != nil {
		t.Fatalf("RejectPlan failed: %v", err)
	}

	reloaded, _ := cfg.StateStore.GetPlan(plan.ID)
	if reloaded.Status != "rejected" {
		t.Errorf("Expected Status 'rejected', got '%s'", reloaded.Status)
	}
}

func TestCurrentPlan(t *testing.T) {
	cfg := Config{
		StateStore: NewInMemoryStateStore(),
	}

	o := New(cfg)
	ctx := context.Background()

	if o.CurrentPlan() != nil {
		t.Error("Expected CurrentPlan to be nil initially")
	}

	if _, err := o.CreatePlan(ctx, "Test", ""); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	if o.CurrentPlan() == nil {
		t.Error("Expected CurrentPlan to be set after CreatePlan")
	}
}

func TestDefaultCheckpointPlanner(t *testing.T) {
	planner := NewDefaultCheckpointPlanner()
	ctx := context.Background()

	// Test with database-related feature
	checkpoints, err := planner.CreateCheckpoints(ctx, "Add user database", "Create user table and model")
	if err != nil {
		t.Fatalf("CreateCheckpoints failed: %v", err)
	}

	if len(checkpoints) == 0 {
		t.Error("Expected checkpoints to be generated for database feature")
	}

	// Verify database checkpoint exists
	found := false
	for _, cp := range checkpoints {
		if cp.Name == "Database Schema" {
			found = true
			if cp.Complexity != "low" {
				t.Errorf("Expected Database Schema complexity 'low', got '%s'", cp.Complexity)
			}
			break
		}
	}
	if !found {
		t.Error("Expected 'Database Schema' checkpoint for database feature")
	}
}

func TestDefaultCheckpointPlannerAPI(t *testing.T) {
	planner := NewDefaultCheckpointPlanner()
	ctx := context.Background()

	checkpoints, err := planner.CreateCheckpoints(ctx, "Create REST API", "Build CRUD endpoints")
	if err != nil {
		t.Fatalf("CreateCheckpoints failed: %v", err)
	}

	found := false
	for _, cp := range checkpoints {
		if cp.Name == "API Endpoints" {
			found = true
			if cp.Complexity != "medium" {
				t.Errorf("Expected API Endpoints complexity 'medium', got '%s'", cp.Complexity)
			}
			// Should have behavior and adversarial gates
			hasBehavior := false
			hasAdversarial := false
			for _, gate := range cp.Gates {
				if gate == "behavior" {
					hasBehavior = true
				}
				if gate == "adversarial" {
					hasAdversarial = true
				}
			}
			if !hasBehavior {
				t.Error("Expected 'behavior' gate in API checkpoint")
			}
			if !hasAdversarial {
				t.Error("Expected 'adversarial' gate in API checkpoint")
			}
			break
		}
	}
	if !found {
		t.Error("Expected 'API Endpoints' checkpoint for API feature")
	}
}

func TestDefaultCheckpointPlannerUI(t *testing.T) {
	planner := NewDefaultCheckpointPlanner()
	ctx := context.Background()

	checkpoints, err := planner.CreateCheckpoints(ctx, "Build login form", "Create UI with email and password fields")
	if err != nil {
		t.Fatalf("CreateCheckpoints failed: %v", err)
	}

	found := false
	for _, cp := range checkpoints {
		if cp.Name == "User Interface" {
			found = true
			// Should have UI gate
			hasUI := false
			for _, gate := range cp.Gates {
				if gate == "ui" {
					hasUI = true
					break
				}
			}
			if !hasUI {
				t.Error("Expected 'ui' gate in UI checkpoint")
			}
			break
		}
	}
	if !found {
		t.Error("Expected 'User Interface' checkpoint for UI feature")
	}
}

func TestInMemoryStateStore(t *testing.T) {
	store := NewInMemoryStateStore()

	// Create and save a plan
	plan := checkpoint.NewPlan("plan-1", "Test", "Test")
	cp := checkpoint.NewCheckpoint("cp-1", "Checkpoint 1", "Test", checkpoint.ComplexityLow)
	plan.AddCheckpoint(cp)

	err := store.SavePlan(plan)
	if err != nil {
		t.Fatalf("SavePlan failed: %v", err)
	}

	// Retrieve plan
	reloaded, err := store.GetPlan("plan-1")
	if err != nil {
		t.Fatalf("GetPlan failed: %v", err)
	}
	if reloaded.Feature != "Test" {
		t.Errorf("Expected Feature 'Test', got '%s'", reloaded.Feature)
	}
	if len(reloaded.Checkpoints) != 1 {
		t.Errorf("Expected 1 checkpoint, got %d", len(reloaded.Checkpoints))
	}

	// Get all plans
	plans, err := store.GetAllPlans()
	if err != nil {
		t.Fatalf("GetAllPlans failed: %v", err)
	}
	if len(plans) != 1 {
		t.Errorf("Expected 1 plan, got %d", len(plans))
	}

	// Delete plan
	err = store.DeletePlan("plan-1")
	if err != nil {
		t.Fatalf("DeletePlan failed: %v", err)
	}
	_, err = store.GetPlan("plan-1")
	if err == nil {
		t.Error("Expected error when getting deleted plan")
	}
}

func TestShouldStopHook(t *testing.T) {
	cfg := Config{
		StateStore:        NewInMemoryStateStore(),
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}

	o := New(cfg)
	ctx := context.Background()

	// Create plan with a feature that generates checkpoints
	plan, _ := o.CreatePlan(ctx, "Add database and API", "Create table and endpoints")
	if len(plan.Checkpoints) == 0 {
		t.Skip("No checkpoints generated, skipping test")
	}

	plan.Approve("user")
	if err := cfg.StateStore.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	hook := NewShouldStopHook(o)

	// No current checkpoint - should allow stop
	if !hook.ShouldStop(nil) {
		t.Error("Expected ShouldStop to return true with no current checkpoint")
	}

	// Set current checkpoint
	plan.Checkpoints[0].MarkInProgress()
	if err := cfg.StateStore.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	// Has pending gates - should not allow stop
	if hook.ShouldStop(nil) {
		t.Error("Expected ShouldStop to return false with pending gates")
	}

	// Pass all gates
	for _, gateType := range plan.Checkpoints[0].Gates {
		plan.Checkpoints[0].GateResults[gateType] = &checkpoint.GateResult{
			Status: "passed",
		}
	}

	if !hook.ShouldStop(nil) {
		t.Error("Expected ShouldStop to return true when all gates passed")
	}
}

func TestConsoleViewerShowPlan(t *testing.T) {
	viewer := NewConsoleViewer()

	plan := checkpoint.NewPlan("plan-1", "Test Feature", "Test description")
	plan.AddCheckpoint(checkpoint.NewCheckpoint("cp-1", "Checkpoint 1", "First step", checkpoint.ComplexityLow))

	// Should not panic
	viewer.ShowPlan(plan)
}

func TestConsoleViewerShowCheckpoint(t *testing.T) {
	viewer := NewConsoleViewer()

	cp := checkpoint.NewCheckpoint("cp-1", "Checkpoint 1", "First step", checkpoint.ComplexityLow)
	cp.AddTask("Task 1")
	cp.AddTask("Task 2")

	// Should not panic
	viewer.ShowCheckpoint(cp)
}

func TestConsoleViewerUpdateCheckpoint(t *testing.T) {
	viewer := NewConsoleViewer()

	cp := checkpoint.NewCheckpoint("cp-1", "Checkpoint 1", "First step", checkpoint.ComplexityLow)
	cp.AddGate("behavior")
	cp.GateResults["behavior"] = &checkpoint.GateResult{
		Status: "passed",
	}

	// Should not panic
	viewer.UpdateCheckpoint(cp)
}

func TestPlanGetCurrent(t *testing.T) {
	plan := checkpoint.NewPlan("plan-1", "Test", "Test")
	cp1 := checkpoint.NewCheckpoint("cp-1", "1", "1", checkpoint.ComplexityLow)
	cp2 := checkpoint.NewCheckpoint("cp-2", "2", "2", checkpoint.ComplexityMedium)

	plan.AddCheckpoint(cp1)
	plan.AddCheckpoint(cp2)

	// No current
	current := plan.GetCurrent()
	if current != nil {
		t.Error("Expected nil current with no in-progress checkpoints")
	}

	// Set cp2 as in progress
	cp2.MarkInProgress()

	current = plan.GetCurrent()
	if current == nil || current.ID != "cp-2" {
		t.Error("Expected cp-2 as current")
	}
}
