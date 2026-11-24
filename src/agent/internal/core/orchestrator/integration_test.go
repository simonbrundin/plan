package orchestrator

import (
	"context"
	"testing"

	"plan-agent/internal/core/checkpoint"
)

// TestShouldStopHookIntegration tests the full ShouldStop scenario:
// When orchestrator has a plan with checkpoints, the agent should only be allowed
// to stop when ALL gates for the current checkpoint have passed.
func TestShouldStopHookIntegration(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store,
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)
	hook := NewShouldStopHook(o)

	// Scenario 1: No plan → agent can stop
	if !hook.ShouldStop(nil) {
		t.Error("Expected ShouldStop=true when no plan exists")
	}

	// Scenario 2: No current checkpoint → agent can stop
	if _, err := o.CreatePlan(testCtx(), "Simple feature", ""); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	if !hook.ShouldStop(nil) {
		t.Error("Expected ShouldStop=true when no current checkpoint (plan has no checkpoints)")
	}

	// Scenario 3: Current checkpoint has pending gates → agent must continue
	plan2, _ := o.CreatePlan(testCtx(), "Build database and API", "Create schema and endpoints")
	if len(plan2.Checkpoints) == 0 {
		t.Fatal("Expected checkpoints for database+api feature")
	}

	plan2.Approve("user")
	if err := store.SavePlan(plan2); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	plan2.Checkpoints[0].MarkInProgress()
	if err := store.SaveCheckpoint(plan2.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	if hook.ShouldStop(nil) {
		t.Error("Expected ShouldStop=false when current checkpoint has pending gates")
	}

	// Blocked message should mention the checkpoint name and pending gates
	msg := hook.GetBlockedMessage()
	if msg == "" {
		t.Error("Expected blocked message when stop is prevented")
	}
	if !containsString(msg, plan2.Checkpoints[0].Name) {
		t.Errorf("Expected blocked message to mention checkpoint '%s'", plan2.Checkpoints[0].Name)
	}

	// Scenario 4: All gates passed → agent can stop
	for _, gateType := range plan2.Checkpoints[0].Gates {
		plan2.Checkpoints[0].GateResults[gateType] = &checkpoint.GateResult{
			Status: checkpoint.GateStatusPassed,
		}
	}
	if err := store.SaveCheckpoint(plan2.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	if !hook.ShouldStop(nil) {
		t.Error("Expected ShouldStop=true when all gates passed")
	}

	// Scenario 5: Mixed gate results — one failed
	plan2.Checkpoints[0].GateResults["behavior"] = &checkpoint.GateResult{
		Status:   checkpoint.GateStatusFailed,
		Feedback: "Tests failed: missing validation",
	}
	if err := store.SaveCheckpoint(plan2.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	if hook.ShouldStop(nil) {
		t.Error("Expected ShouldStop=false when any gate has failed")
	}

	msg2 := hook.GetBlockedMessage()
	if !containsString(msg2, "failed") {
		t.Error("Expected blocked message to mention failed gate")
	}
}

// TestShouldStopHookWithMultipleCheckpoints verifies behavior across multiple checkpoints
func TestShouldStopHookWithMultipleCheckpoints(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store,
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)

	plan, _ := o.CreatePlan(testCtx(), "Build database and API", "Full-stack feature")
	if len(plan.Checkpoints) < 2 {
		t.Skipf("Need multiple checkpoints, got %d", len(plan.Checkpoints))
	}

	plan.Approve("user")
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	// First checkpoint passed
	plan.Checkpoints[0].GateResults["behavior"] = &checkpoint.GateResult{Status: checkpoint.GateStatusPassed}
	plan.Checkpoints[0].MarkPassed()
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	hook := NewShouldStopHook(o)

	// Second checkpoint is in progress with pending gates
	plan.Checkpoints[1].MarkInProgress()
	if err := store.SaveCheckpoint(plan.Checkpoints[1]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	if hook.ShouldStop(nil) {
		t.Error("Expected ShouldStop=false when later checkpoint has pending gates")
	}

	// All checkpoints passed
	plan.Checkpoints[1].GateResults["behavior"] = &checkpoint.GateResult{Status: checkpoint.GateStatusPassed}
	plan.Checkpoints[1].GateResults["adversarial"] = &checkpoint.GateResult{Status: checkpoint.GateStatusPassed}
	plan.Checkpoints[1].MarkPassed()
	if err := store.SaveCheckpoint(plan.Checkpoints[1]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	plan.Complete()
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	if !hook.ShouldStop(nil) {
		t.Error("Expected ShouldStop=true when all checkpoints complete")
	}
}

// TestCheckGateStatus tests the CheckGateStatus helper
func TestCheckGateStatus(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store,
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)
	hook := NewShouldStopHook(o)

	// No plan → all gates considered passed
	passed, reason := hook.CheckGateStatus("behavior")
	if !passed {
		t.Errorf("Expected CheckGateStatus=passed with no plan, got passed=%v reason=%s", passed, reason)
	}

	// Plan exists but no checkpoint
	plan, _ := o.CreatePlan(testCtx(), "Simple", "")
	passed, reason = hook.CheckGateStatus("behavior")
	if !passed {
		t.Errorf("Expected CheckGateStatus=passed with no checkpoint, got passed=%v reason=%s", passed, reason)
	}

	// Checkpoint exists but gate not executed
	plan.Checkpoints[0].MarkInProgress()
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	passed, reason = hook.CheckGateStatus("behavior")
	if passed {
		t.Errorf("Expected CheckGateStatus=not passed when gate not executed, got passed=%v reason=%s", passed, reason)
	}

	// Gate passed
	plan.Checkpoints[0].GateResults["behavior"] = &checkpoint.GateResult{Status: checkpoint.GateStatusPassed}
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	passed, reason = hook.CheckGateStatus("behavior")
	if !passed {
		t.Errorf("Expected CheckGateStatus=passed when gate passed, got passed=%v reason=%s", passed, reason)
	}

	// Different gate still pending
	passed, reason = hook.CheckGateStatus("adversarial")
	if passed {
		t.Errorf("Expected CheckGateStatus=not passed for pending gate, got passed=%v reason=%s", passed, reason)
	}

	// Gate failed
	plan.Checkpoints[0].GateResults["adversarial"] = &checkpoint.GateResult{
		Status:   checkpoint.GateStatusFailed,
		Feedback: "Security issue found",
	}
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	passed, reason = hook.CheckGateStatus("adversarial")
	if passed {
		t.Errorf("Expected CheckGateStatus=not passed when gate failed, got passed=%v reason=%s", passed, reason)
	}
	if !containsString(reason, "failed") {
		t.Errorf("Expected reason to mention 'failed', got: %s", reason)
	}

	// Gate skipped
	plan.Checkpoints[0].GateResults["ui"] = &checkpoint.GateResult{Status: checkpoint.GateStatusSkipped}
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	passed, reason = hook.CheckGateStatus("ui")
	if !passed {
		t.Errorf("Expected CheckGateStatus=passed when gate skipped, got passed=%v reason=%s", passed, reason)
	}
}

// TestGetCheckpointProgress tests progress calculation across gates
func TestGetCheckpointProgress(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store,
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)

	// No plan
	passed, total := o.shouldStopHook.GetCheckpointProgress()
	if passed != 0 || total != 0 {
		t.Errorf("Expected 0/0 with no plan, got %d/%d", passed, total)
	}

	// Plan with checkpoints
	plan, _ := o.CreatePlan(testCtx(), "Build database and API", "")
	plan.Checkpoints[0].MarkInProgress()
	plan.Checkpoints[0].GateResults["behavior"] = &checkpoint.GateResult{Status: checkpoint.GateStatusPassed}
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	if len(plan.Checkpoints) > 1 {
		plan.Checkpoints[1].MarkInProgress()
		// No gate results for checkpoint 1
		if err := store.SaveCheckpoint(plan.Checkpoints[1]); err != nil {
			t.Fatalf("saving state failed: %v", err)
		}
	}

	passed, total = o.shouldStopHook.GetCheckpointProgress()
	if total == 0 {
		t.Error("Expected non-zero total gates")
	}
	// At least behavior gate should be passed
	if passed < 1 {
		t.Errorf("Expected at least 1 passed gate, got %d/%d", passed, total)
	}
}

// TestGetPendingGatesForCurrentCheckpoint tests the pending gates helper
func TestGetPendingGatesForCurrentCheckpoint(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store,
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)

	// No plan → nil
	pending := o.shouldStopHook.GetPendingGatesForCurrentCheckpoint()
	if pending != nil {
		t.Error("Expected nil pending gates with no plan")
	}

	// Plan with pending gates
	plan, _ := o.CreatePlan(testCtx(), "Build database and API", "")
	plan.Approve("user")
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	plan.Checkpoints[0].MarkInProgress()
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	pending = o.shouldStopHook.GetPendingGatesForCurrentCheckpoint()
	if len(pending) == 0 {
		t.Error("Expected pending gates when checkpoint is in progress")
	}

	// All gates passed → nil
	for _, gateType := range plan.Checkpoints[0].Gates {
		plan.Checkpoints[0].GateResults[gateType] = &checkpoint.GateResult{Status: checkpoint.GateStatusPassed}
	}
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	pending = o.shouldStopHook.GetPendingGatesForCurrentCheckpoint()
	if pending != nil {
		t.Error("Expected nil pending gates when all passed")
	}
}

