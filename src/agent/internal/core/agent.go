package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
	"plan-agent/internal/core/tools"
)

// ToolCallCancellation provides cancellation for tool executions
type ToolCallCancellation struct {
	manager *tools.CancellationManager
}

// NewToolCallCancellation creates a new tool call cancellation manager
func NewToolCallCancellation() *ToolCallCancellation {
	return &ToolCallCancellation{
		manager: tools.GetGlobalCancellationManager(),
	}
}

// Register registers a cancel channel for a tool call, returns the signal channel
func (t *ToolCallCancellation) Register(toolCallID string) <-chan struct{} {
	return t.manager.Register(toolCallID)
}

// Cancel signals cancellation for a tool call
func (t *ToolCallCancellation) Cancel(toolCallID string) {
	t.manager.Cancel(toolCallID)
}

// Unregister removes the cancel channel for a tool call
func (t *ToolCallCancellation) Unregister(toolCallID string) {
	t.manager.Unregister(toolCallID)
}

// CancelAll cancels all running tool executions
func (t *ToolCallCancellation) CancelAll() {
	t.manager.CancelAll()
}

// Event types emitted by the agent
type EventType string

const (
	EventAgentStart    EventType = "agent_start"
	EventAgentEnd      EventType = "agent_end"
	EventTurnStart     EventType = "turn_start"
	EventTurnEnd       EventType = "turn_end"
	EventMessageStart  EventType = "message_start"
	EventMessageEnd    EventType = "message_end"
	EventMessageUpdate EventType = "message_update"
	EventToolStart     EventType = "tool_execution_start"
	EventToolEnd       EventType = "tool_execution_end"
	EventToolUpdate    EventType = "tool_execution_update"

	StopReasonEndTurn      = "end_turn"
	StopReasonStopSequence = "stop_sequence"
)

// AgentEvent represents an event emitted by the agent
type AgentEvent struct {
	Type        EventType                     `json:"type"`
	Message     *messages.AssistantMessage    `json:"message,omitempty"`
	ToolCallID  string                        `json:"toolCallId,omitempty"`
	ToolName    string                        `json:"toolName,omitempty"`
	Args        map[string]any                `json:"args,omitempty"`
	Result      *provider.ToolResult          `json:"result,omitempty"`
	IsError     bool                          `json:"isError,omitempty"`
	ToolResults []*messages.ToolResultMessage `json:"toolResults,omitempty"`
	Messages    []messages.Message            `json:"messages,omitempty"`
}

// EventHandler is called for each agent event
type EventHandler func(event *AgentEvent)

// AgentLoopConfig contains the configuration for the agent loop
type AgentLoopConfig struct {
	Model           provider.Model
	Tools           []*tools.ToolDefinition
	ProviderTools   []provider.ToolDefinition // pre-converted tools, populated by NewAgent
	SystemPrompt    string
	Provider        provider.LLMProvider
	ProviderManager *provider.ProviderManager // optional fallback manager

	// Hooks for customization
	ShouldStop      func(ctx *TurnContext) bool
	PrepareNextTurn func(ctx *TurnContext) *TurnUpdate

	// OnTurnComplete is called after each turn completes (useful for metrics/usage tracking)
	OnTurnComplete func(msg *messages.AssistantMessage, providerName string)

	// Tool execution mode
	ToolExecution provider.ToolExecutionMode

	// MaxTurns limits how many model turns a single request may consume.
	// A limit prevents a confused model from looping indefinitely through tools.
	MaxTurns int
}

// TurnContext contains context for turn-level hooks
type TurnContext struct {
	Message     *messages.AssistantMessage
	ToolResults []*messages.ToolResultMessage
	Context     *AgentContext
}

// TurnUpdate can modify the next turn
type TurnUpdate struct {
	Context  *AgentContext
	Model    *provider.Model
	Thinking *provider.ThinkingLevel
}

// AgentContext holds the current agent state
type AgentContext struct {
	SystemPrompt string
	Messages     []messages.Message
	Tools        []*tools.ToolDefinition
}

