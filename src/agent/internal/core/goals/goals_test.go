package goals

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONStore(t *testing.T) {
	// Create temp dir
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "goals.json")

	ctx := context.Background()

	// Create store
	store, err := NewJSONStore(ctx, JSONStoreConfig{Path: path})
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Test Create
	goal := &Goal{
		ID:     "test-goal",
		Title:  "Test Goal",
		Status: "define",
	}

	if err := store.Create(ctx, goal); err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}

	// Test GetByID
	got, err := store.GetByID(ctx, "test-goal")
	if err != nil {
		t.Fatalf("failed to get goal: %v", err)
	}
	if got.Title != "Test Goal" {
		t.Errorf("expected title 'Test Goal', got '%s'", got.Title)
	}

	// Test Update
	got.Status = "execute"
	if err := store.Update(ctx, "test-goal", got); err != nil {
		t.Fatalf("failed to update goal: %v", err)
	}

	// Verify update persisted
	store2, err := NewJSONStore(ctx, JSONStoreConfig{Path: path})
	if err != nil {
		t.Fatalf("failed to reload store: %v", err)
	}
	got2, _ := store2.GetByID(ctx, "test-goal")
	if got2.Status != "execute" {
		t.Errorf("expected status 'execute', got '%s'", got2.Status)
	}

	// Test GetAll
	all, err := store.GetAll(ctx)
	if err != nil {
		t.Fatalf("failed to get all: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 goal, got %d", len(all))
	}

	// Test Delete
	if err := store.Delete(ctx, "test-goal"); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Test NotFound
	_, err = store.GetByID(ctx, "test-goal")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestValidateGoal(t *testing.T) {
	tests := []struct {
		name    string
		goal    *Goal
		wantErr error
	}{
		{
			name:    "empty title",
			goal:    &Goal{ID: "test", Title: ""},
			wantErr: ErrEmptyTitle,
		},
		{
			name:    "empty id",
			goal:    &Goal{ID: "", Title: "Test"},
			wantErr: ErrEmptyID,
		},
		{
			name:    "valid goal",
			goal:    &Goal{ID: "test", Title: "Test"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGoal(tt.goal)
			if err != tt.wantErr {
				t.Errorf("ValidateGoal() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultStatuses(t *testing.T) {
	statuses := DefaultStatuses()
	if len(statuses) != 5 {
		t.Errorf("expected 5 statuses, got %d", len(statuses))
	}
}

func TestLoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent.json")

	ctx := context.Background()

	store, err := NewJSONStore(ctx, JSONStoreConfig{Path: path})
	if err != nil {
		t.Fatalf("failed to create store for non-existent file: %v", err)
	}

	// Should have defaults
	statuses, _ := store.GetStatuses(ctx)
	if len(statuses) != 5 {
		t.Errorf("expected 5 default statuses, got %d", len(statuses))
	}

	// File should exist now
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}
}
