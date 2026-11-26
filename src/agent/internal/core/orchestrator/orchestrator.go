package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"plan/agent/internal/core"
	"plan/agent/internal/core/checkpoint"
	"plan/agent/internal/core/gate"
	"plan/agent/internal/core/provider"
)

// Orchestrator coordinates the checkpoint-gate workflow
type Orchestrator struct {
	id                string
	gates             *gate.GateRegistry
	stateStore        StateStore
	agentFactory      AgentFactory
	checkpointPlanner CheckpointPlanner
	shouldStopHook    ShouldStopHook
	viewer            Viewer
	learningsManager  *LearningsManager
	testPlanner       TestPlanner
	mu                sync.RWMutex
	currentPlan       *checkpoint.CheckpointPlan
	learningFile      string
}

// Config contains configuration for the orchestrator
type Config struct {
	// GateRegistry contains all registered gates
	GateRegistry *gate.GateRegistry

	// StateStore persists checkpoint plan state
	StateStore StateStore

	// AgentFactory creates agents for sub-tasks
	AgentFactory AgentFactory

	// CheckpointPlanner creates checkpoint plans from features
	CheckpointPlanner CheckpointPlanner

	// PlannerProvider is an optional LLM provider manager used to create
	// a SmartCheckpointPlanner when CheckpointPlanner is not provided.
	PlannerProvider *provider.ProviderManager

	// PlannerModel is the model to use for planning (defaults to MiniMax-M2.7)
	PlannerModel string

	// Viewer displays checkpoint status to humans
	Viewer Viewer

	// LearningFile is the path to the learnings file
	LearningFile string

	// LearningsManager manages learnings persistence
	LearningsManager *LearningsManager

	// TestPlanner generates tests for checkpoints (TDD-style)
	TestPlanner TestPlanner

	// TestPlannerProvider is an optional LLM provider for test planning
	TestPlannerProvider *provider.ProviderManager

	// TestPlannerModel is the model for test planning
	TestPlannerModel string

	// UIGateProvider is an optional LLM provider for UI gate sub-agents
	UIGateProvider *provider.ProviderManager

	// UIGateModel is the model for UI gate (defaults to MiniMax-M2.7)
	UIGateModel string
}

// PlannerProvider is an interface for getting LLM completions
// This allows the orchestrator to create a SmartCheckpointPlanner
type PlannerProvider interface {
	// Complete returns a completion for the given prompt
	Complete(ctx context.Context, prompt string) (string, error)
}

// New creates a new orchestrator
func New(cfg Config) *Orchestrator {
	// Set up UI gate with provider if configured
	if cfg.GateRegistry == nil && cfg.UIGateProvider != nil {
		// Create a gate-compatible agent factory for UI gate
		agentFactory := NewUIGateAgentFactory(cfg.UIGateProvider, cfg.UIGateModel)
		cfg.GateRegistry = NewGateRegistryWithAgentFactory(agentFactory)
	} else if cfg.GateRegistry == nil {
		cfg.GateRegistry = NewDefaultGateRegistry()
	}
	if cfg.StateStore == nil {
		cfg.StateStore = NewInMemoryStateStore()
	}
	if cfg.Viewer == nil {
		cfg.Viewer = &ConsoleViewer{}
	}
	// If CheckpointPlanner is not set but PlannerProvider is, create a SmartCheckpointPlanner
	if cfg.CheckpointPlanner == nil && cfg.PlannerProvider != nil {
		plannerAgent := NewPlannerAgent(cfg.PlannerProvider, cfg.PlannerModel)
		cfg.CheckpointPlanner = NewSmartCheckpointPlanner(plannerAgent)
	}

	o := &Orchestrator{
		id:                uuid.New().String(),
		gates:             cfg.GateRegistry,
		stateStore:        cfg.StateStore,
		agentFactory:      cfg.AgentFactory,
		checkpointPlanner: cfg.CheckpointPlanner,
		viewer:            cfg.Viewer,
		learningFile:      cfg.LearningFile,
	}

	// Set up the ShouldStop hook
	o.shouldStopHook = NewShouldStopHook(o)

	// Set up learnings manager if configured
	if cfg.LearningsManager != nil {
		o.learningsManager = cfg.LearningsManager
	} else if cfg.LearningFile != "" {
		o.learningsManager = NewLearningsManager(cfg.LearningFile)
	}

	// Set up test planner if configured
	if cfg.TestPlanner != nil {
		o.testPlanner = cfg.TestPlanner
	} else if cfg.TestPlannerProvider != nil {
		// Create LLM-based test planner
		testAgent := NewPlannerAgent(cfg.TestPlannerProvider, cfg.TestPlannerModel)
		o.testPlanner = NewLLMTestPlanner(testAgent)
	} else {
		// Use default test planner (generates basic tests)
		o.testPlanner = NewDefaultTestPlanner()
	}

	return o
}

