package orchestrator

import (
	"context"
	"time"

	"plan/agent/internal/core/gate"
	"plan/agent/internal/core/messages"
	"plan/agent/internal/core/provider"
)

// NewUIGateAgentFactory creates a gate.AgentFactory that uses the provider for LLM calls
func NewUIGateAgentFactory(provider *provider.ProviderManager, model string) gate.AgentFactory {
	return &uiGateAgentFactory{
		provider: provider,
		model:    model,
	}
}

// uiGateAgentFactory implements gate.AgentFactory for UI gate
type uiGateAgentFactory struct {
	provider *provider.ProviderManager
	model    string
}

// CreateAgent creates a gate.Agent
func (f *uiGateAgentFactory) CreateAgent(config gate.AgentConfig) gate.Agent {
	return &uiGateAgent{
		name:         config.Name,
		systemPrompt: config.SystemPrompt,
		provider:     f.provider,
		model:        f.model,
	}
}

// uiGateAgent implements gate.Agent for UI review
type uiGateAgent struct {
	name         string
	systemPrompt string
	provider     *provider.ProviderManager
	model        string
}

// Run executes the agent with the given prompt
func (a *uiGateAgent) Run(ctx context.Context, prompt string) (string, error) {
	var msgs []messages.Message

	if a.systemPrompt != "" {
		msgs = append(msgs, &messages.SystemMessage{
			Role:      messages.RoleSystem,
			Content:   []messages.ContentBlock{&messages.TextBlock{Text: a.systemPrompt}},
			Timestamp: time.Now().UnixMilli(),
		})
	}

	msgs = append(msgs, &messages.UserMessage{
		Role:      messages.RoleUser,
		Content:   []messages.ContentBlock{&messages.TextBlock{Text: prompt}},
		Timestamp: time.Now().UnixMilli(),
	})

	model := provider.Model{Name: a.model}
	if model.Name == "" {
		model.Name = "MiniMax-M2.7" // Default model
	}

	events, _, err := a.provider.Stream(ctx, model, msgs, nil, provider.StreamingOptions{})
	if err != nil {
		return "", err
	}

	var response string
	for event := range events {
		switch e := event.(type) {
		case *provider.TextDeltaEvent:
			response += e.Delta
		case *provider.DoneEvent:
			return response, nil
		case *provider.ErrorEvent:
			return response, e.Error
		}
	}

	return response, nil
}

// GateAgent wraps an existing agent to implement gate.Agent interface
type GateAgent struct {
	name         string
	systemPrompt string
	provider     *provider.ProviderManager
	model        string
}

// NewGateAgent creates a new gate agent
func NewGateAgent(name, systemPrompt string, provider *provider.ProviderManager, model string) *GateAgent {
	return &GateAgent{
		name:         name,
		systemPrompt: systemPrompt,
		provider:     provider,
		model:        model,
	}
}

// Run executes the agent
func (a *GateAgent) Run(ctx context.Context, prompt string) (string, error) {
	var msgs []messages.Message

	if a.systemPrompt != "" {
		msgs = append(msgs, &messages.SystemMessage{
			Role:      messages.RoleSystem,
			Content:   []messages.ContentBlock{&messages.TextBlock{Text: a.systemPrompt}},
			Timestamp: time.Now().UnixMilli(),
		})
	}

	msgs = append(msgs, &messages.UserMessage{
		Role:      messages.RoleUser,
		Content:   []messages.ContentBlock{&messages.TextBlock{Text: prompt}},
		Timestamp: time.Now().UnixMilli(),
	})

	model := provider.Model{Name: a.model}

	events, _, err := a.provider.Stream(ctx, model, msgs, nil, provider.StreamingOptions{})
	if err != nil {
		return "", err
	}

	var response string
	for event := range events {
		switch e := event.(type) {
		case *provider.TextDeltaEvent:
			response += e.Delta
		case *provider.DoneEvent:
			return response, nil
		case *provider.ErrorEvent:
			return response, e.Error
		}
	}

	return response, nil
}

// SimpleAgentFactory is a simple factory that creates GateAgents
type SimpleAgentFactory struct {
	provider *provider.ProviderManager
	model    string
}

// NewSimpleAgentFactory creates a new simple agent factory
func NewSimpleAgentFactory(provider *provider.ProviderManager, model string) *SimpleAgentFactory {
	return &SimpleAgentFactory{
		provider: provider,
		model:    model,
	}
}

// CreateAgent creates a gate.Agent
func (f *SimpleAgentFactory) CreateAgent(config gate.AgentConfig) gate.Agent {
	return NewGateAgent(config.Name, config.SystemPrompt, f.provider, f.model)
}
