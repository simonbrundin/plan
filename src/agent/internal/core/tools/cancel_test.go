package tools

import (
	"testing"
	"time"
)

func TestCancellationManager_Register(t *testing.T) {
	cm := NewCancellationManager()

	// Register a cancel channel
	ch := cm.Register("tool-1")
	if ch == nil {
		t.Fatal("Register returned nil channel")
	}

	// Verify it's the same channel
	ch2 := cm.Register("tool-1")
	if ch != ch2 {
		t.Error("Register returned different channel for same ID")
	}

	// Register a different ID
	ch3 := cm.Register("tool-2")
	if ch == ch3 {
		t.Error("Different IDs should return different channels")
	}
}

func TestCancellationManager_Cancel(t *testing.T) {
	cm := NewCancellationManager()

	// Register
	ch := cm.Register("tool-1")

	// Cancel
	cm.Cancel("tool-1")

	// Verify channel is closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Channel should be closed after Cancel")
		}
	default:
		t.Error("Channel should be closed")
	}
}

func TestCancellationManager_IsCancelled(t *testing.T) {
	cm := NewCancellationManager()

	// Not cancelled yet
	if cm.IsCancelled("tool-1") {
		t.Error("Should not be cancelled before Cancel is called")
	}

	// Register and cancel
	cm.Register("tool-1")
	cm.Cancel("tool-1")

	// Now should be cancelled
	if !cm.IsCancelled("tool-1") {
		t.Error("Should be cancelled after Cancel is called")
	}
}

func TestCancellationManager_Unregister(t *testing.T) {
	cm := NewCancellationManager()

	// Register
	cm.Register("tool-1")

	// Unregister
	cm.Unregister("tool-1")

	// Should not be cancelled
	if cm.IsCancelled("tool-1") {
		t.Error("Should not be cancelled after Unregister")
	}
}

func TestCancellationManager_CancelAll(t *testing.T) {
	cm := NewCancellationManager()

	// Register multiple
	cm.Register("tool-1")
	cm.Register("tool-2")
	cm.Register("tool-3")

	if cm.Count() != 3 {
		t.Errorf("Count = %d, want 3", cm.Count())
	}

	// Cancel all
	cm.CancelAll()

	if cm.Count() != 0 {
		t.Errorf("Count = %d, want 0 after CancelAll", cm.Count())
	}
}

func TestCancellationManager_Concurrent(t *testing.T) {
	cm := NewCancellationManager()

	// Concurrent register and cancel
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			cm.Register("tool-1")
			time.Sleep(1 * time.Millisecond)
			cm.Cancel("tool-1")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGetGlobalCancellationManager(t *testing.T) {
	// Get twice should return same instance
	m1 := GetGlobalCancellationManager()
	m2 := GetGlobalCancellationManager()

	if m1 != m2 {
		t.Error("GetGlobalCancellationManager should return same instance")
	}
}
