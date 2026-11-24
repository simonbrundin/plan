package tools

import (
	"sync"
)

// CancellationManager manages tool execution cancellation
type CancellationManager struct {
	mu          sync.RWMutex
	toolCancels map[string]chan struct{} // tool call ID -> cancel channel
}

// NewCancellationManager creates a new cancellation manager
func NewCancellationManager() *CancellationManager {
	return &CancellationManager{
		toolCancels: make(map[string]chan struct{}),
	}
}

// Register registers a cancel channel for a tool call
func (cm *CancellationManager) Register(toolCallID string) chan struct{} {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// If already exists, return existing
	if ch, ok := cm.toolCancels[toolCallID]; ok {
		return ch
	}

	ch := make(chan struct{})
	cm.toolCancels[toolCallID] = ch
	return ch
}

// Get returns the cancel channel for a tool call
func (cm *CancellationManager) Get(toolCallID string) <-chan struct{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if ch, ok := cm.toolCancels[toolCallID]; ok {
		return ch
	}
	return nil
}

// Cancel signals cancellation for a tool call
func (cm *CancellationManager) Cancel(toolCallID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ch, ok := cm.toolCancels[toolCallID]; ok {
		// Check if already closed using non-blocking receive
		select {
		case _, ok := <-ch:
			if !ok {
				// Already closed, remove from map
				delete(cm.toolCancels, toolCallID)
				return
			}
		default:
			// Not closed, close it
		}
		close(ch)
	}
}

// Unregister removes the cancel channel for a tool call
func (cm *CancellationManager) Unregister(toolCallID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if ch, ok := cm.toolCancels[toolCallID]; ok {
		// Only close if not already closed
		select {
		case _, ok := <-ch:
			if !ok {
				// Already closed, remove from map
				delete(cm.toolCancels, toolCallID)
			}
		default:
			// Not closed yet, close and remove
			close(ch)
			delete(cm.toolCancels, toolCallID)
		}
	}
}

// CancelAll cancels all running tool executions
func (cm *CancellationManager) CancelAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for id, ch := range cm.toolCancels {
		select {
		case _, ok := <-ch:
			if !ok {
				continue
			}
		default:
		}
		close(ch)
		delete(cm.toolCancels, id)
	}
}

// IsCancelled checks if a tool call has been cancelled
func (cm *CancellationManager) IsCancelled(toolCallID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if ch, ok := cm.toolCancels[toolCallID]; ok {
		select {
		case _, ok := <-ch:
			return !ok
		default:
			return false
		}
	}
	return false
}

// Count returns the number of active tool executions
func (cm *CancellationManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.toolCancels)
}

// Global cancellation manager for the agent
var globalCancellationManager *CancellationManager
var globalCancelMu sync.Once

// GetGlobalCancellationManager returns the global cancellation manager
func GetGlobalCancellationManager() *CancellationManager {
	globalCancelMu.Do(func() {
		globalCancellationManager = NewCancellationManager()
	})
	return globalCancellationManager
}
