package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"plan-agent/internal/api/db"
	"plan-agent/internal/core"
	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
	"plan-agent/internal/planclient"
)

// CreateMessageRequest is the request body for creating a message
type CreateMessageRequest struct {
	Content     string       `json:"content"`
	GoalContext *GoalContext `json:"goal_context,omitempty"`
}

// GoalContext gives interactive planning chat the goal it is discussing.
type GoalContext struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// CreateConversationRequest is the request body for creating a conversation
type CreateConversationRequest struct {
	Model    string `json:"model"`
	Provider string `json:"provider"`
	Title    string `json:"title"`
}

// UpdateConversationRequest is the request body for updating a conversation
type UpdateConversationRequest struct {
	Title string `json:"title"`
}

// ListConversations returns all conversations
// @Summary List all conversations
// @Tags conversations
// @Produce json
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/conversations [get]
func ListConversations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	conversations, err := db.ListConversations(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// CreateConversation creates a new conversation
// @Summary Create a new conversation
// @Tags conversations
// @Accept json
// @Produce json
// @Param request body CreateConversationRequest true "Conversation request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/conversations [post]
func CreateConversation(c *gin.Context) {
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Model == "" {
		req.Model = "MiniMax-M2.7"
	}
	if req.Provider == "" {
		req.Provider = "minimax"
	}
	if req.Title == "" {
		req.Title = "New conversation"
	}

	conv, err := db.CreateConversation(c.Request.Context(), req.Model, req.Provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		db.UpdateConversationTitle(c.Request.Context(), conv.ID, req.Title)
	}
	c.JSON(http.StatusCreated, conv)
}

// GetConversation returns a single conversation
// @Summary Get a conversation by ID
// @Tags conversations
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/conversations/{id} [get]
func GetConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	conv, err := db.GetConversation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	c.JSON(http.StatusOK, conv)
}

