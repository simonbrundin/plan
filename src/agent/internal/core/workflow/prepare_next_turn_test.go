package workflow

import (
	"context"
	"testing"

	"plan/agent/internal/core"
	"plan/agent/internal/core/checkpoint"
	"plan/agent/internal/core/messages"
	"plan/agent/internal/core/orchestrator"
	"plan/agent/internal/core/tools"
)

func TestCreatePrepareNextTurnHook(t *testing.T) {
	store := orchestrator.NewInMemoryStateStore()
	cfg := orchestrator.Config{
		StateStore:        store,
		CheckpointPlanner: orchestrator.NewDefaultCheckpointPlanner(),
	}
	o := orchestrator.New(cfg)

	// Create a plan with checkpoints (using feature that generates checkpoints)
	plan, _ := o.CreatePlan(context.Background(), "Build database and API", "Create users table and REST endpoints")
	if len(plan.Checkpoints) == 0 {
		t.Fatal("Expected checkpoints")
	}

	// Approve and mark checkpoint in progress
	plan.Approve("tester")
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	plan.Checkpoints[0].MarkInProgress()
	// Ensure GateResults map is initialized
	plan.Checkpoints[0].GateResults = make(map[string]*checkpoint.GateResult)
	plan.Checkpoints[0].GateResults["behavior"] = &checkpoint.GateResult{
		Status:   checkpoint.GateStatusPassed,
		Feedback: "Tests passed",
	}
	if err := store.SaveCheckpoint(plan.Checkpoints[0]); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	// Create the hook
	hook := createPrepareNextTurnHook(o)

	// Create a turn context
	turnCtx := &core.TurnContext{
		Context: &core.AgentContext{
			SystemPrompt: "You are an agent.",
			Messages:     []messages.Message{},
			Tools:        []*tools.ToolDefinition{},
		},
	}

	// Execute the hook
	update := hook(turnCtx)

	if update == nil {
		t.Fatal("Expected non-nil update")
	}

	if update.Context == nil {
		t.Fatal("Expected non-nil context in update")
	}

	// Verify checkpoint progress was added to system prompt
	if update.Context.SystemPrompt == "" {
		t.Error("Expected system prompt to be updated")
	}

	if !contains(update.Context.SystemPrompt, "Checkpoint Progress") {
		t.Error("Expected checkpoint progress in system prompt")
	}

	if !contains(update.Context.SystemPrompt, plan.Checkpoints[0].Name) {
		t.Errorf("Expected checkpoint name '%s' in system prompt", plan.Checkpoints[0].Name)
	}
}

func TestPrepareNextTurnHookWithNoPlan(t *testing.T) {
	store := orchestrator.NewInMemoryStateStore()
	cfg := orchestrator.Config{
		StateStore: store,
	}
	o := orchestrator.New(cfg)

	hook := createPrepareNextTurnHook(o)

	turnCtx := &core.TurnContext{
		Context: &core.AgentContext{
			SystemPrompt: "Original prompt",
		},
	}

	update := hook(turnCtx)

	// Should return nil when no plan exists
	if update != nil {
		t.Error("Expected nil update when no plan exists")
	}
}

func TestPrepareNextTurnHookUpdatesExistingProgress(t *testing.T) {
	store := orchestrator.NewInMemoryStateStore()
	cfg := orchestrator.Config{
		StateStore:        store,
		CheckpointPlanner: orchestrator.NewDefaultCheckpointPlanner(),
	}
	o := orchestrator.New(cfg)

	// Create a plan with checkpoints
	plan, _ := o.CreatePlan(context.Background(), "Build database", "Create table and API")
	if len(plan.Checkpoints) == 0 {
		t.Fatal("Expected checkpoints")
	}
	plan.Approve("tester")
	plan.Checkpoints[0].MarkInProgress()
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	hook := createPrepareNextTurnHook(o)

	// First turn - add checkpoint progress
	turnCtx1 := &core.TurnContext{
		Context: &core.AgentContext{
			SystemPrompt: "You are an agent.\n\n## Current Checkpoint Progress\n**Old Status**",
		},
	}

	update1 := hook(turnCtx1)
	if update1 == nil {
		t.Fatal("Expected update")
	}

	// Second turn - should update existing progress
	turnCtx2 := &core.TurnContext{
		Context: update1.Context,
	}

	update2 := hook(turnCtx2)
	if update2 == nil {
		t.Fatal("Expected update on second turn")
	}

	// Should still contain checkpoint progress
	if !contains(update2.Context.SystemPrompt, "Checkpoint Progress") {
		t.Error("Expected checkpoint progress to be preserved")
	}
}

