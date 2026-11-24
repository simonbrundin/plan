package workflow

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"plan-agent/internal/core"
	"plan-agent/internal/core/checkpoint"
	"plan-agent/internal/core/gate"
	"plan-agent/internal/core/orchestrator"
	"plan-agent/internal/core/tools"
)

// Type aliases for convenience
type (
	// Config is a type alias for orchestrator.Config
	Config = orchestrator.Config
	// Orchestrator is a type alias for the orchestrator
	Orchestrator = orchestrator.Orchestrator
	// ExecutionResult is a type alias for the execution result
	ExecutionResult = orchestrator.ExecutionResult
	// CheckpointPlan is a type alias for the checkpoint plan
	CheckpointPlan = checkpoint.CheckpointPlan
	// Checkpoint is a type alias for the checkpoint
	Checkpoint = checkpoint.Checkpoint
)

// GlobalOrchestrator holds the global orchestrator instance
var (
	globalMu           sync.RWMutex
	globalOrchestrator *orchestrator.Orchestrator
	globalStateStore   orchestrator.StateStore
)

// InitOrchestrator initializes the global orchestrator
func InitOrchestrator(cfg orchestrator.Config) (*orchestrator.Orchestrator, error) {
	// StateStore is optional - fall back to in-memory instead of failing.
	// Förr slog init-felet tillbaka, vilket gjorde att orchestratorn blev nil i
	// server.go och att inga gate-endpoints registrerades (gates kördes aldrig).
	// Skicka in en CheckpointStore när planen ska överleva en omstart.
	if cfg.StateStore == nil {
		cfg.StateStore = orchestrator.NewInMemoryStateStore()
	}

	// Set up gate registry if not provided
	if cfg.GateRegistry == nil {
		cfg.GateRegistry = createGateRegistry()
	}

	// Set up viewer if not provided
	if cfg.Viewer == nil {
		cfg.Viewer = orchestrator.NewConsoleViewer()
	}

	// Note: CheckpointPlanner setup is now handled by orchestrator.New():
	// - If CheckpointPlanner is provided, use it
	// - If PlannerProvider is provided, create SmartCheckpointPlanner with LLM agent
	// - Otherwise, use DefaultCheckpointPlanner

	// Set up learnings manager
	learningsFile := cfg.LearningFile
	if learningsFile == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			learningsFile = home + "/.config/agent/learnings.md"
		}
	}
	cfg.LearningFile = learningsFile

	o := orchestrator.New(cfg)

	globalMu.Lock()
	globalOrchestrator = o
	globalStateStore = cfg.StateStore
	globalMu.Unlock()

	// Register plan getter for tools package
	tools.RegisterPlanGetter(func() interface{} {
		globalMu.RLock()
		defer globalMu.RUnlock()
		if globalOrchestrator == nil {
			return nil
		}
		return globalOrchestrator.CurrentPlan()
	})

	return o, nil
}

// GetOrchestrator returns the global orchestrator instance
func GetOrchestrator() *orchestrator.Orchestrator {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalOrchestrator
}

// ResetOrchestrator clears the global orchestrator (for testing)
func ResetOrchestrator() {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalOrchestrator = nil
}

// createGateRegistry creates the default gate registry
func createGateRegistry() *gate.GateRegistry {
	registry := gate.NewGateRegistry()
	registry.Register(gate.NewBehaviorGate())
	registry.Register(gate.NewUIGate("", ""))
	registry.Register(gate.NewAdversarialGate())
	registry.Register(gate.NewSimpleHumanGate())
	return registry
}

// AgentConfigForOrchestrator returns an AgentLoopConfig with orchestrator hooks
type AgentConfigForOrchestrator struct {
	BaseConfig   *core.AgentLoopConfig
	Orchestrator *orchestrator.Orchestrator
}

// NewAgentWithOrchestrator creates an agent configured with orchestrator hooks
func NewAgentWithOrchestrator(baseConfig *core.AgentLoopConfig) *core.AgentRuntime {
	o := GetOrchestrator()
	if o == nil {
		// No orchestrator, create agent without hooks
		return core.NewAgent(baseConfig)
	}

	// Clone the config to avoid modifying the original
	config := *baseConfig

	// Add the ShouldStop hook
	config.ShouldStop = o.ShouldStopHookFunc()

	// Add PrepareNextTurn hook to update context with learnings
	config.PrepareNextTurn = createPrepareNextTurnHook(o)

	// Inject learnings into system prompt BEFORE first turn
	config.SystemPrompt = injectLearningsIntoPrompt(config.SystemPrompt, o)

	return core.NewAgent(&config)
}

