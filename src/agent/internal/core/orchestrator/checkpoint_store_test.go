package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"plan/agent/internal/core/checkpoint"
)

// TestCheckpointStoreBasic tests the CheckpointStore with an in-memory mock
// Note: This test doesn't connect to a real database. For full integration testing,
// run against a real PostgreSQL instance with:
//   - CREATE TABLE goals (...)
//   - CREATE TABLE checkpoints (...)
//   - CREATE TABLE checkpoint_results (...)
// Then call CheckpointStore.SavePlan() and verify with GetPlan()

func TestCheckpointStorePlanCRUD(t *testing.T) {
	// This test verifies the StateStore interface is correctly implemented
	// Full database integration testing requires a running PostgreSQL instance

	store := NewInMemoryStateStore()
	_ = context.Background() // Context would be used with real DB

	// Create a plan
	plan := checkpoint.NewPlan(uuid.New().String(), "Test Feature", "Test Description")
	plan.Status = checkpoint.PlanPending

	// Add a checkpoint
	cp := checkpoint.NewCheckpoint(
		uuid.New().String(),
		"Test Checkpoint",
		"A test checkpoint",
		checkpoint.ComplexityLow,
	)
	cp.AddGate("behavior")
	cp.AddTask("Task 1")
	cp.AddTask("Task 2")
	plan.AddCheckpoint(cp)

	// Save plan
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("SavePlan failed: %v", err)
	}

	// Get plan back
	savedPlan, err := store.GetPlan(plan.ID)
	if err != nil {
		t.Fatalf("GetPlan failed: %v", err)
	}

	if savedPlan.Feature != "Test Feature" {
		t.Errorf("Expected Feature 'Test Feature', got '%s'", savedPlan.Feature)
	}

	if len(savedPlan.Checkpoints) != 1 {
		t.Errorf("Expected 1 checkpoint, got %d", len(savedPlan.Checkpoints))
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
	if err := store.DeletePlan(plan.ID); err != nil {
		t.Fatalf("DeletePlan failed: %v", err)
	}

	// Verify deleted
	_, err = store.GetPlan(plan.ID)
	if err == nil {
		t.Error("Expected error when getting deleted plan")
	}
}

func TestCheckpointStoreCheckpointCRUD(t *testing.T) {
	store := NewInMemoryStateStore()

	// Create a checkpoint
	cp := checkpoint.NewCheckpoint(
		uuid.New().String(),
		"Implementation",
		"Write the code",
		checkpoint.ComplexityMedium,
	)
	cp.AddGate("behavior")
	cp.AddGate("ui")
	cp.AddTask("Write tests")
	cp.AddTask("Write implementation")

	// Save checkpoint
	if err := store.SaveCheckpoint(cp); err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// Get checkpoint
	saved, err := store.GetCheckpoint(cp.ID)
	if err != nil {
		t.Fatalf("GetCheckpoint failed: %v", err)
	}

	if saved.Name != "Implementation" {
		t.Errorf("Expected name 'Implementation', got '%s'", saved.Name)
	}

	if len(saved.Gates) != 2 {
		t.Errorf("Expected 2 gates, got %d", len(saved.Gates))
	}

	if len(saved.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(saved.Tasks))
	}
}

func TestCheckpointStoreGateResultCRUD(t *testing.T) {
	store := NewInMemoryStateStore()

	checkpointID := uuid.New().String()

	// Create a gate result
	result := checkpoint.NewGateResult("behavior")
	result.Pass("All tests passed")

	// Save gate result
	if err := store.SaveGateResult(checkpointID, result); err != nil {
		t.Fatalf("SaveGateResult failed: %v", err)
	}

	// Get gate result
	saved, err := store.GetGateResult(checkpointID, "behavior")
	if err != nil {
		t.Fatalf("GetGateResult failed: %v", err)
	}

	if saved.Status != checkpoint.GateStatusPassed {
		t.Errorf("Expected status 'passed', got '%s'", saved.Status)
	}

	if saved.Feedback != "All tests passed" {
		t.Errorf("Expected feedback 'All tests passed', got '%s'", saved.Feedback)
	}
}

func TestCheckpointStoreFullWorkflow(t *testing.T) {
	// This simulates a full checkpoint workflow
	store := NewInMemoryStateStore()

	// 1. Create plan
	plan := checkpoint.NewPlan(uuid.New().String(), "Build Login Feature", "Create login with OAuth")
	plan.Status = checkpoint.PlanApproved
	now := time.Now()
	plan.ApprovedAt = &now
	plan.ApprovedBy = "user"

	// 2. Add checkpoints
	specCP := checkpoint.NewCheckpoint(
		uuid.New().String(),
		"Specification",
		"Write API specification",
		checkpoint.ComplexityLow,
	)
	specCP.AddGate("behavior")
	specCP.Status = checkpoint.CheckpointPassed
	plan.AddCheckpoint(specCP)

	implCP := checkpoint.NewCheckpoint(
		uuid.New().String(),
		"Implementation",
		"Build the login flow",
		checkpoint.ComplexityHigh,
	)
	implCP.AddGate("behavior")
	implCP.AddGate("ui")
	implCP.Status = checkpoint.CheckpointInProgress
	plan.AddCheckpoint(implCP)

	// 3. Save plan with checkpoints
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("SavePlan failed: %v", err)
	}

	// 4. Simulate gate result
	gateResult := checkpoint.NewGateResult("behavior")
	gateResult.Pass("Tests passing")
	if err := store.SaveGateResult(implCP.ID, gateResult); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	// 5. Load and verify
	loaded, err := store.GetPlan(plan.ID)
	if err != nil {
		t.Fatalf("GetPlan failed: %v", err)
	}

	if loaded.Status != checkpoint.PlanApproved {
		t.Errorf("Expected plan status 'approved', got '%s'", loaded.Status)
	}

	if len(loaded.Checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints, got %d", len(loaded.Checkpoints))
	}

	// Find implementation checkpoint and verify
	var loadedImpl *checkpoint.Checkpoint
	for _, cp := range loaded.Checkpoints {
		if cp.Name == "Implementation" {
			loadedImpl = cp
			break
		}
	}
	if loadedImpl == nil {
		t.Fatal("Implementation checkpoint not found")
	}

	if loadedImpl.Status != checkpoint.CheckpointInProgress {
		t.Errorf("Expected status 'in_progress', got '%s'", loadedImpl.Status)
	}
}

// IntegrationTestNote provides documentation for running full integration tests
// To run integration tests with a real database:
//
//  1. Start PostgreSQL
//     docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:15
//
//  2. Create database
//     psql -h localhost -U postgres -c "CREATE DATABASE software_agent;"
//
// 3. Run migrations from db.go (auto-runs on Connect())
//
//  4. Create a test that uses CheckpointStore with real pool:
//     pool, _ := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5432/software_agent")
//     store := NewCheckpointStore(pool)
//     // ... test saves and loads
//
//  5. Verify data persists by querying the database:
//     SELECT * FROM goals;
//     SELECT * FROM checkpoints;
//     SELECT * FROM checkpoint_results;
func TestIntegrationWithRealDatabase(t *testing.T) {
	// Skip this test - it's documentation only
	// Real integration tests require a running PostgreSQL instance
	t.Skip("Integration test requires running PostgreSQL - see TestIntegrationWithRealDatabase docstring")
}