func TestBuildCheckpointProgress(t *testing.T) {
	cp := &checkpoint.Checkpoint{
		Name:        "Database Schema",
		Description: "Set up database",
		Status:      checkpoint.CheckpointInProgress,
		Tasks:       []string{"Create users table", "Create posts table"},
		Gates:       []string{"behavior", "adversarial"},
		GateResults: map[string]*checkpoint.GateResult{
			"behavior": {Status: checkpoint.GateStatusPassed},
		},
	}

	progress := buildCheckpointProgress(cp)

	if !contains(progress, "Database Schema") {
		t.Error("Expected checkpoint name in progress")
	}

	if !contains(progress, "in_progress") {
		t.Error("Expected status in progress")
	}

	if !contains(progress, "Create users table") {
		t.Error("Expected tasks in progress")
	}

	if !contains(progress, "behavior") {
		t.Error("Expected gate info in progress")
	}

	// Should show pending gates
	if !contains(progress, "adversarial") {
		t.Error("Expected pending gate info")
	}
}

func TestUpdateCheckpointProgress(t *testing.T) {
	tests := []struct {
		name         string
		systemPrompt string
		newProgress  string
		wantContains []string
		dontContain  []string
	}{
		{
			name:         "no existing progress",
			systemPrompt: "You are an agent.",
			newProgress:  "## Current Checkpoint Progress\n**Checkpoint**: Test",
			wantContains: []string{"Current Checkpoint Progress", "Test"},
		},
		{
			name:         "with existing progress",
			systemPrompt: "You are an agent.\n\n## Current Checkpoint Progress\n**Old**\n\n## Other Section",
			newProgress:  "## Current Checkpoint Progress\n**New**",
			wantContains: []string{"New"},
			dontContain:  []string{"Old"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateCheckpointProgress(tt.systemPrompt, tt.newProgress)

			for _, want := range tt.wantContains {
				if !contains(result, want) {
					t.Errorf("Expected '%s' in result", want)
				}
			}

			for _, dont := range tt.dontContain {
				if contains(result, dont) {
					t.Errorf("Did not expect '%s' in result", dont)
				}
			}
		})
	}
}

func TestPrepareNextTurnHookPreservesContext(t *testing.T) {
	store := orchestrator.NewInMemoryStateStore()
	cfg := orchestrator.Config{
		StateStore:        store,
		CheckpointPlanner: orchestrator.NewDefaultCheckpointPlanner(),
	}
	o := orchestrator.New(cfg)

	plan, _ := o.CreatePlan(context.Background(), "Build API", "Create endpoints")
	if len(plan.Checkpoints) == 0 {
		t.Fatal("Expected checkpoints")
	}
	plan.Approve("tester")
	plan.Checkpoints[0].MarkInProgress()
	if err := store.SavePlan(plan); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	hook := createPrepareNextTurnHook(o)

	turnCtx := &core.TurnContext{
		Context: &core.AgentContext{
			SystemPrompt: "Original prompt",
			Messages:     []messages.Message{},
			Tools:        []*tools.ToolDefinition{},
		},
	}

	update := hook(turnCtx)

	// Context should be the same object, just modified
	if update.Context != turnCtx.Context {
		t.Error("Expected same context object")
	}
}

// Helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