// injectLearningsIntoPrompt adds learnings from the learnings manager to the system prompt
func injectLearningsIntoPrompt(systemPrompt string, o *orchestrator.Orchestrator) string {
	learningsCtx := o.GenerateLearningsContext()
	if learningsCtx == "" {
		return systemPrompt
	}

	// Check if learnings are already in the prompt
	if strings.Contains(systemPrompt, "## Relevant Learnings") {
		return systemPrompt
	}

	// Prepend learnings to system prompt
	// Place after any existing header but before main instructions
	if systemPrompt == "" {
		return learningsCtx
	}

	// Insert learnings after any initial header/section
	return systemPrompt + "\n\n" + learningsCtx
}

// createPrepareNextTurnHook creates the PrepareNextTurn hook
// This hook runs after each turn and can modify the next turn's context
func createPrepareNextTurnHook(o *orchestrator.Orchestrator) func(*core.TurnContext) *core.TurnUpdate {
	return func(turnCtx *core.TurnContext) *core.TurnUpdate {
		plan := o.CurrentPlan()
		if plan == nil {
			return nil
		}

		// Get current checkpoint progress
		current := plan.GetCurrent()
		if current == nil {
			return nil
		}

		// Build checkpoint progress message
		progress := buildCheckpointProgress(current)
		if progress == "" {
			return nil
		}

		// Get existing system prompt
		agentCtx := turnCtx.Context
		if agentCtx == nil {
			return nil
		}

		// Check if checkpoint context is already in system prompt
		if strings.Contains(agentCtx.SystemPrompt, "## Current Checkpoint") {
			// Update existing checkpoint progress
			agentCtx.SystemPrompt = updateCheckpointProgress(agentCtx.SystemPrompt, progress)
		} else {
			// Add checkpoint progress to system prompt
			agentCtx.SystemPrompt += "\n\n" + progress
		}

		// CRITICAL: If human gate is pending, inject instruction to call ask_human_approval
		if hasPendingHumanGate(current) {
			approvalInstruction := buildHumanApprovalInstruction(current)
			agentCtx.SystemPrompt += "\n\n" + approvalInstruction
		}

		return &core.TurnUpdate{
			Context: agentCtx,
		}
	}
}

// hasPendingHumanGate checks if the checkpoint has a pending human gate
func hasPendingHumanGate(cp *checkpoint.Checkpoint) bool {
	for _, gateType := range cp.Gates {
		if gateType == "human" {
			result, exists := cp.GateResults[gateType]
			if !exists || result.Status != checkpoint.GateStatusPassed {
				return true
			}
		}
	}
	return false
}

// buildHumanApprovalInstruction builds the instruction for asking human approval
func buildHumanApprovalInstruction(cp *checkpoint.Checkpoint) string {
	var lines []string

	lines = append(lines, "## 🚨 HUMAN APPROVAL REQUIRED")
	lines = append(lines, "")
	lines = append(lines, "This checkpoint requires human approval before you can continue.")
	lines = append(lines, "")
	lines = append(lines, "**You MUST use the `ask_human_approval` tool NOW** to request approval.")
	lines = append(lines, "")
	lines = append(lines, "Do NOT continue working on other tasks. Stop and call `ask_human_approval` immediately.")
	lines = append(lines, "")
	lines = append(lines, "Example:")
	lines = append(lines, "```")
	lines = append(lines, "await ask_human_approval({")
	lines = append(lines, fmt.Sprintf("  checkpoint_name: \"%s\",", cp.Name))
	lines = append(lines, fmt.Sprintf("  summary: \"Completed %d tasks\",", len(cp.Tasks)))
	lines = append(lines, "  feedback_request: \"Please verify the implementation\"")
	lines = append(lines, "})")
	lines = append(lines, "```")

	return strings.Join(lines, "\n")
}

