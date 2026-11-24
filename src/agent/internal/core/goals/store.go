package goals

import (
	"context"
	"time"
)

// Goal represents a goal in the system
type Goal struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Icon        string     `json:"icon,omitempty"`
	Status      string     `json:"status"`
	Priority    int        `json:"priority,omitempty"`
	ParentID    string     `json:"parent_id,omitempty"`
	Created     time.Time  `json:"created"`
	Started     *time.Time `json:"started,omitempty"`
	Finished    *time.Time `json:"finished,omitempty"`
	Children    []*Goal    `json:"children,omitempty"`
}

// Status represents a goal status
type Status struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// Store defines the interface for goal storage
// Implement this to connect to any backend (JSON, Plan API, SQLite, etc.)
type Store interface {
	// GetAll returns all goals as a flat list
	GetAll(ctx context.Context) ([]*Goal, error)

	// GetByID returns a single goal by ID
	GetByID(ctx context.Context, id string) (*Goal, error)

	// GetChildren returns direct children of a goal
	GetChildren(ctx context.Context, parentID string) ([]*Goal, error)

	// Create creates a new goal
	Create(ctx context.Context, goal *Goal) error

	// Update updates an existing goal
	Update(ctx context.Context, id string, goal *Goal) error

	// Delete deletes a goal
	Delete(ctx context.Context, id string) error

	// GetStatuses returns all available statuses
	GetStatuses(ctx context.Context) ([]Status, error)
}

// DefaultStatuses returns the default goal statuses
func DefaultStatuses() []Status {
	return []Status{
		{Name: "define", Label: "Defining"},
		{Name: "plan", Label: "Planning"},
		{Name: "execute", Label: "Executing"},
		{Name: "verify", Label: "Verifying"},
		{Name: "completed", Label: "Completed"},
	}
}

// ValidateGoal validates a goal before saving
func ValidateGoal(g *Goal) error {
	if g.Title == "" {
		return ErrEmptyTitle
	}
	if g.ID == "" {
		return ErrEmptyID
	}
	return nil
}