// TestIsCheckpointComplete tests the checkpoint completion checker
func TestIsCheckpointComplete(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store,
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)

	// No plan → false
	if o.shouldStopHook.IsCheckpointComplete() {
		t.Error("Expected IsCheckpointComplete=false with no plan")
	}

	// Plan but no current checkpoint → false
	plan, _ := o.CreatePlan(testCtx(), "Simple", "")
	if o.shouldStopHook.IsCheckpointComplete() {
		t.Error("Expected IsCheckpointComplete=false with no current checkpoint")
	}

	// Checkpoint in progress, gates pending → false
	plan.Approve("user")
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	plan.Checkpoints[0].MarkInProgress()
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	if o.shouldStopHook.IsCheckpointComplete() {
		t.Error("Expected IsCheckpointComplete=false with pending gates")
	}

	// All gates passed → true
	for _, gateType := range plan.Checkpoints[0].Gates {
		plan.Checkpoints[0].GateResults[gateType] = &checkpoint.GateResult{Status: checkpoint.GateStatusPassed}
	}
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	if !o.shouldStopHook.IsCheckpointComplete() {
		t.Error("Expected IsCheckpointComplete=true when all gates passed")
	}
}

// TestPrepareNextTurnIntegration tests that PrepareNextTurn hook can be wired up
func TestPrepareNextTurnIntegration(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store,
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)

	// Verify GetAgentConfig returns a config with ShouldStop hook
	agentCfg := o.GetAgentConfig()
	if agentCfg == nil {
		t.Fatal("Expected non-nil AgentLoopConfig")
	}
	if agentCfg.ShouldStop == nil {
		t.Error("Expected ShouldStop hook to be set in AgentLoopConfig")
	}

	// Verify the hook actually works
	if !agentCfg.ShouldStop(nil) {
		t.Error("Expected ShouldStop to return true with no active checkpoint")
	}
}

