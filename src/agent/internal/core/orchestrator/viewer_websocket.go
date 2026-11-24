package orchestrator

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"plan-agent/internal/core/checkpoint"
)

// WebSocketViewer sends checkpoint updates via WebSocket
type WebSocketViewer struct {
	mu          sync.RWMutex
	connections map[string]*websocket.Conn
	hub         *ViewerHub
}

// ViewerHub manages WebSocket connections for checkpoint updates
type ViewerHub struct {
	mu          sync.RWMutex
	connections map[string]*websocket.Conn
	register    chan *ViewerConnection
	unregister  chan *ViewerConnection
	broadcast   chan *ViewerMessage
}

// ViewerConnection represents a single WebSocket connection
type ViewerConnection struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

// ViewerMessage is the JSON message format for checkpoint updates
type ViewerMessage struct {
	Type       string                     `json:"type"`
	PlanID     string                     `json:"planId,omitempty"`
	Checkpoint *checkpoint.Checkpoint     `json:"checkpoint,omitempty"`
	Plan       *checkpoint.CheckpointPlan `json:"plan,omitempty"`
	Message    string                     `json:"message,omitempty"`
	Error      string                     `json:"error,omitempty"`
}

// NewWebSocketViewer creates a WebSocket viewer with a hub
func NewWebSocketViewer() *WebSocketViewer {
	hub := &ViewerHub{
		connections: make(map[string]*websocket.Conn),
		register:    make(chan *ViewerConnection),
		unregister:  make(chan *ViewerConnection),
		broadcast:   make(chan *ViewerMessage, 256),
	}

	// Start hub goroutine
	go hub.run()

	return &WebSocketViewer{
		connections: hub.connections,
		hub:         hub,
	}
}

// NewViewerHub creates a standalone hub for external use
func NewViewerHub() *ViewerHub {
	hub := &ViewerHub{
		connections: make(map[string]*websocket.Conn),
		register:    make(chan *ViewerConnection),
		unregister:  make(chan *ViewerConnection),
		broadcast:   make(chan *ViewerMessage, 256),
	}
	go hub.run()
	return hub
}

// GetHub returns the hub for external registration
func (w *WebSocketViewer) GetHub() *ViewerHub {
	return w.hub
}

// Register adds a new WebSocket connection to the hub
func (h *ViewerHub) Register(conn *websocket.Conn) *ViewerConnection {
	connection := &ViewerConnection{
		ID:   generateID(),
		Conn: conn,
		Send: make(chan []byte, 256),
	}
	h.register <- connection
	return connection
}

// unregister removes a connection from the hub
func (h *ViewerHub) unregisterConn(conn *ViewerConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.connections[conn.ID]; ok {
		delete(h.connections, conn.ID)
		close(conn.Send)
	}
}

// run is the hub's main event loop
func (h *ViewerHub) run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn.ID] = conn.Conn
			h.mu.Unlock()
			log.Printf("ViewerHub: registered connection %s", conn.ID)
			go h.writePump(conn)
			go h.readPump(conn)

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn.ID]; ok {
				delete(h.connections, conn.ID)
				close(conn.Send)
				log.Printf("ViewerHub: unregistered connection %s", conn.ID)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			data, merr := json.Marshal(msg)
			if merr != nil {
				log.Printf("ViewerHub: failed to marshal message: %v", merr)
				continue
			}

			h.mu.RLock()
			for id, conn := range h.connections {
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					// Connection is slow or closed, close it
					conn.Close()
					delete(h.connections, id)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// writePump handles outgoing messages for a connection
func (h *ViewerHub) writePump(conn *ViewerConnection) {
	defer func() {
		conn.Conn.Close()
	}()

	for message := range conn.Send {
		if err := conn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

// readPump handles incoming messages (for ping/pong and close)
func (h *ViewerHub) readPump(conn *ViewerConnection) {
	defer func() {
		h.unregister <- conn
		conn.Conn.Close()
	}()

	for {
		if _, _, err := conn.Conn.ReadMessage(); err != nil {
			return
		}
	}
}

// Broadcast sends a message to all connected viewers
func (h *ViewerHub) Broadcast(msg *ViewerMessage) {
	h.broadcast <- msg
}

// Broadcastf sends a formatted message to all connected viewers
func (h *ViewerHub) Broadcastf(msgType, message string) {
	h.Broadcast(&ViewerMessage{
		Type:    msgType,
		Message: message,
	})
}

// --- Viewer interface implementation ---

// ShowPlan displays a plan for approval (sends to WebSocket)
func (w *WebSocketViewer) ShowPlan(plan *checkpoint.CheckpointPlan) {
	w.hub.Broadcast(&ViewerMessage{
		Type:   "plan_created",
		PlanID: plan.ID,
		Plan:   plan,
	})
}

// ShowCheckpoint displays a checkpoint when it starts
func (w *WebSocketViewer) ShowCheckpoint(cp *checkpoint.Checkpoint) {
	w.hub.Broadcast(&ViewerMessage{
		Type:       "checkpoint_started",
		Checkpoint: cp,
	})
}

// UpdateCheckpoint displays updated checkpoint status
func (w *WebSocketViewer) UpdateCheckpoint(cp *checkpoint.Checkpoint) {
	msgType := "checkpoint_updated"

	switch cp.Status {
	case checkpoint.CheckpointPassed:
		msgType = "checkpoint_passed"
	case checkpoint.CheckpointFailed:
		msgType = "checkpoint_failed"
	case checkpoint.CheckpointInProgress:
		msgType = "checkpoint_in_progress"
	}

	w.hub.Broadcast(&ViewerMessage{
		Type:       msgType,
		Checkpoint: cp,
	})
}

// BroadcastGateUpdate sends a gate-specific update
func (w *WebSocketViewer) BroadcastGateUpdate(cp *checkpoint.Checkpoint, gateType string, result *checkpoint.GateResult) {
	msgType := "gate_updated"
	if result.Status == checkpoint.GateStatusPassed {
		msgType = "gate_passed"
	} else if result.Status == checkpoint.GateStatusFailed {
		msgType = "gate_failed"
	} else if result.Status == checkpoint.GateStatusRunning {
		msgType = "gate_started"
	}

	// Create a minimal checkpoint with just the gate result
	cpUpdate := &checkpoint.Checkpoint{
		ID:     cp.ID,
		Name:   cp.Name,
		Status: cp.Status,
		GateResults: map[string]*checkpoint.GateResult{
			gateType: result,
		},
	}

	w.hub.Broadcast(&ViewerMessage{
		Type:       msgType,
		Checkpoint: cpUpdate,
	})
}

// BroadcastError sends an error message
func (w *WebSocketViewer) BroadcastError(err error) {
	w.hub.Broadcast(&ViewerMessage{
		Type:  "error",
		Error: err.Error(),
	})
}

// generateID creates a unique ID for a connection
func generateID() string {
	return randomString(8)
}

// randomString generates a random alphanumeric string
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// Ensure WebSocketViewer implements Viewer interface
var _ Viewer = (*WebSocketViewer)(nil)