// RunResult contains the result of running the agent
type RunResult struct {
	Messages []messages.Message
	Error    error
}

// Agent represents the agent runtime
type AgentRuntime struct {
	config          *AgentLoopConfig
	currentProvider string // Track which provider was used for the last turn
}

// Model returns the current model configuration
func (a *AgentRuntime) Model() provider.Model {
	return a.config.Model
}

// NewAgent creates a new agent instance
func NewAgent(config *AgentLoopConfig) *AgentRuntime {
	if config.ToolExecution == "" {
		config.ToolExecution = provider.ToolExecutionParallel
	}
	// Pre-convert tools once at startup
	if len(config.Tools) > 0 && len(config.ProviderTools) == 0 {
		config.ProviderTools = convertToolDefsToProvider(config.Tools)
	}
	return &AgentRuntime{
		config: config,
	}
}

// SendMessage sends a single user message to the agent with no conversation history.
// For multi-turn conversations, use SendMessageWithHistory instead.
func (a *AgentRuntime) SendMessage(ctx context.Context, prompt string, handler EventHandler) *RunResult {
	return a.SendMessageWithHistory(ctx, nil, prompt, handler)
}

// SendMessageWithHistory sends a user message to the agent, preserving conversation history.
// The history should contain all previous messages; the new user message is appended.
// Returns the updated messages slice including the new exchange.
func (a *AgentRuntime) SendMessageWithHistory(ctx context.Context, history []messages.Message, prompt string, handler EventHandler) *RunResult {
	newMessages := make([]messages.Message, 0, len(history)+1)
	if history != nil {
		newMessages = append(newMessages, history...)
	}
	newMessages = append(newMessages, &messages.UserMessage{
		Role:      messages.RoleUser,
		Content:   []messages.ContentBlock{&messages.TextBlock{Text: prompt}},
		Timestamp: time.Now().UnixMilli(),
	})
	return a.runWithMessages(ctx, newMessages, handler)
}

// runWithMessages runs the agent with specific messages
func (a *AgentRuntime) runWithMessages(ctx context.Context, newMessages []messages.Message, handler EventHandler) *RunResult {
	agentCtx := &AgentContext{
		SystemPrompt: a.config.SystemPrompt,
		Messages:     newMessages,
		Tools:        a.config.Tools,
	}

	// Add system prompt if provided
	if agentCtx.SystemPrompt != "" {
		agentCtx.Messages = append([]messages.Message{
			&messages.SystemMessage{
				Role:      messages.RoleSystem,
				Content:   []messages.ContentBlock{&messages.TextBlock{Text: agentCtx.SystemPrompt}},
				Timestamp: time.Now().UnixMilli(),
			},
		}, agentCtx.Messages...)
	}

	handler(&AgentEvent{Type: EventAgentStart})

	maxTurns := a.config.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 20
	}
	stoppedNormally := false

	for turn := 0; turn < maxTurns; turn++ {
		handler(&AgentEvent{Type: EventTurnStart})

		// Stream response
		msg, err := a.streamResponse(ctx, agentCtx, handler)
		if err != nil {
			handler(&AgentEvent{Type: EventAgentEnd, Messages: agentCtx.Messages})
			return &RunResult{Messages: agentCtx.Messages, Error: err}
		}

		agentCtx.Messages = append(agentCtx.Messages, msg)

		// Check for stop - but first verify checkpoint gates allow stopping
		// This implements the Helix "exit code 2" pattern: the agent can only stop
		// when ShouldStop hook returns true (all gates passed)
		if msg.StopReason == StopReasonEndTurn || msg.StopReason == StopReasonStopSequence {
			// Build TurnContext for gate checking (no tools were executed)
			turnCtx := &TurnContext{
				Message:     msg,
				ToolResults: nil,
				Context:     agentCtx,
			}

			// Check ShouldStop - this runs checkpoint gates
			// If ShouldStop is nil, gates are not configured → allow stop
			// If ShouldStop returns true → gates passed → allow stop
			// If ShouldStop returns false → gates not passed → BLOCK stop
			if a.config.ShouldStop != nil && !a.config.ShouldStop(turnCtx) {
				// Gates blocked stopping - continue working
				// Emit turn end event so agent knows it tried to stop
				handler(&AgentEvent{
					Type:        EventTurnEnd,
					Message:     msg,
					ToolResults: nil,
				})
				continue
			}

			// ShouldStop returned true (or wasn't set) - allow stopping
			stoppedNormally = true
			break
		}

		// Execute tool calls
		toolResults := a.executeToolCalls(ctx, agentCtx, msg, handler)

		for _, tr := range toolResults {
			agentCtx.Messages = append(agentCtx.Messages, tr)
		}

		// Turn-level hooks
		turnCtx := &TurnContext{
			Message:     msg,
			ToolResults: toolResults,
			Context:     agentCtx,
		}

		if a.config.ShouldStop != nil && a.config.ShouldStop(turnCtx) {
			stoppedNormally = true
			break
		}

		if a.config.PrepareNextTurn != nil {
			update := a.config.PrepareNextTurn(turnCtx)
			if update != nil {
				if update.Context != nil {
					agentCtx = update.Context
				}
				if update.Model != nil {
					a.config.Model = *update.Model
				}
			}
		}

		handler(&AgentEvent{
			Type:        EventTurnEnd,
			Message:     msg,
			ToolResults: toolResults,
		})

		// Call OnTurnComplete hook if configured
		if a.config.OnTurnComplete != nil {
			a.config.OnTurnComplete(msg, a.currentProvider)
		}
	}

	handler(&AgentEvent{Type: EventAgentEnd, Messages: agentCtx.Messages})

	if stoppedNormally {
		return &RunResult{Messages: agentCtx.Messages}
	}
	return &RunResult{
		Messages: agentCtx.Messages,
		Error:    fmt.Errorf("agent stopped after reaching the maximum of %d turns", maxTurns),
	}
}