// GetMessages returns all messages for a conversation
// @Summary Get messages for a conversation
// @Tags messages
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/conversations/{id}/messages [get]
func GetMessages(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	messages, err := db.GetMessages(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// CreateMessage sends a user message, runs the agent, and streams tool calls as they complete.
// Uses chunked transfer encoding for real-time updates.
// @Summary Send a message to the agent
// @Tags messages
// @Accept json
// @Produce application/x-ndjson
// @Param id path string true "Conversation ID"
// @Param request body CreateMessageRequest true "Message request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/conversations/{id}/messages [post]
func CreateMessage(c *gin.Context) {
	ctx := c.Request.Context()
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv, err := db.GetConversation(ctx, convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	userMsg, err := db.CreateMessage(ctx, convID, "user", req.Content, 0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var planClient *planclient.Client
	var recorder *runRecorder
	if req.GoalContext != nil {
		planClient, err = planclient.New()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Plan client is not configured"})
			return
		}
		recorder = newRunRecorder(ctx, planClient, req.GoalContext.ID)
	}

	agent, err := createAgentFromConversation(ctx, conv, req.GoalContext, planClient, recorder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	history, err := db.GetMessages(ctx, convID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set up chunked streaming
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Cache-Control", "no-cache")

	// Stream user message immediately
	sendChunk(c, gin.H{"type": "user_message", "message": userMsg})

	// Run agent and stream events as they happen
	prompt := req.Content
	if req.GoalContext != nil {
		prompt = formatGoalContext(*req.GoalContext) + "\n\nUSER REQUEST:\n" + req.Content
	}

	runAgentStreaming(agent, ctx, dbMessagesToAgentMessages(history), prompt, convID, c, recorder)
}

func formatGoalContext(goal GoalContext) string {
	context := fmt.Sprintf("GOAL #%d: %s", goal.ID, goal.Title)
	if goal.Description != "" {
		context += "\n\nGOAL DESCRIPTION:\n" + goal.Description
	}
	return "We are discussing this Plan goal:\n\n" + context
}

// sendChunk writes a newline-delimited JSON chunk to the response
func sendChunk(c *gin.Context, data interface{}) {
	if data == nil {
		return
	}
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	c.Writer.Write(b)
	c.Writer.Write([]byte("\n"))
	c.Writer.Flush()
}

// runAgentStreaming runs the agent and streams events as they occur
func runAgentStreaming(agent *core.AgentRuntime, ctx context.Context, history []messages.Message, prompt string, convID uuid.UUID, c *gin.Context, recorder *runRecorder) {
	var pendingTools []toolCall
	var responseText, thinkingText string
	// toolTurn räknar verktygsanropen i den här turen så att historiken
	// går att följa.
	toolTurn := 0

	result := agent.SendMessageWithHistory(ctx, history, prompt, func(event *core.AgentEvent) {
		log.Printf("agent event: %s", event.Type)
		switch event.Type {
		case core.EventMessageEnd:
			if event.Message != nil {
				for _, block := range event.Message.Content {
					if tb, ok := block.(*messages.TextBlock); ok {
						responseText += tb.Text
					}
				}
			}
		case core.EventToolStart:
			argsJSON := "{}"
			if event.Args != nil {
				if b, err := json.Marshal(event.Args); err == nil {
					argsJSON = string(b)
				}
			}
			tc := toolCall{
				id:   event.ToolCallID,
				name: event.ToolName,
				args: argsJSON,
			}
			pendingTools = append(pendingTools, tc)
			toolTurn++

			recorder.Record(ctx, planclient.EventToolCall, "", map[string]any{
				"tool": event.ToolName,
				"args": event.Args,
				"turn": toolTurn,
			})

			// Stream tool_request immediately
			msg, err := db.CreateToolMessage(ctx, convID, "tool_request", tc.name, tc.args, "", false)
			if err == nil {
				sendChunk(c, gin.H{"type": "tool_request", "message": msg})
			}

		case core.EventToolUpdate:
			// Stream tool output updates (for streaming tools like bash)
			if event.Result != nil {
				output := extractTextResult(event.Result)
				if output != "" {
					sendChunk(c, gin.H{
						"type":       "tool_output",
						"toolCallId": event.ToolCallID,
						"toolName":   event.ToolName,
						"output":     output,
						"isError":    event.IsError,
					})
				}
			}

		case core.EventToolEnd:
			resultText := extractTextResult(event.Result)
			isError := event.IsError
			for i := range pendingTools {
				if pendingTools[i].id == event.ToolCallID {
					pendingTools[i].result = resultText
					pendingTools[i].isError = isError
					break
				}
			}

			recorder.Record(ctx, planclient.EventToolResult, "", map[string]any{
				"tool":     event.ToolName,
				"is_error": isError,
				"result":   resultText,
			})

			// Stream tool_result immediately
			toolName, toolArgs := "", ""
			for _, tc := range pendingTools {
				if tc.id == event.ToolCallID {
					toolName, toolArgs = tc.name, tc.args
					break
				}
			}
			msg, err := db.CreateToolMessage(ctx, convID, "tool_result", toolName, toolArgs, resultText, isError)
			if err == nil {
				sendChunk(c, gin.H{"type": "tool_result", "message": msg})
			}
		}
	})

	// Ett fel i agenten ska synas - annars ser chatten ut att ha tyst
	// misslyckats.
	if result != nil && result.Error != nil {
		log.Printf("Agent run failed for conversation %s: %v", convID, result.Error)
		sendChunk(c, gin.H{"type": "error", "error": result.Error.Error()})
		recorder.Record(ctx, planclient.EventError, "", map[string]any{
			"error": result.Error.Error(),
		})
	}

	log.Printf("agent response length: %d", len(responseText))

	// Process final response
	thinkingText = extractThinkBlocks(responseText)
	responseText = stripThinkBlocks(responseText)

	// Persist and stream thinking
	if thinkingText != "" {
		if msg, err := db.CreateThinkingMessage(ctx, convID, thinkingText); err == nil {
			sendChunk(c, gin.H{"type": "thinking", "message": msg})
		}
	}

	// Persist and stream assistant message
	if responseText != "" {
		if msg, err := db.CreateMessage(ctx, convID, "assistant", responseText, 0, 0); err == nil {
			sendChunk(c, gin.H{"type": "assistant_message", "message": msg})
		}
	}

	// Send completion marker
	sendChunk(c, gin.H{"type": "done"})
}

// toolCall represents a captured tool call from the agent event stream
type toolCall struct {
	id      string
	name    string
	args    string
	result  string
	isError bool
}

// extractThinkBlocks extracts content between<think> and</think> tags
func extractThinkBlocks(text string) string {
	const start = "<think>"
	const end = "</think>"
	var result string
	for {
		startIdx := 0
		endIdx := 0
		for i := 0; i <= len(text)-len(start); i++ {
			if text[i:i+len(start)] == start {
				startIdx = i + len(start)
				break
			}
		}
		for i := startIdx; i <= len(text)-len(end); i++ {
			if text[i:i+len(end)] == end {
				endIdx = i
				break
			}
		}
		if startIdx == 0 || endIdx == 0 || endIdx <= startIdx {
			break
		}
		if result != "" {
			result += "\n\n"
		}
		result += text[startIdx:endIdx]
		text = text[endIdx+len(end):]
	}
	return result
}

// stripThinkBlocks removes<think> and</think> tags from text
func stripThinkBlocks(text string) string {
	// Byten kopieras rakt av - att gå via string(text[i]) skulle förstöra
	// multi-byte-tecken (till exempel å, ä, ö).
	var buf bytes.Buffer
	inThink := false
	i := 0
	for i < len(text) {
		if i <= len(text)-len("<think>") && text[i:i+len("<think>")] == "<think>" {
			inThink = true
			i += len("<think>")
			continue
		}
		if i <= len(text)-len("</think>") && text[i:i+len("</think>")] == "</think>" {
			inThink = false
			i += len("</think>")
			continue
		}
		if !inThink {
			buf.WriteByte(text[i])
		}
		i++
	}
	return buf.String()
}

// extractTextResult extracts text from a ToolResult's content blocks
func extractTextResult(result *provider.ToolResult) string {
	if result == nil {
		return ""
	}
	for _, block := range result.Content {
		if tb, ok := block.(*messages.TextBlock); ok {
			return tb.Text
		}
	}
	return ""
}

// createAgentFromConversation builds an AgentRuntime configured for a conversation
func createAgentFromConversation(ctx context.Context, conv *db.Conversation, goalContext *GoalContext, planClient *planclient.Client, recorder *runRecorder) (*core.AgentRuntime, error) {
	keys, err := db.GetAPIKeys(ctx)
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string)
	modelMap := make(map[string]string)
	for _, k := range keys {
		keyMap[k.Provider] = k.APIKey
		modelMap[k.Provider] = k.Model
	}

	cfg := resolveProvider(conv, keyMap, modelMap)
	if cfg == nil {
		return nil, &NoAPIKeyError{}
	}

	availableTools := getDefaultTools()
	if goalContext != nil && planClient != nil {
		availableTools = append(availableTools, createSubgoalTool(planClient, goalContext.ID, recorder))
	}

	return core.NewAgent(&core.AgentLoopConfig{
		Model:         provider.Model{Provider: cfg.name, Name: cfg.model, APIKey: cfg.apiKey, Endpoint: cfg.url},
		Provider:      createProvider(cfg.name, cfg.url),
		Tools:         availableTools,
		SystemPrompt:  buildSystemPrompt(),
		ToolExecution: provider.ToolExecutionParallel,
		MaxTurns:      conversationMaxTurns(),
	}), nil
}

// conversationMaxTurns limits a single chat request so a confused model cannot
// keep invoking tools indefinitely. The default is intentionally conservative.
func conversationMaxTurns() int {
	const defaultMaxTurns = 12

	value, err := strconv.Atoi(os.Getenv("AGENT_MAX_TURNS"))
	if err != nil || value < 1 {
		return defaultMaxTurns
	}
	return value
}

// providerEnvKey returns the environment variable name for a provider's API key
func providerEnvKey(provider string) string {
	switch provider {
	case "agentgateway":
		return "AGENTGATEWAY_API_KEY"
	case "minimax":
		return "MINIMAX_API_KEY"
	case "openai":
		return "OPENAI_API_KEY"
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	default:
		return provider + "_API_KEY"
	}
}

// resolveProvider picks the best available provider for a conversation
func resolveProvider(conv *db.Conversation, keyMap, modelMap map[string]string) *selectedProvider {
	// Prefer conversation's configured provider
	if conv.Provider != "" && conv.Model != "" {
		apiKey := keyMap[conv.Provider]
		if apiKey == "" {
			apiKey = os.Getenv(providerEnvKey(conv.Provider))
		}
		if apiKey != "" {
			model := conv.Model
			if m := modelMap[conv.Provider]; m != "" {
				model = m
			}
			return &selectedProvider{name: conv.Provider, apiKey: apiKey, model: model}
		}
	}

	// Fall back to provider priority
	return selectProvider(keyMap, modelMap, os.Getenv)
}

// UpdateConversationTitle updates a conversation's title
// @Summary Update a conversation
// @Tags conversations
// @Accept json
// @Produce json
// @Param id path string true "Conversation ID"
// @Param request body UpdateConversationRequest true "Update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/conversations/{id} [patch]
func UpdateConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	var req UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		if err := db.UpdateConversationTitle(c.Request.Context(), id, req.Title); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	conv, err := db.GetConversation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	c.JSON(http.StatusOK, conv)
}
