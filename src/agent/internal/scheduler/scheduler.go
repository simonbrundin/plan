package scheduler

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"plan-agent/internal/approvals"
	"plan-agent/internal/definer"
	"plan-agent/internal/planclient"
)

// TriggerSchedule beskriver att körningen startades av schemaläggningen.
const TriggerSchedule = "schedule"

// Scheduler polls Plan API and processes goals
type Scheduler struct {
	planClient   *planclient.Client
	approvals    *approvals.Store
	definer      *definer.Definer
	pollInterval time.Duration
	planURL      string
	// initiator beskriver vem eller vad som startade körningen.
	initiator string
}

// Config contains scheduler configuration
type Config struct {
	PollInterval time.Duration
	PlanURL      string
	// Initiator beskriver vem eller vad som startade körningen.
	Initiator string
}

// New creates a new scheduler
func New(planClient *planclient.Client, approvals *approvals.Store, definer *definer.Definer, cfg Config) *Scheduler {
	interval := cfg.PollInterval
	if interval == 0 {
		interval = 1 * time.Minute
	}

	planURL := cfg.PlanURL
	if planURL == "" {
		planURL = os.Getenv("PLAN_WEB_URL")
		if planURL == "" {
			planURL = "http://localhost:3000"
		}
	}

	return &Scheduler{
		planClient:   planClient,
		approvals:    approvals,
		definer:      definer,
		pollInterval: interval,
		planURL:      planURL,
		initiator:    cfg.Initiator,
	}
}

// Run starts the scheduler (blocks forever)
func (s *Scheduler) Run(ctx context.Context) {
	log.Printf("Scheduler starting, polling every %s for agent=true goals in define", s.pollInterval)

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	// Run once immediately
	if err := s.runOnce(ctx); err != nil {
		log.Printf("scheduler run failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler stopping")
			return
		case <-ticker.C:
			if err := s.runOnce(ctx); err != nil {
				log.Printf("scheduler run failed: %v", err)
			}
		}
	}
}

// RunOnce processes goals once (useful for testing)
func (s *Scheduler) RunOnce(ctx context.Context) error {
	return s.runOnce(ctx)
}

func (s *Scheduler) runOnce(ctx context.Context) error {
	goals, err := s.planClient.GetAgentGoals(ctx)
	if err != nil {
		log.Printf("Failed to fetch agent goals: %v", err)
		return err
	}

	defineStatusID, err := s.planClient.GetStatusIDByName(ctx, "define")
	if err != nil {
		log.Printf("Failed to resolve define phase: %v", err)
		return err
	}

	log.Printf("Found %d agent goals", len(goals))

	for _, goal := range goals {
		if !isAgentDefineGoal(goal, defineStatusID) {
			continue
		}
		s.processGoal(ctx, goal)
	}

	return nil
}

func isAgentDefineGoal(goal planclient.Goal, defineStatusID int) bool {
	return goal.Agent && goal.StatusID == defineStatusID
}

func (s *Scheduler) processGoal(ctx context.Context, goal planclient.Goal) {
	log.Printf("Processing goal #%d: %s", goal.ID, goal.Title)

	// Check if there's already a pending approval for this goal
	existing, err := s.approvals.Get(goal.ID)
	if err != nil {
		log.Printf("Failed to check existing approval for goal #%d: %v", goal.ID, err)
		return
	}

	// Get feedback if any
	feedback := ""
	attempts := 1
	if existing != nil {
		feedback = existing.Feedback
		attempts = existing.Attempts + 1
	}

	// Define the goal
	definition, err := s.definer.Define(ctx, goal.Title, goal.Title, feedback)
	if err != nil {
		log.Printf("Failed to define goal #%d: %v", goal.ID, err)
		s.publish(ctx, goal, "failed", fmt.Sprintf("Agenten kunde inte specificera målet: %v", err))
		return
	}

	// Save pending approval
	approval := &approvals.PendingApproval{
		GoalID:      goal.ID,
		Title:       definition.Title,
		Description: definition.Description,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		Attempts:    attempts,
		Feedback:    feedback,
		Status:      approvals.StatusPending,
	}

	if err := s.approvals.Save(approval); err != nil {
		log.Printf("Failed to save approval for goal #%d: %v", goal.ID, err)
		return
	}

	log.Printf("Saved pending approval for goal #%d: %s", goal.ID, definition.Title)

	// Förslaget syns i Plan - där tas beslutet. Slack används inte.
	s.publish(ctx, goal, planclient.RunStatusWaitingApproval, fmt.Sprintf("Förslag: %s\n\n%s", definition.Title, definition.Description))
	s.recordApprovalRequest(ctx, goal, definition.Title)
}

// recordApprovalRequest writes the request for a human decision to the run
// history, so it is clear that a person - not the agent - has to decide.
func (s *Scheduler) recordApprovalRequest(ctx context.Context, goal planclient.Goal, title string) {
	runs, err := s.planClient.GetAgentRuns(ctx, &goal.ID)
	if err != nil {
		log.Printf("Failed to list agent runs for goal #%d: %v", goal.ID, err)
		return
	}
	if len(runs) == 0 {
		return
	}

	err = s.planClient.CreateAgentRunEvent(ctx, runs[0].ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventApprovalRequested,
		Phase: "define",
		Payload: map[string]any{
			"title": title,
		},
	})
	if err != nil {
		log.Printf("Failed to record approval request for goal #%d: %v", goal.ID, err)
	}
}

// publish creates or updates the agent run in Plan so the proposal is
// visible where the goal lives.
func (s *Scheduler) publish(ctx context.Context, goal planclient.Goal, status, summary string) {
	runs, err := s.planClient.GetAgentRuns(ctx, &goal.ID)
	if err != nil {
		log.Printf("Failed to list agent runs for goal #%d: %v", goal.ID, err)
		return
	}

	if len(runs) > 0 {
		run := runs[0]
		phase := run.Phase
		req := planclient.UpdateAgentRunRequest{
			Phase:   &phase,
			Status:  &status,
			Summary: &summary,
		}
		if _, err := s.planClient.UpdateAgentRun(ctx, run.ID, req); err != nil {
			log.Printf("Failed to update agent run %d: %v", run.ID, err)
		}
		return
	}

	req := planclient.CreateAgentRunRequest{
		GoalID:    goal.ID,
		Phase:     "define",
		Status:    status,
		Trigger:   TriggerSchedule,
		Initiator: s.initiator,
		Summary:   summary,
	}
	if _, err := s.planClient.CreateAgentRun(ctx, req); err != nil {
		log.Printf("Failed to create agent run for goal #%d: %v", goal.ID, err)
	}
}
