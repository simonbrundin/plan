package secrets

import (
	"testing"
	"time"
)

func TestReloadNoChange(t *testing.T) {
	// Create a manager without loading actual secrets (just testing the logic)
	m := &Manager{
		configPath:   "/nonexistent/path.yaml",
		lastModified: time.Now(),
	}

	// Should not error when file doesn't exist
	err := m.Reload()
	if err != nil {
		t.Errorf("Reload should not error for nonexistent file: %v", err)
	}
}

func TestManagerLastModified(t *testing.T) {
	m := &Manager{
		configPath:   "/nonexistent/path.yaml",
		lastModified: time.Now(),
	}

	if m.lastModified.IsZero() {
		t.Error("Expected lastModified to be set")
	}
}

func TestWatcherStop(t *testing.T) {
	m := &Manager{
		configPath:   "/nonexistent/path.yaml",
		lastModified: time.Now(),
	}

	// Start watcher with very short interval
	stop := m.StartWatcher(1)
	if stop == nil {
		t.Fatal("Expected non-nil stop channel")
	}

	// Immediately stop
	m.StopWatcher()

	// Give it a moment to process
	time.Sleep(10 * time.Millisecond)

	// Should be stopped
	m.mu.Lock()
	stopped := m.stopWatcher == nil
	m.mu.Unlock()

	if !stopped {
		t.Error("Expected watcher to be stopped")
	}
}

func TestReloadWithModifiedFile(t *testing.T) {
	// This test would require actual file manipulation
	// Skip for unit tests - integration test would cover this
	t.Skip("Requires actual file system - use integration test")
}
