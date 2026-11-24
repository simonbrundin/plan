package orchestrator

import (
	"context"
	"time"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

// PlannerAgent implements the Agent interface using a PlannerProvider (ProviderManager)
type PlannerAgent struct {
	provider *provider.ProviderManager
	model    string
}

// NewPlannerAgent creates a new planner agent
func NewPlannerAgent(provider *provider.ProviderManager, model string) *PlannerAgent {
	if model == "" {
		model = "MiniMax-M2.7" // default model for planning
	}
	return &PlannerAgent{
		provider: provider,
		model:    model,
	}
}

// Run executes the planner with the given prompt
func (a *PlannerAgent) Run(ctx context.Context, prompt string) (string, error) {
	// Wrap the prompt in a user message
	msgs := []messages.Message{
		&messages.UserMessage{
			Role:      messages.RoleUser,
			Content:   []messages.ContentBlock{&messages.TextBlock{Text: prompt}},
			Timestamp: time.Now().UnixMilli(),
		},
	}

	// Use the provider to get a completion
	model := provider.Model{Name: a.model}
	return a.provider.Complete(ctx, model, msgs)
}
