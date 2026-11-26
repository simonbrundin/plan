package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"plan/agent/internal/core/messages"
)

// Tool execution modes
type ToolExecutionMode string

const (
	ToolExecutionSequential ToolExecutionMode = "sequential"
	ToolExecutionParallel   ToolExecutionMode = "parallel"
)

// ContentBlock is a content block at the provider level
type ContentBlock interface {
	GetBlockType() string
}

// TextBlock at provider level
type TextBlock struct {
	Text string `json:"text"`
}

func (t *TextBlock) GetBlockType() string { return "text" }

// ToolCallBlock at provider level
type ToolCallBlock struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func (t *ToolCallBlock) GetBlockType() string { return "toolCall" }

// Model represents an LLM model configuration
type Model struct {
	Provider   string         // e.g., "openai", "anthropic"
	Name       string         // e.g., "gpt-4", "claude-3-opus"
	APIKey     string         // API key for the provider
	Endpoint   string         // Custom endpoint (optional)
	Parameters map[string]any // Model-specific parameters
}

// StreamingOptions controls streaming behavior
type StreamingOptions struct {
	Temperature  *float64
	MaxTokens    *int
	Thinking     *ThinkingLevel
	SystemPrompt string
	// Usage ber om att tokenanvändningen ska följa med svaret.
	//
	// Vissa leverantörer ger inte användningen i en strömmad svar, så
	// då används ett vanligt svar i stället. Det kostar lite mer
	// fördröjning men ger riktiga siffror att räkna kostnad med.
	Usage bool
}

// ThinkingLevel controls reasoning effort
type ThinkingLevel string

const (
	ThinkingOff     ThinkingLevel = "off"
	ThinkingMinimal ThinkingLevel = "minimal"
	ThinkingLow     ThinkingLevel = "low"
	ThinkingMedium  ThinkingLevel = "medium"
	ThinkingHigh    ThinkingLevel = "high"
	ThinkingXHigh   ThinkingLevel = "xhigh"
	ThinkingMax     ThinkingLevel = "max"
)

// StreamEvent represents a chunk from the LLM stream
type StreamEvent interface {
	GetEventType() string
}

// Event types
type (
	StartEvent struct {
		Partial *messages.AssistantMessage
	}
	TextDeltaEvent struct {
		Delta   string
		Partial *messages.AssistantMessage
	}
	ToolCallDeltaEvent struct {
		ID        string
		Name      string
		Arguments string
		Partial   *messages.AssistantMessage
	}
	DoneEvent struct {
		Message *messages.AssistantMessage
	}
	ErrorEvent struct {
		Error   error
		Message *messages.AssistantMessage
	}
)

func (e *StartEvent) GetEventType() string         { return "start" }
func (e *TextDeltaEvent) GetEventType() string     { return "text_delta" }
func (e *ToolCallDeltaEvent) GetEventType() string { return "toolcall_delta" }
func (e *DoneEvent) GetEventType() string          { return "done" }
func (e *ErrorEvent) GetEventType() string         { return "error" }

// LLMProvider defines the interface for LLM providers
type LLMProvider interface {
	// Stream streams a completion from the LLM
	Stream(ctx context.Context, model Model, messages []messages.Message, tools []ToolDefinition, opts StreamingOptions) (<-chan StreamEvent, error)

	// Name returns the provider name
	Name() string
}

// ToolDefinition defines a tool for the LLM
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content   []messages.ContentBlock
	Details   any
	IsError   bool
	Usage     *messages.Usage
	Terminate bool
}

// --- Shared conversion helpers ---

// ConvertTextContent converts text content blocks to a string
func ConvertTextContent(blocks []messages.ContentBlock) string {
	var sb bytes.Buffer
	for _, block := range blocks {
		if tb, ok := block.(*messages.TextBlock); ok {
			sb.WriteString(tb.Text)
		}
	}
	return sb.String()
}

// ConvertAssistantContent converts assistant message content for OpenAI-compatible APIs
func ConvertAssistantContent(blocks []messages.ContentBlock) (textContent any, toolCalls []map[string]any) {
	for _, block := range blocks {
		switch tb := block.(type) {
		case *messages.TextBlock:
			textContent = joinText(textContent, tb.Text)
		case *messages.ToolCallBlock:
			toolCalls = append(toolCalls, convertToolCall(tb))
		}
	}
	return textContent, toolCalls
}

func joinText(existing any, newText string) any {
	if existing == nil {
		return newText
	}
	if str, ok := existing.(string); ok {
		return str + newText
	}
	return newText
}

func convertToolCall(tc *messages.ToolCallBlock) map[string]any {
	argsStr := formatArgs(tc.Arguments)
	return map[string]any{
		"type": "function",
		"id":   tc.ID,
		"function": map[string]any{
			"name":      tc.Name,
			"arguments": argsStr,
		},
	}
}