// streamResponse streams an LLM response
func (a *AgentRuntime) streamResponse(
	ctx context.Context,
	agentCtx *AgentContext,
	handler EventHandler,
) (*messages.AssistantMessage, error) {

	var stream <-chan provider.StreamEvent
	var providerName string
	var err error

	// Use provider manager for automatic fallback if available
	if a.config.ProviderManager != nil {
		stream, providerName, err = a.config.ProviderManager.Stream(
			ctx,
			a.config.Model,
			agentCtx.Messages,
			a.config.ProviderTools,
			provider.StreamingOptions{
				SystemPrompt: agentCtx.SystemPrompt,
			},
		)
	} else if a.config.Provider != nil {
		stream, err = a.config.Provider.Stream(
			ctx,
			a.config.Model,
			agentCtx.Messages,
			a.config.ProviderTools,
			provider.StreamingOptions{
				SystemPrompt: agentCtx.SystemPrompt,
			},
		)
		providerName = "default"
	}

	if err != nil {
		return nil, fmt.Errorf("stream error: %w", err)
	}

	// Log which provider was used for this stream
	if providerName != "" {
		a.currentProvider = providerName
		log.Printf("Agent using provider: %s", providerName)
	}

	msg := &messages.AssistantMessage{
		Role:      messages.RoleAssistant,
		Content:   make([]messages.ContentBlock, 0),
		Timestamp: time.Now().UnixMilli(),
	}

	for event := range stream {
		switch e := event.(type) {
		case *provider.StartEvent:
			msg = e.Partial
			handler(&AgentEvent{Type: EventMessageStart, Message: msg})

		case *provider.TextDeltaEvent:
			msg.Content = append(msg.Content, &messages.TextBlock{Text: e.Delta})

		case *provider.ToolCallDeltaEvent:
			// Handle streaming tool call args
			if tc := findToolCall(msg, e.ID); tc != nil {
				if tc.Arguments == nil {
					tc.Arguments = make(map[string]any)
				}
				tc.Arguments["_raw_args"] = e.Arguments
			}

		case *provider.DoneEvent:
			msg = e.Message
			handler(&AgentEvent{Type: EventMessageEnd, Message: msg})
			return msg, nil

		case *provider.ErrorEvent:
			handler(&AgentEvent{Type: EventMessageEnd, Message: msg})
			return msg, e.Error
		}
	}

	return msg, nil
}