// ID returns the orchestrator's ID
func (o *Orchestrator) ID() string {
	return o.id
}

// StateStore returns the state store
func (o *Orchestrator) StateStore() StateStore {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.stateStore
}

// Viewer returns the viewer
func (o *Orchestrator) Viewer() Viewer {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.viewer
}

// CurrentPlan returns the current checkpoint plan
func (o *Orchestrator) CurrentPlan() *checkpoint.CheckpointPlan {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.currentPlan
}

// LearningsManager returns the learnings manager
func (o *Orchestrator) LearningsManager() *LearningsManager {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.learningsManager
}

// AddLearning adds a learning entry
func (o *Orchestrator) AddLearning(category, title, description, lesson string) error {
	if o.learningsManager == nil {
		return nil
	}
	return o.learningsManager.AddLearning(category, title, description, lesson)
}

// AddCheckpointLearning adds a learning from a checkpoint
func (o *Orchestrator) AddCheckpointLearning(cpName, issue, fix, lesson string) error {
	if o.learningsManager == nil {
		return nil
	}
	return o.learningsManager.AddCheckpointLearning(cpName, issue, fix, lesson)
}

// GetLearnings returns all learnings as a string
func (o *Orchestrator) GetLearnings() (string, error) {
	if o.learningsManager == nil {
		return "", nil
	}
	return o.learningsManager.GetLearnings()
}

// GenerateLearningsContext generates context for agent prompts
// If categories are provided, only learnings matching those categories are included
func (o *Orchestrator) GenerateLearningsContext(categories ...string) string {
	if o.learningsManager == nil {
		return ""
	}
	return o.learningsManager.GeneratePromptContext(categories...)
}

// PlanTests generates tests for a checkpoint using TDD approach
// Tests are generated BEFORE implementation, so they will initially FAIL
func (o *Orchestrator) PlanTests(ctx context.Context, cp *checkpoint.Checkpoint) ([]*TestCase, error) {
	if o.testPlanner == nil {
		return nil, fmt.Errorf("no test planner configured")
	}
	return o.testPlanner.PlanTests(ctx, cp)
}

// CreatePlan creates a checkpoint plan for a feature
func (o *Orchestrator) CreatePlan(ctx context.Context, feature, description string) (*checkpoint.CheckpointPlan, error) {
	plan := checkpoint.NewPlan(uuid.New().String(), feature, description)

	// Use checkpoint planner if available
	if o.checkpointPlanner != nil {
		checkpoints, err := o.checkpointPlanner.CreateCheckpoints(ctx, feature, description)
		if err != nil {
			return nil, fmt.Errorf("checkpoint planning failed: %w", err)
		}
		for _, cp := range checkpoints {
			plan.AddCheckpoint(cp)
		}
	}

	// Save plan to state store
	if err := o.stateStore.SavePlan(plan); err != nil {
		log.Printf("Warning: failed to save plan: %v", err)
	}

	o.mu.Lock()
	o.currentPlan = plan
	o.mu.Unlock()

	return plan, nil
}