func formatArgs(args map[string]any) string {
	if args == nil {
		return "{}"
	}
	if argsBytes, err := json.Marshal(args); err == nil {
		return string(argsBytes)
	}
	return "{}"
}

// ParseJSONArgs parses JSON arguments from a tool call
func ParseJSONArgs(args string) map[string]any {
	if args == "" || args == "{}" {
		return make(map[string]any)
	}
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(args), &result); err != nil {
		return map[string]any{"_raw": args}
	}
	return result
}

// ExtractToolCall extracts a ToolCallBlock from a map
// For streaming, it accumulates arguments in _raw_args
func ExtractToolCall(tc map[string]any) *messages.ToolCallBlock {
	index, _ := tc["index"].(float64)
	id, _ := tc["id"].(string)
	name, _ := tc["function"].(map[string]any)["name"].(string)
	args, _ := tc["function"].(map[string]any)["arguments"].(string)

	if name == "" {
		return nil
	}

	if id == "" {
		id = fmt.Sprintf("call_%d_%s", int(index), name)
	}

	return &messages.ToolCallBlock{
		ID:        id,
		Name:      name,
		Arguments: map[string]any{"_raw_args": args}, // Store raw args for streaming
	}
}

// AccumulateToolCallArgs accumulates streaming arguments for a tool call
func AccumulateToolCallArgs(tc *messages.ToolCallBlock, newArgs string) {
	if tc == nil {
		return
	}
	if existing, ok := tc.Arguments["_raw_args"].(string); ok {
		tc.Arguments["_raw_args"] = existing + newArgs
	} else {
		tc.Arguments["_raw_args"] = newArgs
	}
}

// ParseToolCallArguments parses the accumulated raw arguments
func ParseToolCallArguments(tc *messages.ToolCallBlock) map[string]any {
	if tc == nil {
		return make(map[string]any)
	}
	rawArgs, ok := tc.Arguments["_raw_args"].(string)
	if !ok {
		// No raw args stored, return existing
		return tc.Arguments
	}
	// Parse the complete JSON
	result := ParseJSONArgs(rawArgs)
	if _, hasRaw := result["_raw"]; hasRaw {
		// JSON was incomplete, return what we have
		return result
	}
	// Replace _raw_args with parsed result
	tc.Arguments = result
	return result
}

// --- Shared streaming helpers ---

// newAssistantMessage creates a fresh partial assistant message for streaming
func newAssistantMessage() *messages.AssistantMessage {
	return &messages.AssistantMessage{
		Role:      messages.RoleAssistant,
		Content:   make([]messages.ContentBlock, 0),
		Timestamp: time.Now().UnixMilli(),
	}
}

// convertMessages converts agent messages to OpenAI-compatible format
func convertMessages(msgs []messages.Message) []map[string]any {
	result := make([]map[string]any, 0, len(msgs))
	for _, m := range msgs {
		msgMap := messageToMap(m)
		if msgMap != nil {
			result = append(result, msgMap)
		}
	}
	return result
}

func messageToMap(m messages.Message) map[string]any {
	switch msg := m.(type) {
	case *messages.UserMessage:
		content := ConvertTextContent(msg.Content)
		if len(content) > 0 {
			return map[string]any{"role": "user", "content": content}
		}
	case *messages.AssistantMessage:
		textContent, toolCalls := ConvertAssistantContent(msg.Content)
		msgMap := map[string]any{"role": "assistant"}
		if textContent != nil {
			msgMap["content"] = textContent
		}
		if len(toolCalls) > 0 {
			msgMap["tool_calls"] = toolCalls
		}
		return msgMap
	case *messages.ToolResultMessage:
		toolContent := ""
		for _, c := range msg.Content {
			if tb, ok := c.(*messages.TextBlock); ok {
				toolContent = tb.Text
				break
			}
		}
		return map[string]any{
			"role":         "tool",
			"tool_call_id": msg.ToolCallID,
			"content":      toolContent,
		}
	case *messages.SystemMessage:
		content := ConvertTextContent(msg.Content)
		if len(content) > 0 {
			return map[string]any{"role": "system", "content": content}
		}
	}
	return nil
}

// toolDefsToOpenAIFormat converts tool definitions to OpenAI tools format
func toolDefsToOpenAIFormat(toolDefs []ToolDefinition) []map[string]any {
	tools := make([]map[string]any, len(toolDefs))
	for i, t := range toolDefs {
		// Ett verktyg utan schema måste ändå ha ett objekt-schema -
		// MiniMax avvisar "parameters": null.
		schema := t.InputSchema
		if schema == nil {
			schema = map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			}
		}

		tools[i] = map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  schema,
			},
		}
	}
	return tools
}

