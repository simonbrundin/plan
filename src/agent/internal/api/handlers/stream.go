package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"plan/agent/internal/api/db"
)

// StreamMessages handles SSE for real-time message updates
// @Summary Stream messages via SSE
// @Tags messages
// @Produce text/event-stream
// @Param id path string true "Conversation ID"
// @Success 200 {string} string "SSE stream"
// @Failure 400 {object} map[string]string
// @Router /api/v1/conversations/{id}/stream [get]
func StreamMessages(c *gin.Context) {
	convIDStr := c.Param("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Access-Control-Allow-Origin", "http://localhost:3000")

	// Create a channel for new messages
	messageChan := make(chan *MessageEvent)
	stopChan := make(chan struct{})

	// Start polling database for new messages
	go pollMessages(convID, messageChan, stopChan)

	// Stream messages to client
	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			close(stopChan)
			return
		case msg, ok := <-messageChan:
			if !ok {
				return
			}
			data, err := json.Marshal(msg)
			if err != nil {
				return
			}
			if _, err := c.Writer.Write([]byte("data: " + string(data) + "\n\n")); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}

// MessageEvent represents a new message event
type MessageEvent struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// pollMessages polls the database for new messages
func pollMessages(convID uuid.UUID, ch chan<- *MessageEvent, stop <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	defer close(ch)

	lastCheck := time.Now()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			msgs, err := db.GetMessagesSince(context.Background(), convID, lastCheck)
			if err != nil {
				log.Printf("Error polling messages: %v", err)
				continue
			}
			for _, msg := range msgs {
				ch <- &MessageEvent{
					Type:      "message",
					ID:        msg.ID.String(),
					Role:      msg.Role,
					Content:   msg.Content,
					CreatedAt: msg.CreatedAt,
				}
				lastCheck = msg.CreatedAt
			}
		}
	}
}