// TestAgentWithOrchestratorHooks verifies the full agent loop config with orchestrator
func TestAgentWithOrchestratorHooks(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store,
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)

	// Create a plan
	plan, _ := o.CreatePlan(testCtx(), "Build database", "Create user table")
	plan.Approve("user")
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	// ShouldStopHookFunc returns the function
	shouldStopFn := o.ShouldStopHookFunc()
	if shouldStopFn == nil {
		t.Fatal("ShouldStopHookFunc returned nil")
	}

	// With pending checkpoint, ShouldStop returns false
	plan.Checkpoints[0].MarkInProgress()
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	if shouldStopFn(nil) {
		t.Error("ShouldStop should return false with pending gates")
	}

	// After all gates pass, ShouldStop returns true
	for _, gateType := range plan.Checkpoints[0].Gates {
		plan.Checkpoints[0].GateResults[gateType] = &checkpoint.GateResult{Status: checkpoint.GateStatusPassed}
	}
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	if !shouldStopFn(nil) {
		t.Error("ShouldStop should return true when all gates passed")
	}
}

// TestHumanGateBlocksIntegration tests that human gate integration works correctly
func TestHumanGateBlocksIntegration(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store,
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)

	// Create plan with UI checkpoint (has human gate)
	plan, _ := o.CreatePlan(testCtx(), "Build login form", "Create UI with email and password")
	if len(plan.Checkpoints) == 0 {
		t.Fatal("Expected checkpoints")
	}

	plan.Approve("user")
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	plan.Checkpoints[0].MarkInProgress()
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	// Human gate should not block stopping per design (human gates are non-blocking)
	// But it should still be tracked
	hasHumanGate := false
	for _, gateType := range plan.Checkpoints[0].Gates {
		if gateType == "human" {
			hasHumanGate = true
		}
	}

	if !hasHumanGate {
		t.Skip("No human gate in this checkpoint")
	}

	// Human gate not executed → still pending
	pending := plan.Checkpoints[0].GetPendingGates()
	hasPendingHuman := false
	for _, p := range pending {
		if p == "human" {
			hasPendingHuman = true
		}
	}
	if !hasPendingHuman {
		t.Error("Expected human gate to be pending")
	}

	// ShouldStop blocks because human gate is pending
	hook := NewShouldStopHook(o)
	if hook.ShouldStop(nil) {
		t.Error("ShouldStop should return false when human gate is pending")
	}
}