// handleToolCallDelta processes a streaming tool call delta
// For MiniMax, this accumulates args in a pending map; for OpenAI it adds directly to partialMsg
func handleToolCallDelta(tcMap map[string]any, partialMsg *messages.AssistantMessage) *messages.ToolCallBlock {
	tcBlock := ExtractToolCall(tcMap)
	if tcBlock == nil {
		return nil
	}

	// Check if this tool call already exists
	for _, block := range partialMsg.Content {
		if existingTC, ok := block.(*messages.ToolCallBlock); ok && existingTC.ID == tcBlock.ID {
			// Accumulate streaming arguments
			if args, ok := tcBlock.Arguments["_raw_args"].(string); ok {
				AccumulateToolCallArgs(existingTC, args)
			}
			return existingTC
		}
	}

	// New tool call
	partialMsg.Content = append(partialMsg.Content, tcBlock)
	return tcBlock
}

// --- Helper functions ---

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	haystack := []byte(s)
	needle := []byte(substr)
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if string(haystack[i:i+len(needle)]) == string(needle) {
			return true
		}
	}
	return false
}

// --- Retry with Exponential Backoff ---

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxAttempts int           // Maximum number of attempts (default: 3)
	BaseDelay   time.Duration // Initial delay between retries (default: 1s)
	MaxDelay    time.Duration // Maximum delay cap (default: 30s)
	Multiplier  float64       // Exponential multiplier (default: 2.0)
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Second,
		MaxDelay:    30 * time.Second,
		Multiplier:  2.0,
	}
}

// IsRetryableError determines if an error should trigger a retry
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation
	if err == context.Canceled {
		return false
	}

	// Check for rate limiting
	if isRateLimitError(err) {
		return true
	}

	// Check for temporary/transient errors
	if isTemporaryError(err) {
		return true
	}

	return false
}

// isRateLimitError checks if the error is a rate limit error
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "rate limit") ||
		contains(errStr, "429") ||
		contains(errStr, "too many requests") ||
		contains(errStr, "rate_limit_exceeded")
}

// isTemporaryError checks if the error is a temporary/transient error
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	// Network errors
	if contains(err.Error(), "connection reset") ||
		contains(err.Error(), "connection refused") ||
		contains(err.Error(), "timeout") ||
		contains(err.Error(), "temporary failure") ||
		contains(err.Error(), "i/o timeout") ||
		contains(err.Error(), "network") {
		return true
	}

	// Server errors (5xx)
	if contains(err.Error(), "500") ||
		contains(err.Error(), "502") ||
		contains(err.Error(), "503") ||
		contains(err.Error(), "504") ||
		contains(err.Error(), "internal server error") {
		return true
	}

	return false
}

// RetryableRequest is a function that performs a request and returns an error
type RetryableRequest func() error

// Retry executes the request with exponential backoff
func Retry(ctx context.Context, config RetryConfig, req RetryableRequest) error {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = DefaultRetryConfig().MaxAttempts
	}
	if config.BaseDelay <= 0 {
		config.BaseDelay = DefaultRetryConfig().BaseDelay
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = DefaultRetryConfig().MaxDelay
	}
	if config.Multiplier <= 0 {
		config.Multiplier = DefaultRetryConfig().Multiplier
	}

	var lastErr error
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Attempt the request
		lastErr = req()
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if !IsRetryableError(lastErr) {
			return lastErr
		}

		// Don't wait after the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate delay with exponential backoff
		delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt))
		if delay > float64(config.MaxDelay) {
			delay = float64(config.MaxDelay)
		}

		// Add jitter (±25%)
		jitter := delay * 0.25
		delay = delay + (jitter * (2*rand.Float64() - 1))

		waitTime := time.Duration(delay)

		log.Printf("Retry attempt %d/%d after %v: %v", attempt+1, config.MaxAttempts, waitTime, lastErr)

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}

	return fmt.Errorf("after %d attempts: %w", config.MaxAttempts, lastErr)
}

// HTTPRetryer provides retry logic for HTTP requests
type HTTPRetryer struct {
	Client   *http.Client
	Config   RetryConfig
	LogLevel int // 0=silent, 1=attempt, 2=verbose
}

// NewHTTPRetryer creates a new HTTP retryer
func NewHTTPRetryer(client *http.Client, config RetryConfig) *HTTPRetryer {
	return &HTTPRetryer{
		Client:   client,
		Config:   config,
		LogLevel: 1,
	}
}

// Do performs an HTTP request with retry logic
func (h *HTTPRetryer) Do(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	var lastResp *http.Response

	err := Retry(ctx, h.Config, func() error {
		if h.LogLevel >= 2 {
			log.Printf("[HTTP Retryer] %s %s", method, url)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return err
		}

		resp, err := h.Client.Do(req)
		lastResp = resp

		if err != nil {
			return err
		}

		// Don't retry on successful responses
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		// Don't retry on client errors (except rate limiting)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			if resp.StatusCode == 429 {
				return fmt.Errorf("rate limit exceeded: %d", resp.StatusCode)
			}
			return fmt.Errorf("client error: %d", resp.StatusCode)
		}

		// Retry on server errors and rate limiting
		return fmt.Errorf("server error: %d", resp.StatusCode)
	})

	return lastResp, err
}
