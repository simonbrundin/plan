package handlers

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"plan/agent/internal/core"
	"plan/agent/internal/core/messages"
	"plan/agent/internal/core/provider"
	"plan/agent/internal/core/tools"
)

// AgentWebSocketHandler handles WebSocket connections with full agent support
type AgentWebSocketHandler struct {
	providerManager *provider.ProviderManager
}

// NewAgentWebSocketHandler creates a new agent WebSocket handler
func NewAgentWebSocketHandler(pm *provider.ProviderManager) *AgentWebSocketHandler {
	return &AgentWebSocketHandler{
		providerManager: pm,
	}
}

// WsMessage is the JSON message format for agent WebSocket chat
type WsMessage struct {
	Type         string             `json:"type"`
	Model        string             `json:"model,omitempty"`
	Messages     []WsMessageContent `json:"messages,omitempty"`
	Temperature  float64            `json:"temperature,omitempty"`
	ConvID       string             `json:"convId,omitempty"`
	SystemPrompt string             `json:"systemPrompt,omitempty"`
	Tools        []WsToolDefinition `json:"tools,omitempty"`
}

// WsMessageContent represents a message in the WebSocket protocol
type WsMessageContent struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
}

// WsToolDefinition represents a tool definition
type WsToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// WsStreamChunk represents a chunk sent to the client
type WsStreamChunk struct {
	Type       string `json:"type"`
	Content    string `json:"content,omitempty"`
	ToolName   string `json:"toolName,omitempty"`
	ToolCallID string `json:"toolCallId,omitempty"`
	Error      string `json:"error,omitempty"`
	ConvID     string `json:"convId,omitempty"`
	Done       bool   `json:"done,omitempty"`
}

// HandleAgentChat handles WebSocket chat with full agent support including tools
// GET /api/v1/ws/agent
func (h *AgentWebSocketHandler) HandleAgentChat(c *gin.Context) {
	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer func() { _ = conn.Close() }()

	// Generate conversation ID
	convID := uuid.New().String()

	// WebSocket write mutex for thread safety
	var writeMu sync.Mutex

	// Send function that handles write errors
	send := func(chunk WsStreamChunk) {
		writeMu.Lock()
		defer writeMu.Unlock()
		if err := conn.WriteJSON(chunk); err != nil {
			log.Printf("WebSocket write error: %v", err)
		}
	}

	// Event handler that converts agent events to WebSocket messages
	eventHandler := func(event *core.AgentEvent) {
		switch event.Type {
		case core.EventMessageStart:
			// Assistant is starting to respond

		case core.EventMessageUpdate:
			// Assistant is updating (for partial content)

		case core.EventMessageEnd:
			// Assistant finished responding

		case core.EventToolStart:
			send(WsStreamChunk{
				Type:       "tool_start",
				ToolName:   event.ToolName,
				ToolCallID: event.ToolCallID,
				ConvID:     convID,
			})

		case core.EventToolUpdate:
			// Tool is streaming output
			if event.Result != nil && len(event.Result.Content) > 0 {
				for _, block := range event.Result.Content {
					if tb, ok := block.(*messages.TextBlock); ok {
						send(WsStreamChunk{
							Type:       "tool_output",
							ToolName:   event.ToolName,
							ToolCallID: event.ToolCallID,
							Content:    tb.Text,
							ConvID:     convID,
						})
					}
				}
			}

		case core.EventToolEnd:
			send(WsStreamChunk{
				Type:       "tool_end",
				ToolName:   event.ToolName,
				ToolCallID: event.ToolCallID,
				ConvID:     convID,
			})

		case core.EventAgentEnd:
			send(WsStreamChunk{
				Type:   "done",
				ConvID: convID,
				Done:   true,
			})
		}
	}

	// Read messages from client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Parse request
		var req WsMessage
		if err := json.Unmarshal(message, &req); err != nil {
			send(WsStreamChunk{
				Type:   "error",
				Error:  "invalid JSON: " + err.Error(),
				ConvID: convID,
			})
			continue
		}

		// Handle different message types
		switch req.Type {
		case "chat":
			h.handleAgentChat(req, eventHandler)
		case "ping":
			send(WsStreamChunk{Type: "pong", ConvID: convID})
		default:
			send(WsStreamChunk{
				Type:   "error",
				Error:  "unknown message type: " + req.Type,
				ConvID: convID,
			})
		}
	}
}

// handleAgentChat processes a chat request with full agent support
func (h *AgentWebSocketHandler) handleAgentChat(
	req WsMessage,
	handler core.EventHandler,
) {
	// Convert messages
	var agentMessages []messages.Message
	for _, msg := range req.Messages {
		switch msg.Role {
		case "user":
			agentMessages = append(agentMessages, &messages.UserMessage{
				Role: messages.RoleUser,
				Content: []messages.ContentBlock{
					&messages.TextBlock{Text: msg.Content},
				},
			})
		case "assistant":
			agentMessages = append(agentMessages, &messages.AssistantMessage{
				Role: messages.RoleAssistant,
				Content: []messages.ContentBlock{
					&messages.TextBlock{Text: msg.Content},
				},
			})
		case "system":
			agentMessages = append(agentMessages, &messages.SystemMessage{
				Role: messages.RoleSystem,
				Content: []messages.ContentBlock{
					&messages.TextBlock{Text: msg.Content},
				},
			})
		}
	}

	// Convert tools
	var agentTools []*tools.ToolDefinition
	for _, tool := range req.Tools {
		agentTools = append(agentTools, &tools.ToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}

	// Use default tools if none provided
	if len(agentTools) == 0 {
		agentTools = tools.GetDefaultTools()
	}

	// Create system prompt
	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = tools.DefaultSystemPrompt()
	}

	// Create model config
	model := provider.Model{
		Provider: "minimax",
		Name:     req.Model,
	}
	if model.Name == "" {
		model.Name = "MiniMax-M2.7"
	}

	// Create agent
	agentConfig := &core.AgentLoopConfig{
		Model:           model,
		Tools:           agentTools,
		ProviderManager: h.providerManager,
		SystemPrompt:    systemPrompt,
	}

	agent := core.NewAgent(agentConfig)

	// Send connected message
	handler(&core.AgentEvent{
		Type: core.EventAgentStart,
	})

	// Run the agent with the converted messages
	result := agent.SendMessageWithHistory(context.TODO(), agentMessages, "", handler)
	if result.Error != nil {
		handler(&core.AgentEvent{
			Type:    core.EventAgentEnd,
			IsError: true,
		})
	}

	_ = result // result contains final messages if needed
}
