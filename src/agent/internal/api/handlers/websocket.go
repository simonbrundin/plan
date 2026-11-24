package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"plan-agent/internal/api/llm"
)

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// WebSocketHandler handles WebSocket connections for real-time chat
type WebSocketHandler struct {
	client *llm.Client
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		client: llm.DefaultClient(),
	}
}

// WsChatRequest is the JSON message format for WebSocket chat
type WsChatRequest struct {
	Type        string        `json:"type"`
	Model       string        `json:"model,omitempty"`
	Messages    []llm.Message `json:"messages,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	ConvID      string        `json:"convId,omitempty"`
}

// HandleChat handles WebSocket chat connections
// GET /api/v1/ws/chat
func (h *WebSocketHandler) HandleChat(c *gin.Context) {
	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer func() { _ = conn.Close() }()

	// Generate conversation ID
	convID := uuid.New().String()

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
		var req WsChatRequest
		if err := json.Unmarshal(message, &req); err != nil {
			h.sendError(conn, "invalid JSON: "+err.Error())
			continue
		}

		// Handle different message types
		switch req.Type {
		case "chat":
			h.handleChat(conn, convID, req)
		case "ping":
			h.sendJSON(conn, map[string]string{"type": "pong"})
		default:
			h.sendError(conn, "unknown message type: "+req.Type)
		}
	}
}

// handleChat processes a chat request over WebSocket
func (h *WebSocketHandler) handleChat(conn *websocket.Conn, convID string, req WsChatRequest) {
	// Override convID if provided
	if req.ConvID != "" {
		convID = req.ConvID
	}

	// Send conversation ID
	h.sendJSON(conn, map[string]string{
		"type":   "connected",
		"convId": convID,
	})

	// Create chat request
	chatReq := llm.ChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
	}

	// Stream response via WebSocket
	if err := h.client.StreamChat(context.TODO(), chatReq, func(chunk llm.StreamChunk) {
		h.sendJSON(conn, chunk)
	}); err != nil {
		log.Printf("failed to stream chat: %v", err)
	}

	// Send done signal
	h.sendJSON(conn, map[string]interface{}{
		"type":   "done",
		"convId": convID,
	})
}

// sendJSON sends a JSON message over WebSocket
func (h *WebSocketHandler) sendJSON(conn *websocket.Conn, data any) {
	if err := conn.WriteJSON(data); err != nil {
		log.Printf("WebSocket write error: %v", err)
	}
}

// sendError sends an error message over WebSocket
func (h *WebSocketHandler) sendError(conn *websocket.Conn, message string) {
	h.sendJSON(conn, map[string]interface{}{
		"type":  "error",
		"error": message,
	})
}
