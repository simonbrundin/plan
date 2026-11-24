package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

// SubAgentConfig is configuration for spawning a sub-agent
type SubAgentConfig struct {
	Name         string        // Name for logging
	SystemPrompt string        // Custom system prompt for this agent
	Tools        []string      // Tool names to enable
	MaxRetries   int           // Max retries on failure
	Timeout      time.Duration // Timeout for execution
}

// SubAgentResult is the result of a sub-agent execution
type SubAgentResult struct {
	Name     string
	Output   string
	Error    error
	Duration time.Duration
}

// SubAgentManager manages spawning and execution of sub-agents
type SubAgentManager struct {
	provider *provider.ProviderManager
	model    string
	mu       sync.RWMutex
	agents   map[string]Agent
}

// NewSubAgentManager creates a new sub-agent manager
func NewSubAgentManager(provider *provider.ProviderManager, model string) *SubAgentManager {
	return &SubAgentManager{
		provider: provider,
		model:    model,
		agents:   make(map[string]Agent),
	}
}

// Spawn creates a new sub-agent with fresh context
func (m *SubAgentManager) Spawn(ctx context.Context, config SubAgentConfig) (Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent := &FreshContextAgent{
		name:         config.Name,
		provider:     m.provider,
		model:        m.model,
		systemPrompt: config.SystemPrompt,
		maxRetries:   config.MaxRetries,
		timeout:      config.Timeout,
	}

	m.agents[config.Name] = agent
	return agent, nil
}

// SpawnAndRun spawns a sub-agent and runs it immediately
func (m *SubAgentManager) SpawnAndRun(ctx context.Context, config SubAgentConfig, prompt string) *SubAgentResult {
	start := time.Now()

	agent, err := m.Spawn(ctx, config)
	if err != nil {
		return &SubAgentResult{
			Name:     config.Name,
			Error:    err,
			Duration: time.Since(start),
		}
	}

	output, err := agent.Run(ctx, prompt)
	return &SubAgentResult{
		Name:     config.Name,
		Output:   output,
		Error:    err,
		Duration: time.Since(start),
	}
}

// SpawnMultiple creates multiple sub-agents in parallel and runs them
func (m *SubAgentManager) SpawnMultiple(ctx context.Context, configs []SubAgentConfig, prompts []string) []*SubAgentResult {
	if len(configs) != len(prompts) {
		return nil
	}

	results := make([]*SubAgentResult, len(configs))
	var wg sync.WaitGroup

	for i, config := range configs {
		wg.Add(1)
		go func(idx int, cfg SubAgentConfig, prompt string) {
			defer wg.Done()
			results[idx] = m.SpawnAndRun(ctx, cfg, prompt)
		}(i, config, prompts[i])
	}

	wg.Wait()
	return results
}

// FreshContextAgent is an agent that runs with a fresh context (no conversation history)
type FreshContextAgent struct {
	name         string
	provider     *provider.ProviderManager
	model        string
	systemPrompt string
	maxRetries   int
	timeout      time.Duration
}

// Run executes the agent with a fresh context (no accumulated history)
func (a *FreshContextAgent) Run(ctx context.Context, prompt string) (string, error) {
	// Build messages with system prompt if provided
	var msgs []messages.Message

	if a.systemPrompt != "" {
		msgs = append(msgs, &messages.SystemMessage{
			Role:      messages.RoleSystem,
			Content:   []messages.ContentBlock{&messages.TextBlock{Text: a.systemPrompt}},
			Timestamp: time.Now().UnixMilli(),
		})
	}

	// Add user message
	msgs = append(msgs, &messages.UserMessage{
		Role:      messages.RoleUser,
		Content:   []messages.ContentBlock{&messages.TextBlock{Text: prompt}},
		Timestamp: time.Now().UnixMilli(),
	})

	// Set timeout if not already set
	if a.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.timeout)
		defer cancel()
	}

	model := provider.Model{Name: a.model}
	maxRetries := a.maxRetries
	if maxRetries <= 0 {
		maxRetries = 2
	}

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		events, _, err := a.provider.Stream(ctx, model, msgs, nil, provider.StreamingOptions{})
		if err != nil {
			lastErr = err
			log.Printf("SubAgent %s: attempt %d failed: %v", a.name, attempt+1, err)
			continue
		}

		// Collect response
		var response string
		for event := range events {
			switch e := event.(type) {
			case *provider.TextDeltaEvent:
				response += e.Delta
			case *provider.DoneEvent:
				return response, nil
			case *provider.ErrorEvent:
				lastErr = e.Error
			}
		}

		if lastErr != nil {
			continue
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("sub-agent %s failed after %d attempts: %w", a.name, maxRetries+1, lastErr)
	}

	return "", nil
}

// SubAgentPool manages a pool of reusable agents
type SubAgentPool struct {
	manager   *SubAgentManager
	maxAgents int
	available chan Agent
}

// NewSubAgentPool creates a pool of pre-warmed agents
func NewSubAgentPool(provider *provider.ProviderManager, model string, maxAgents int) *SubAgentPool {
	pool := &SubAgentPool{
		manager:   NewSubAgentManager(provider, model),
		maxAgents: maxAgents,
		available: make(chan Agent, maxAgents),
	}

	// Pre-warm the pool
	for i := 0; i < maxAgents; i++ {
		agent, _ := pool.manager.Spawn(context.Background(), SubAgentConfig{
			Name: fmt.Sprintf("pool-%d", i),
		})
		pool.available <- agent
	}

	return pool
}

// Acquire gets an agent from the pool
func (p *SubAgentPool) Acquire(ctx context.Context) (Agent, error) {
	select {
	case agent := <-p.available:
		return agent, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Pool exhausted, spawn a new one
		return p.manager.Spawn(ctx, SubAgentConfig{
			Name: fmt.Sprintf("overflow-%d", time.Now().UnixNano()),
		})
	}
}

// Release returns an agent to the pool
func (p *SubAgentPool) Release(agent Agent) {
	select {
	case p.available <- agent:
		// Returned to pool
	default:
		// Pool full, discard
	}
}

// Close shuts down the pool
func (p *SubAgentPool) Close() {
	close(p.available)
}

// AgentNames returns names of spawned agents for debugging
func (m *SubAgentManager) AgentNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.agents))
	for name := range m.agents {
		names = append(names, name)
	}
	return names
}