// ApprovePlan marks a plan as approved
func (o *Orchestrator) ApprovePlan(planID string, approvedBy string) error {
	plan, err := o.stateStore.GetPlan(planID)
	if err != nil {
		return fmt.Errorf("plan not found: %w", err)
	}

	plan.Approve(approvedBy)
	return o.stateStore.SavePlan(plan)
}

// RejectPlan marks a plan as rejected
func (o *Orchestrator) RejectPlan(planID string) error {
	plan, err := o.stateStore.GetPlan(planID)
	if err != nil {
		return fmt.Errorf("plan not found: %w", err)
	}

	plan.Reject()
	return o.stateStore.SavePlan(plan)
}

// Execute runs the checkpoint-gate workflow
func (o *Orchestrator) Execute(ctx context.Context, feature, description string) (*ExecutionResult, error) {
	// Create or load plan
	plan, err := o.CreatePlan(ctx, feature, description)
	if err != nil {
		return nil, err
	}

	// Show plan to viewer for approval
	o.viewer.ShowPlan(plan)

	// Wait for approval (this would be async in production)
	// For now, we assume the plan is already approved
	if plan.Status != checkpoint.PlanApproved {
		return &ExecutionResult{
			PlanID: plan.ID,
			Status: StatusPendingApproval,
		}, nil
	}

	// Execute checkpoints
	return o.executePlan(ctx, plan)
}

// executePlan executes all checkpoints in the plan
func (o *Orchestrator) executePlan(ctx context.Context, plan *checkpoint.CheckpointPlan) (*ExecutionResult, error) {
	result := &ExecutionResult{
		PlanID:    plan.ID,
		Status:    StatusInProgress,
		StartTime: time.Now(),
	}

	for i, cp := range plan.Checkpoints {
		result.CurrentCheckpoint = i + 1
		result.TotalCheckpoints = len(plan.Checkpoints)

		// Show checkpoint to viewer
		o.viewer.ShowCheckpoint(cp)

		// Execute checkpoint
		cpResult, err := o.executeCheckpoint(ctx, cp)
		if err != nil {
			result.Status = StatusFailed
			result.Error = err.Error()
			return result, err
		}

		result.CheckpointResults = append(result.CheckpointResults, cpResult)

		if !cpResult.Passed {
			result.Status = StatusFailed
			return result, fmt.Errorf("checkpoint %s failed", cp.Name)
		}

		// Save state - ett fel loggas men stoppar inte körningen.
		if err := o.stateStore.SaveCheckpoint(cp); err != nil {
			log.Printf("failed to save checkpoint %s: %v", cp.ID, err)
		}
	}

	// Mark plan as completed
	plan.Complete()
	if err := o.stateStore.SavePlan(plan); err != nil {
		log.Printf("failed to save plan %s: %v", plan.ID, err)
	}

	result.Status = StatusCompleted
	result.EndTime = time.Now()
	return result, nil
}