// buildCheckpointProgress builds a progress message for the current checkpoint
func buildCheckpointProgress(cp *checkpoint.Checkpoint) string {
	var lines []string

	lines = append(lines, "## Current Checkpoint Progress")
	lines = append(lines, fmt.Sprintf("**Checkpoint**: %s", cp.Name))
	lines = append(lines, fmt.Sprintf("**Status**: %s", cp.Status))

	if len(cp.Tasks) > 0 {
		lines = append(lines, "\n**Tasks**:")
		for i, task := range cp.Tasks {
			lines = append(lines, fmt.Sprintf("  %d. %s", i+1, task))
		}
	}

	if len(cp.GateResults) > 0 {
		lines = append(lines, "\n**Gate Status**:")
		for gateType, result := range cp.GateResults {
			status := fmt.Sprintf("❌ %s", result.Status)
			if result.Status == "passed" {
				status = fmt.Sprintf("✅ %s", result.Status)
			} else if result.Status == "pending" || result.Status == "running" {
				status = fmt.Sprintf("⏳ %s", result.Status)
			}
			lines = append(lines, fmt.Sprintf("  - %s: %s", gateType, status))
			if result.Feedback != "" {
				lines = append(lines, fmt.Sprintf("    Feedback: %s", result.Feedback))
			}
		}
	}

	pending := cp.GetPendingGates()
	if len(pending) > 0 {
		lines = append(lines, fmt.Sprintf("\n**Pending Gates**: %s", strings.Join(pending, ", ")))
		lines = append(lines, "\nEnsure all gates pass before marking checkpoint complete.")
	}

	// Add human gate warning
	if hasPendingHumanGate(cp) {
		lines = append(lines, "")
		lines = append(lines, "⚠️ **HUMAN APPROVAL REQUIRED** - Use `ask_human_approval` to request user input")
	}

	return strings.Join(lines, "\n")
}

// updateCheckpointProgress updates the checkpoint progress section in the system prompt
func updateCheckpointProgress(systemPrompt, newProgress string) string {
	// Find and replace the checkpoint section
	startMarker := "## Current Checkpoint Progress"
	start := strings.Index(systemPrompt, startMarker)
	if start == -1 {
		return systemPrompt + "\n\n" + newProgress
	}

	// Find the end (next ## or end of string)
	end := strings.Index(systemPrompt[start:], "\n## ")
	if end == -1 {
		end = len(systemPrompt)
	} else {
		end += start
	}

	// Replace the section
	return systemPrompt[:start] + newProgress + systemPrompt[end:]
}

// WorkflowCommands handles workflow-related commands
type WorkflowCommands struct {
	orchestrator *orchestrator.Orchestrator
}

// NewWorkflowCommands creates a new workflow commands handler
func NewWorkflowCommands(o *orchestrator.Orchestrator) *WorkflowCommands {
	return &WorkflowCommands{orchestrator: o}
}

// StartWorkflow starts a new workflow for a feature
func (w *WorkflowCommands) StartWorkflow(ctx context.Context, feature, description string) (*checkpoint.CheckpointPlan, error) {
	plan, err := w.orchestrator.CreatePlan(ctx, feature, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}
	return plan, nil
}

// ApprovePlan approves a checkpoint plan
func (w *WorkflowCommands) ApprovePlan(planID, approvedBy string) error {
	return w.orchestrator.ApprovePlan(planID, approvedBy)
}

// RejectPlan rejects a checkpoint plan
func (w *WorkflowCommands) RejectPlan(planID string) error {
	return w.orchestrator.RejectPlan(planID)
}

// GetPlanStatus returns the status of a plan
func (w *WorkflowCommands) GetPlanStatus(planID string) (*checkpoint.CheckpointPlan, error) {
	store := getStateStore()
	if store == nil {
		return nil, fmt.Errorf("orchestrator not initialized")
	}
	return store.GetPlan(planID)
}

// GetStateStore returns the global state store for external use
func GetStateStore() orchestrator.StateStore {
	return getStateStore()
}

// SetStateStore sets the global state store (useful for testing)
func SetStateStore(store orchestrator.StateStore) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalStateStore = store
}

// GetCurrentCheckpoint returns the current checkpoint being executed
func (w *WorkflowCommands) GetCurrentCheckpoint() *checkpoint.Checkpoint {
	plan := w.orchestrator.CurrentPlan()
	if plan == nil {
		return nil
	}
	return plan.GetCurrent()
}

// GetCheckpointProgress returns progress info across all checkpoints
func (w *WorkflowCommands) GetCheckpointProgress() (passed, total int) {
	plan := w.orchestrator.CurrentPlan()
	if plan == nil {
		return 0, 0
	}
	for _, cp := range plan.Checkpoints {
		for _, gateType := range cp.Gates {
			total++
			result, exists := cp.GateResults[gateType]
			if exists && result.Status == checkpoint.GateStatusPassed {
				passed++
			}
		}
	}
	return passed, total
}

// SubmitHumanDecision submits a human decision for a checkpoint
func (w *WorkflowCommands) SubmitHumanDecision(checkpointID string, approved bool, feedback string) error {
	// Find the human gate and submit decision
	// This would integrate with the human gate's channel
	_ = checkpointID
	_ = approved
	_ = feedback
	return nil
}

// getStateStore returns the global state store
func getStateStore() orchestrator.StateStore {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalStateStore
}
