package goals

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Global store for default tools
var (
	defaultStore     Store
	defaultStoreOnce sync.Once
	defaultStoreErr  error
)

// DefaultStorePath returns the default path for goals.json
func DefaultStorePath() string {
	// Try to find repo root
	exe, err := os.Executable()
	if err == nil {
		// bin/agent -> agent/ -> repo/ -> goals.json
		base := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
		path := filepath.Join(base, "goals.json")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Fallback: try working directory
	wd, _ := os.Getwd()
	for i := 0; i < 6; i++ {
		path := filepath.Join(wd, "goals.json")
		if _, err := os.Stat(path); err == nil {
			return path
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}

	// Ultimate fallback
	home, _ := os.UserHomeDir()
	if home != "" {
		path := filepath.Join(home, ".config", "agent", "goals.json")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "goals.json"
}

// DefaultStore returns a singleton JSON store initialized with the default path
func DefaultStore(ctx context.Context) (Store, error) {
	defaultStoreOnce.Do(func() {
		path := DefaultStorePath()
		defaultStore, defaultStoreErr = NewJSONStore(ctx, JSONStoreConfig{Path: path})
	})
	return defaultStore, defaultStoreErr
}

// JSONStore is a file-based goal store
type JSONStore struct {
	mu     sync.RWMutex
	path   string
	goals  map[string]*Goal
	status []Status
}

// JSONStoreConfig holds configuration for JSONStore
type JSONStoreConfig struct {
	Path string
}

// NewJSONStore creates a new JSON file-based goal store
func NewJSONStore(ctx context.Context, cfg JSONStoreConfig) (*JSONStore, error) {
	store := &JSONStore{
		path:   cfg.Path,
		goals:  make(map[string]*Goal),
		status: DefaultStatuses(),
	}

	if err := store.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// File doesn't exist, create with defaults
		if err := store.save(); err != nil {
			return nil, err
		}
	}

	return store, nil
}

// load reads the goals from file
func (s *JSONStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	var doc struct {
		Goals    []*Goal  `json:"goals"`
		Statuses []Status `json:"statuses"`
	}

	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}

	s.goals = make(map[string]*Goal)
	for _, g := range doc.Goals {
		s.goals[g.ID] = g
	}

	if len(doc.Statuses) > 0 {
		s.status = doc.Statuses
	}

	return nil
}

// save writes the goals to file
func (s *JSONStore) save() error {
	allGoals := make([]*Goal, 0, len(s.goals))
	for _, g := range s.goals {
		allGoals = append(allGoals, g)
	}

	doc := struct {
		Goals    []*Goal  `json:"goals"`
		Statuses []Status `json:"statuses"`
	}{
		Goals:    allGoals,
		Statuses: s.status,
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// GetAll returns all goals
func (s *JSONStore) GetAll(ctx context.Context) ([]*Goal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	goals := make([]*Goal, 0, len(s.goals))
	for _, g := range s.goals {
		goals = append(goals, g)
	}

	return goals, nil
}

// GetByID returns a goal by ID
func (s *JSONStore) GetByID(ctx context.Context, id string) (*Goal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	g, ok := s.goals[id]
	if !ok {
		return nil, ErrNotFound
	}

	return g, nil
}

// GetChildren returns direct children of a goal
func (s *JSONStore) GetChildren(ctx context.Context, parentID string) ([]*Goal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var children []*Goal
	for _, g := range s.goals {
		if g.ParentID == parentID {
			children = append(children, g)
		}
	}

	return children, nil
}

// Create creates a new goal
func (s *JSONStore) Create(ctx context.Context, goal *Goal) error {
	if err := ValidateGoal(goal); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.goals[goal.ID]; exists {
		return ErrDuplicate
	}

	if goal.Created.IsZero() {
		goal.Created = time.Now()
	}

	s.goals[goal.ID] = goal

	return s.save()
}

// Update updates an existing goal
func (s *JSONStore) Update(ctx context.Context, id string, goal *Goal) error {
	if err := ValidateGoal(goal); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.goals[id]; !exists {
		return ErrNotFound
	}

	goal.ID = id
	s.goals[id] = goal

	return s.save()
}

// Delete deletes a goal
func (s *JSONStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.goals[id]; !exists {
		return ErrNotFound
	}

	delete(s.goals, id)

	return s.save()
}

// GetStatuses returns all available statuses
func (s *JSONStore) GetStatuses(ctx context.Context) ([]Status, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.status, nil
}
