package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"plan-agent/internal/api/llm"
)

// LLMHandler handles LLM API requests
type LLMHandler struct {
	client *llm.Client
}

// NewLLMHandler creates a new LLM handler
func NewLLMHandler() *LLMHandler {
	return &LLMHandler{
		client: llm.DefaultClient(),
	}
}

// AskRequest is the request body for /ask endpoint
type AskRequest struct {
	Model       string        `json:"model"`
	Messages    []llm.Message `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

// Ask handles streaming chat requests
// POST /api/v1/llm/ask
func (h *LLMHandler) Ask(c *gin.Context) {
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get conversation ID from header or generate new
	convIDStr := c.GetHeader("X-Conversation-ID")
	var convID uuid.UUID
	if convIDStr != "" {
		convID, _ = uuid.Parse(convIDStr)
	} else {
		convID = uuid.New()
	}

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Conversation-ID", convID.String())
	c.Header("Transfer-Encoding", "chunked")

	// Stream response
	chatReq := llm.ChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
	}

	h.client.StreamChat(c.Request.Context(), chatReq, func(chunk llm.StreamChunk) {
		data, _ := json.Marshal(chunk)
		c.Writer.Write([]byte("data: " + string(data) + "\n\n"))
		c.Writer.Flush()
	})
}

// Health checks if the LLM service is configured
// GET /api/v1/llm/health
func (h *LLMHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"provider": h.client.ProviderName(),
		"status":   "ok",
	})
}

// Providers returns available providers
// GET /api/v1/llm/providers
func (h *LLMHandler) Providers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"providers": []gin.H{
			{"name": "minimax", "base_url": "https://api.minimax.io/anthropic"},
			{"name": "anthropic", "base_url": "https://api.anthropic.com"},
		},
		"default": h.client.ProviderName(),
	})
}
