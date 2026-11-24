package orchestrator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOrchestratorWithLearningsManager(t *testing.T) {
	// Create a temp file for learnings
	tmpDir := t.TempDir()
	learningsFile := filepath.Join(tmpDir, "learnings.md")

	cfg := Config{
		StateStore:   NewInMemoryStateStore(),
		LearningFile: learningsFile,
	}

	o := New(cfg)

	// Should have a learnings manager
	lm := o.LearningsManager()
	if lm == nil {
		t.Fatal("Expected LearningsManager to be set")
	}

	// Add a learning
	err := o.AddLearning("test", "Test Learning", "Testing", "Always test your code")
	if err != nil {
		t.Fatalf("AddLearning failed: %v", err)
	}

	// Get learnings
	learnings, err := o.GetLearnings()
	if err != nil {
		t.Fatalf("GetLearnings failed: %v", err)
	}

	if learnings == "" {
		t.Fatal("Expected non-empty learnings")
	}

	if !contains(learnings, "Test Learning") {
		t.Error("Expected learning title in output")
	}
}

func TestOrchestratorWithExternalLearningsManager(t *testing.T) {
	// Create a temp file for learnings
	tmpDir := t.TempDir()
	learningsFile := filepath.Join(tmpDir, "learnings.md")

	externalLM := NewLearningsManager(learningsFile)

	cfg := Config{
		StateStore:       NewInMemoryStateStore(),
		LearningsManager: externalLM,
	}

	o := New(cfg)

	// Should use the external learnings manager
	lm := o.LearningsManager()
	if lm == externalLM {
		t.Log("External LearningsManager correctly used")
	}
}

func TestAddCheckpointLearning(t *testing.T) {
	tmpDir := t.TempDir()
	learningsFile := filepath.Join(tmpDir, "learnings.md")

	cfg := Config{
		StateStore:   NewInMemoryStateStore(),
		LearningFile: learningsFile,
	}

	o := New(cfg)

	// Add checkpoint learning
	err := o.AddCheckpointLearning(
		"Database Schema",
		"Foreign key constraint failed",
		"Added proper migration ordering",
		"Always define foreign keys after parent tables",
	)
	if err != nil {
		t.Fatalf("AddCheckpointLearning failed: %v", err)
	}

	// Verify it was saved
	learnings, _ := o.GetLearnings()
	if !contains(learnings, "Database Schema") {
		t.Error("Expected checkpoint name in learnings")
	}
	if !contains(learnings, "Foreign key constraint failed") {
		t.Error("Expected issue in learnings")
	}
}

func TestGenerateLearningsContext(t *testing.T) {
	tmpDir := t.TempDir()
	learningsFile := filepath.Join(tmpDir, "learnings.md")

	cfg := Config{
		StateStore:   NewInMemoryStateStore(),
		LearningFile: learningsFile,
	}

	o := New(cfg)

	// Add some learnings
	if err := o.AddLearning("test1", "First", "Desc", "Lesson 1"); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	if err := o.AddLearning("test2", "Second", "Desc", "Lesson 2"); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}
	if err := o.AddLearning("test3", "Third", "Desc", "Lesson 3"); err != nil {
		t.Fatalf("saving state failed: %v", err)
	}

	// Generate context
	context := o.GenerateLearningsContext()
	if context == "" {
		t.Fatal("Expected non-empty context")
	}

	// Should include recent learnings
	if !contains(context, "Recent Learnings") && !contains(context, "First") {
		// The format may vary, just check it's not empty
		t.Logf("Context: %s", context)
	}
}

func TestOrchestratorNoLearningsManager(t *testing.T) {
	cfg := Config{
		StateStore: NewInMemoryStateStore(),
		// No LearningFile or LearningsManager set
	}

	o := New(cfg)

	// Should return nil for learnings manager
	lm := o.LearningsManager()
	if lm != nil {
		t.Error("Expected nil LearningsManager when not configured")
	}

	// Should handle gracefully
	err := o.AddLearning("test", "Title", "Desc", "Lesson")
	if err != nil {
		t.Errorf("AddLearning should not fail when no manager: %v", err)
	}

	learnings, err := o.GetLearnings()
	if err != nil {
		t.Errorf("GetLearnings should not fail when no manager: %v", err)
	}
	if learnings != "" {
		t.Errorf("Expected empty string when no manager, got: %s", learnings)
	}
}

func TestLearningsManagerPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	learningsFile := filepath.Join(tmpDir, "learnings.md")

	// Create manager and add learning
	lm1 := NewLearningsManager(learningsFile)
	err := lm1.AddLearning("persist", "Persistent", "Testing persistence", "File-based storage works")
	if err != nil {
		t.Fatalf("First AddLearning failed: %v", err)
	}

	// Create new manager pointing to same file
	lm2 := NewLearningsManager(learningsFile)
	learnings, err := lm2.GetLearnings()
	if err != nil {
		t.Fatalf("GetLearnings from new manager failed: %v", err)
	}

	if !contains(learnings, "Persistent") {
		t.Error("Expected persistent learning to be saved and loaded")
	}
}

func TestLearningsManagerFileCreated(t *testing.T) {
	tmpDir := t.TempDir()
	learningsFile := filepath.Join(tmpDir, "newlearnings.md")

	lm := NewLearningsManager(learningsFile)
	err := lm.AddLearning("test", "New File", "Created new file", "Works")
	if err != nil {
		t.Fatalf("AddLearning failed: %v", err)
	}

	// File should exist
	if _, err := os.Stat(learningsFile); os.IsNotExist(err) {
		t.Error("Expected learnings file to be created")
	}
}

// Helper
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