// executeToolCalls executes tool calls from an assistant message
func (a *AgentRuntime) executeToolCalls(
	_ context.Context,
	agentCtx *AgentContext,
	msg *messages.AssistantMessage,
	handler EventHandler,
) []*messages.ToolResultMessage {

	var results []*messages.ToolResultMessage

	for _, block := range msg.Content {
		tc, ok := block.(*messages.ToolCallBlock)
		if !ok {
			continue
		}

		results = append(results, a.executeSingleToolCall(tc, agentCtx, handler))
	}

	return results
}

func (a *AgentRuntime) executeSingleToolCall(
	tc *messages.ToolCallBlock,
	agentCtx *AgentContext,
	handler EventHandler,
) *messages.ToolResultMessage {
	handler(&AgentEvent{
		Type:       EventToolStart,
		ToolCallID: tc.ID,
		ToolName:   tc.Name,
		Args:       tc.Arguments,
	})

	// Register cancellation channel for this tool call
	signal := tools.GetGlobalCancellationManager().Register(tc.ID)
	defer tools.GetGlobalCancellationManager().Unregister(tc.ID)

	toolResult := a.runToolOrError(tc, agentCtx, signal, newToolUpdateHandler(tc, handler))

	handler(&AgentEvent{
		Type:       EventToolEnd,
		ToolCallID: tc.ID,
		ToolName:   tc.Name,
		Result:     toolResult,
		IsError:    toolResult.IsError,
	})

	return &messages.ToolResultMessage{
		Role:       messages.RoleToolResult,
		ToolCallID: tc.ID,
		ToolName:   tc.Name,
		Content:    toolResult.Content,
		IsError:    toolResult.IsError,
		Details:    toolResult.Details,
		Timestamp:  time.Now().UnixMilli(),
	}
}

func (a *AgentRuntime) runToolOrError(
	tc *messages.ToolCallBlock,
	agentCtx *AgentContext,
	signal <-chan struct{},
	onToolUpdate func(partial *provider.ToolResult),
) *provider.ToolResult {
	tool := findTool(agentCtx.Tools, tc.Name)
	if tool == nil {
		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{
				Text: fmt.Sprintf("Tool '%s' not found", tc.Name),
			}},
			IsError: true,
		}
	}

	args := provider.ParseToolCallArguments(tc)

	if err := tools.ValidateArguments(args, tool.InputSchema); err != nil {
		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: err.Error()}},
			IsError: true,
		}
	}

	result, err := tool.Execute(tc.ID, args, signal, onToolUpdate)
	if err != nil {
		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: err.Error()}},
			IsError: true,
		}
	}

	return result
}

func newToolUpdateHandler(tc *messages.ToolCallBlock, handler EventHandler) func(partial *provider.ToolResult) {
	return func(partial *provider.ToolResult) {
		handler(&AgentEvent{
			Type:       EventToolUpdate,
			ToolCallID: tc.ID,
			ToolName:   tc.Name,
			Result:     partial,
		})
	}
}

func findTool(tools []*tools.ToolDefinition, name string) *tools.ToolDefinition {
	for _, t := range tools {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// Helper functions

func convertToolDefsToProvider(tDefs []*tools.ToolDefinition) []provider.ToolDefinition {
	result := make([]provider.ToolDefinition, len(tDefs))
	for i, t := range tDefs {
		result[i] = t.ToProviderFormat()
	}
	return result
}

func findToolCall(msg *messages.AssistantMessage, id string) *messages.ToolCallBlock {
	for _, block := range msg.Content {
		if tc, ok := block.(*messages.ToolCallBlock); ok && tc.ID == id {
			return tc
		}
	}
	return nil
}
