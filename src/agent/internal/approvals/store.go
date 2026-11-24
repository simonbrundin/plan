package approvals

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PendingApproval represents a goal definition awaiting approval
type PendingApproval struct {
	GoalID         int64  `json:"goal_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	CreatedAt      string `json:"created_at"`
	Attempts       int    `json:"attempts"`
	Feedback       string `json:"feedback,omitempty"`
	Status         string `json:"status,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// Approval statusar.
const (
	// StatusPending betyder att förslaget väntar på ditt beslut.
	StatusPending = "pending"
	// StatusRejected betyder att du avvisade forslaget. Da far agenten
	// om att gora om definitionen med din feedback.
	StatusRejected = "rejected"
)

// Store handles file-based storage of pending approvals
type Store struct {
	baseDir string
}

// New creates a new approval store
func New() (*Store, error) {
	baseDir := os.Getenv("AGENT_DATA_DIR")
	if baseDir == "" {
		baseDir = "agent-data"
	}

	// Create directory structure
	pendingDir := filepath.Join(baseDir, "pending-approvals")
	if err := os.MkdirAll(pendingDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create pending approvals directory: %w", err)
	}

	return &Store{
		baseDir: baseDir,
	}, nil
}

// pendingPath returns the file path for a pending approval
func (s *Store) pendingPath(goalID int64) string {
	return filepath.Join(s.baseDir, "pending-approvals", fmt.Sprintf("goal-%d.json", goalID))
}

// Save stores a pending approval
func (s *Store) Save(approval *PendingApproval) error {
	data, err := json.MarshalIndent(approval, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal approval: %w", err)
	}

	path := s.pendingPath(approval.GoalID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write approval file: %w", err)
	}

	return nil
}

// Get retrieves a pending approval by goal ID
func (s *Store) Get(goalID int64) (*PendingApproval, error) {
	path := s.pendingPath(goalID)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read approval file: %w", err)
	}

	var approval PendingApproval
	if err := json.Unmarshal(data, &approval); err != nil {
		return nil, fmt.Errorf("failed to parse approval file: %w", err)
	}

	return &approval, nil
}

// Delete removes a pending approval
func (s *Store) Delete(goalID int64) error {
	path := s.pendingPath(goalID)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete approval file: %w", err)
	}

	return nil
}

// List returns all pending approvals
func (s *Store) List() ([]*PendingApproval, error) {
	pendingDir := filepath.Join(s.baseDir, "pending-approvals")

	entries, err := os.ReadDir(pendingDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read pending approvals directory: %w", err)
	}

	var approvals []*PendingApproval
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		var approval PendingApproval
		data, err := os.ReadFile(filepath.Join(pendingDir, entry.Name()))
		if err != nil {
			continue // Skip unreadable files
		}

		if err := json.Unmarshal(data, &approval); err != nil {
			continue // Skip malformed files
		}

		approvals = append(approvals, &approval)
	}

	return approvals, nil
}

// UpdateFeedback updates the feedback on an existing approval
func (s *Store) UpdateFeedback(goalID int64, feedback string) error {
	approval, err := s.Get(goalID)
	if err != nil {
		return err
	}
	if approval == nil {
		return fmt.Errorf("no pending approval for goal %d", goalID)
	}

	approval.Feedback = feedback

	return s.Save(approval)
}