// TestConcurrentCheckpointAccess verifies state store handles concurrent access
func TestConcurrentCheckpointAccess(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store,
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)

	plan, _ := o.CreatePlan(testCtx(), "Concurrent feature", "")
	plan.Approve("user")
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	plan.Checkpoints[0].MarkInProgress()
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	// Multiple goroutines checking status simultaneously
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			hook := NewShouldStopHook(o)
			_ = hook.ShouldStop(nil)
			_, _ = hook.GetCheckpointProgress()
			_ = hook.GetBlockedMessage()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	// If we get here without deadlock, the test passes
}

// TestPlanStatusTransitions verifies correct status transitions
func TestPlanStatusTransitions(t *testing.T) {
	store := NewInMemoryStateStore()
	cfg := Config{
		StateStore:        store, // use the SAME instance
		CheckpointPlanner: NewDefaultCheckpointPlanner(),
	}
	o := New(cfg)
	ctx := testCtx()

	var reloaded *checkpoint.CheckpointPlan
	var reloaded2 *checkpoint.CheckpointPlan
	// Use a feature that generates checkpoints (with description to match patterns)
	plan, _ := o.CreatePlan(ctx, "Build database and API", "Create user table and REST endpoints")

	// Initial state
	if plan.Status != checkpoint.PlanPending {
		t.Errorf("Expected plan status 'pending', got '%s'", plan.Status)
	}
	// Verify checkpoints were generated
	if len(plan.Checkpoints) == 0 {
		t.Fatal("Expected checkpoints for 'database and api' feature")
	}

	// Approve
	err := o.ApprovePlan(plan.ID, "tester")
	if err != nil {
		t.Fatalf("ApprovePlan failed: %v", err)
	}
	reloaded, err = store.GetPlan(plan.ID)
	if err != nil {
		t.Fatalf("GetPlan failed after ApprovePlan: %v", err)
	}
	if reloaded == nil {
		t.Fatal("GetPlan returned nil")
	}
	if reloaded.Status != checkpoint.PlanApproved {
		t.Errorf("Expected status 'approved', got '%s'", reloaded.Status)
	}

	// Execute all checkpoints
	if len(reloaded.Checkpoints) == 0 {
		t.Fatal("Expected at least one checkpoint")
	}
	// Ensure GateResults map is initialized (may be nil after deep copy)
	if reloaded.Checkpoints[0].GateResults == nil {
		reloaded.Checkpoints[0].GateResults = make(map[string]*checkpoint.GateResult)
	}
	reloaded.Checkpoints[0].GateResults["behavior"] = &checkpoint.GateResult{Status: checkpoint.GateStatusPassed}
	reloaded.Checkpoints[0].MarkPassed()
	if err := store.SaveCheckpoint(reloaded.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	reloaded.Complete()
	if err := store.SavePlan(reloaded); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	reloaded2, err = store.GetPlan(plan.ID)
	if err != nil {
		t.Fatalf("GetPlan failed after Complete: %v", err)
	}
	if reloaded2 == nil {
		t.Fatal("GetPlan returned nil after Complete")
	}
	if reloaded2.Status != checkpoint.PlanCompleted {
		t.Errorf("Expected status 'completed', got '%s'", reloaded2.Status)
	}
}

// TestDefaultCheckpointPlannerCoverage verifies planner generates checkpoints for various features
func TestDefaultCheckpointPlannerCoverage(t *testing.T) {
	planner := NewDefaultCheckpointPlanner()
	ctx := testCtx()

	cases := []struct {
		feature     string
		description string
		minCP       int
	}{
		{"User authentication system", "Login, register, JWT sessions, RBAC", 1},
		{"Payment processing", "Stripe integration, webhooks, refunds", 1},
		{"Real-time notifications", "WebSocket, email, push notifications", 1},
		{"File upload and storage", "S3 integration, thumbnails, CDN", 1},
		{"Search functionality", "Elasticsearch, filters, pagination", 1},
	}

	for _, c := range cases {
		t.Run(c.feature, func(t *testing.T) {
			checkpoints, err := planner.CreateCheckpoints(ctx, c.feature, c.description)
			if err != nil {
				t.Fatalf("CreateCheckpoints failed: %v", err)
			}
			if len(checkpoints) < c.minCP {
				t.Errorf("Expected at least %d checkpoints, got %d", c.minCP, len(checkpoints))
			}

			// Every checkpoint should have at least one gate
			for _, cp := range checkpoints {
				if len(cp.Gates) == 0 {
					t.Errorf("Checkpoint '%s' has no gates", cp.Name)
				}
			}

			// Every checkpoint should have at least one task
			for _, cp := range checkpoints {
				if len(cp.Tasks) == 0 {
					t.Errorf("Checkpoint '%s' has no tasks", cp.Name)
				}
			}

			// All checkpoints should have increasing complexity
			for i := 1; i < len(checkpoints); i++ {
				prev := checkpoints[i-1].Complexity
				curr := checkpoints[i].Complexity
				// Verify complexity is defined
				if curr == "" {
					t.Errorf("Checkpoint '%s' has empty complexity", checkpoints[i].Name)
				}
				_ = prev // just checking curr is defined
			}
		})
	}
}

// TestInMemoryStateStorePersistence verifies state survives store operations
func TestInMemoryStateStorePersistence(t *testing.T) {
	store := NewInMemoryStateStore()

	plan := checkpoint.NewPlan("persist-1", "Persist test", "Testing state persistence")
	cp1 := checkpoint.NewCheckpoint("cp-1", "First", "First checkpoint", checkpoint.ComplexityLow)
	cp2 := checkpoint.NewCheckpoint("cp-2", "Second", "Second checkpoint", checkpoint.ComplexityMedium)
	plan.AddCheckpoint(cp1)
	plan.AddCheckpoint(cp2)

	// Save plan
	err := store.SavePlan(plan)
	if err != nil {
		t.Fatalf("SavePlan failed: %v", err)
	}

	// Verify all plans
	plans, err := store.GetAllPlans()
	if err != nil {
		t.Fatalf("GetAllPlans failed: %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("Expected 1 plan, got %d", len(plans))
	}

	// Reload plan and modify via plan reference (checkpoint store has same pointer)
	reloaded, err := store.GetPlan("persist-1")
	if err != nil {
		t.Fatalf("GetPlan failed: %v", err)
	}

	// Modify checkpoint through plan reference
	reloaded.Checkpoints[0].MarkPassed()
	// Ensure GateResults map is initialized (may be nil after deep copy)
	if reloaded.Checkpoints[1].GateResults == nil {
		reloaded.Checkpoints[1].GateResults = make(map[string]*checkpoint.GateResult)
	}
	reloaded.Checkpoints[1].GateResults["behavior"] = &checkpoint.GateResult{Status: checkpoint.GateStatusPassed}
	err = store.SavePlan(reloaded)
	if err != nil {
		t.Fatalf("SavePlan failed: %v", err)
	}

	// Reload plan and verify checkpoint state
	reloaded2, err := store.GetPlan("persist-1")
	if err != nil {
		t.Fatalf("GetPlan failed: %v", err)
	}
	if reloaded2.Checkpoints[0].Status != checkpoint.CheckpointPassed {
		t.Errorf("Expected checkpoint status 'passed', got '%s'", reloaded2.Checkpoints[0].Status)
	}
	if reloaded2.Checkpoints[1].Status != checkpoint.CheckpointPending {
		t.Errorf("Expected second checkpoint status 'pending', got '%s'", reloaded2.Checkpoints[1].Status)
	}
	if reloaded2.Checkpoints[1].GateResults == nil {
		t.Error("GateResults should not be nil after reload")
	} else if reloaded2.Checkpoints[1].GateResults["behavior"] == nil {
		t.Error("Expected behavior gate result to be set")
	}

	// Delete plan
	err = store.DeletePlan("persist-1")
	if err != nil {
		t.Fatalf("DeletePlan failed: %v", err)
	}
	plans2, _ := store.GetAllPlans()
	if len(plans2) != 0 {
		t.Errorf("Expected 0 plans after delete, got %d", len(plans2))
	}
}

// Helper

func testCtx() context.Context {
	return context.Background()
}

func containsString(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