// executeCheckpoint executes a single checkpoint
func (o *Orchestrator) executeCheckpoint(ctx context.Context, cp *checkpoint.Checkpoint) (*CheckpointResult, error) {
	result := &CheckpointResult{
		CheckpointID: cp.ID,
		StartTime:    time.Now(),
	}

	// Mark as in progress
	cp.MarkInProgress()
	if err := o.stateStore.SaveCheckpoint(cp); err != nil {
		log.Printf("failed to save checkpoint %s: %v", cp.ID, err)
	}
	o.viewer.UpdateCheckpoint(cp)

	// Run gates in order
	for _, gateType := range cp.Gates {
		g := o.gates.Get(gate.GateType(gateType))
		if g == nil {
			log.Printf("Warning: gate %s not found", gateType)
			continue
		}

		gateResult, err := g.Run(ctx, cp)
		if err != nil {
			result.GateResults = append(result.GateResults, &GateResultWrapper{
				GateType: gateType,
				Passed:   false,
				Error:    err.Error(),
			})
			result.EndTime = time.Now()
			result.Passed = false
			cp.MarkFailed()
			if err := o.stateStore.SaveCheckpoint(cp); err != nil {
				log.Printf("failed to save checkpoint %s: %v", cp.ID, err)
			}
			return result, err
		}

		cp.SetGateResult(gateType, gateResult)
		if err := o.stateStore.SaveCheckpoint(cp); err != nil {
			log.Printf("failed to save checkpoint %s: %v", cp.ID, err)
		}

		result.GateResults = append(result.GateResults, &GateResultWrapper{
			GateType: gateType,
			Passed:   gateResult.Status == checkpoint.GateStatusPassed,
			Feedback: gateResult.Feedback,
			Issues:   gateResult.Issues,
		})

		// Update viewer
		o.viewer.UpdateCheckpoint(cp)

		// Handle failure - depending on gate type
		if gateResult.Status == checkpoint.GateStatusFailed {
			if gate.GateType(gateType).IsBlocking() {
				cp.MarkFailed()
				if err := o.stateStore.SaveCheckpoint(cp); err != nil {
					log.Printf("failed to save checkpoint %s: %v", cp.ID, err)
				}
				result.EndTime = time.Now()
				result.Passed = false
				return result, fmt.Errorf("gate %s failed: %s", gateType, gateResult.Feedback)
			}
		}
	}

	// All gates passed
	cp.MarkPassed()
	if err := o.stateStore.SaveCheckpoint(cp); err != nil {
		log.Printf("failed to save checkpoint %s: %v", cp.ID, err)
	}

	result.EndTime = time.Now()
	result.Passed = cp.Status == checkpoint.CheckpointPassed

	return result, nil
}

// ShouldStopHookFunc returns the ShouldStop function for AgentRuntime
func (o *Orchestrator) ShouldStopHookFunc() func(*core.TurnContext) bool {
	return o.shouldStopHook.ShouldStop
}

// GetAgentConfig returns the agent configuration including hooks
func (o *Orchestrator) GetAgentConfig() *core.AgentLoopConfig {
	return &core.AgentLoopConfig{
		ShouldStop: o.ShouldStopHookFunc(),
	}
}

// AgentFactory creates agents
type AgentFactory interface {
	CreateAgent(cfg AgentConfig) Agent
}

// AgentConfig is configuration for creating an agent
type AgentConfig struct {
	Name         string
	SystemPrompt string
	Tools        []string
}

// Agent represents an agent that can execute tasks
type Agent interface {
	Run(ctx context.Context, prompt string) (string, error)
}

// CheckpointPlanner creates checkpoint plans
type CheckpointPlanner interface {
	CreateCheckpoints(ctx context.Context, feature, description string) ([]*checkpoint.Checkpoint, error)
}

// Viewer displays checkpoint status
type Viewer interface {
	ShowPlan(plan *checkpoint.CheckpointPlan)
	ShowCheckpoint(cp *checkpoint.Checkpoint)
	UpdateCheckpoint(cp *checkpoint.Checkpoint)
}

// ExecutionResult represents the result of executing a plan
type ExecutionResult struct {
	PlanID            string
	Status            string
	StartTime         time.Time
	EndTime           time.Time
	CurrentCheckpoint int
	TotalCheckpoints  int
	CheckpointResults []*CheckpointResult
	Error             string
}

// CheckpointResult represents the result of executing a checkpoint
type CheckpointResult struct {
	CheckpointID string
	Passed       bool
	StartTime    time.Time
	EndTime      time.Time
	GateResults  []*GateResultWrapper
}

// GateResultWrapper wraps gate results for the execution result
type GateResultWrapper struct {
	GateType string
	Passed   bool
	Feedback string
	Issues   []checkpoint.Issue
	Error    string
}

// Execution statuses
const (
	StatusPendingApproval = "pending_approval"
	StatusInProgress      = "in_progress"
	StatusCompleted       = "completed"
	StatusFailed          = "failed"
)
