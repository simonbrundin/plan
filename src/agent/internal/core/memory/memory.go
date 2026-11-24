package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MemoryType represents the type of memory
type MemoryType string

const (
	MemoryTypeEpisodic   MemoryType = "episodic"   // Past experiences/conversations
	MemoryTypeSemantic   MemoryType = "semantic"   // Facts and knowledge
	MemoryTypeProcedural MemoryType = "procedural" // Learned procedures/patterns
	MemoryTypeWorking    MemoryType = "working"    // Current context (short-term)
)

// Memory represents a single memory entry
type Memory struct {
	ID           string                 `json:"id"`
	Type         MemoryType             `json:"type"`
	Content      string                 `json:"content"`
	Summary      string                 `json:"summary,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	AccessCount  int                    `json:"access_count"`
	LastAccessed time.Time              `json:"last_accessed,omitempty"`
	Embedding    []float32              `json:"embedding,omitempty"` // For semantic search
}

// MemoryStore manages all memory types
type MemoryStore struct {
	mu       sync.RWMutex
	memories map[string][]*Memory
	basePath string
}

// NewMemoryStore creates a new memory store
func NewMemoryStore(basePath string) (*MemoryStore, error) {
	store := &MemoryStore{
		memories: make(map[string][]*Memory),
		basePath: basePath,
	}

	// Load existing memories
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create memory directory: %w", err)
	}

	for _, memType := range []MemoryType{MemoryTypeEpisodic, MemoryTypeSemantic, MemoryTypeProcedural} {
		if err := store.load(memType); err != nil {
			// Log but don't fail - memories might not exist yet
			continue
		}
	}

	return store, nil
}

// load loads memories of a specific type from disk
func (s *MemoryStore) load(memType MemoryType) error {
	filePath := filepath.Join(s.basePath, fmt.Sprintf("%s.json", memType))

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var memories []*Memory
	if err := json.Unmarshal(data, &memories); err != nil {
		return err
	}

	s.memories[string(memType)] = memories
	return nil
}

// save saves memories of a specific type to disk
func (s *MemoryStore) save(memType MemoryType) error {
	filePath := filepath.Join(s.basePath, fmt.Sprintf("%s.json", memType))

	s.mu.RLock()
	memories := s.memories[string(memType)]
	s.mu.RUnlock()

	data, err := json.MarshalIndent(memories, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// Add adds a new memory
func (s *MemoryStore) Add(memType MemoryType, content, summary string, tags []string, metadata map[string]interface{}) (*Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mem := &Memory{
		ID:          fmt.Sprintf("%s-%d", memType, time.Now().UnixNano()),
		Type:        memType,
		Content:     content,
		Summary:     summary,
		Tags:        tags,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AccessCount: 0,
	}

	s.memories[string(memType)] = append(s.memories[string(memType)], mem)

	if err := s.save(memType); err != nil {
		return nil, err
	}

	return mem, nil
}

// AddEpisodic adds an episodic memory (past experience)
func (s *MemoryStore) AddEpisodic(conversationID, content, summary string, metadata map[string]interface{}) (*Memory, error) {
	tags := []string{"conversation", conversationID}
	if metadata != nil {
		if task, ok := metadata["task"].(string); ok {
			tags = append(tags, "task:"+task)
		}
		if success, ok := metadata["success"].(bool); ok && success {
			tags = append(tags, "success")
		} else {
			tags = append(tags, "failure")
		}
	}
	return s.Add(MemoryTypeEpisodic, content, summary, tags, metadata)
}

// AddSemantic adds a semantic memory (fact/knowledge)
func (s *MemoryStore) AddSemantic(fact, source string, confidence float64) (*Memory, error) {
	metadata := map[string]interface{}{
		"source":     source,
		"confidence": confidence,
	}
	tags := []string{"fact", "knowledge"}
	return s.Add(MemoryTypeSemantic, fact, fact[:min(100, len(fact))], tags, metadata)
}

// AddProcedural adds a procedural memory (learned procedure)
func (s *MemoryStore) AddProcedural(title, procedure, context string, tags []string) (*Memory, error) {
	content := fmt.Sprintf("## %s\n\n### Context\n%s\n\n### Procedure\n%s", title, context, procedure)
	metadata := map[string]interface{}{
		"context": context,
	}
	if tags == nil {
		tags = []string{"procedure", "learned"}
	}
	return s.Add(MemoryTypeProcedural, content, title, tags, metadata)
}

// Get retrieves a memory by ID
func (s *MemoryStore) Get(memType MemoryType, id string) *Memory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, mem := range s.memories[string(memType)] {
		if mem.ID == id {
			mem.AccessCount++
			mem.LastAccessed = time.Now()
			return mem
		}
	}
	return nil
}

// GetRecent returns the most recent memories of a type
func (s *MemoryStore) GetRecent(memType MemoryType, limit int) []*Memory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	memories := s.memories[string(memType)]
	if limit > len(memories) {
		limit = len(memories)
	}

	// Return last 'limit' memories (most recent)
	result := make([]*Memory, limit)
	copy(result, memories[len(memories)-limit:])
	return result
}

// GetByTag returns memories with a specific tag
func (s *MemoryStore) GetByTag(memType MemoryType, tag string) []*Memory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Memory
	for _, mem := range s.memories[string(memType)] {
		for _, t := range mem.Tags {
			if t == tag {
				result = append(result, mem)
				break
			}
		}
	}
	return result
}

// Search performs a simple text search across memories
func (s *MemoryStore) Search(memType MemoryType, query string, limit int) []*Memory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	var result []*Memory

	for _, mem := range s.memories[string(memType)] {
		if strings.Contains(strings.ToLower(mem.Content), query) ||
			strings.Contains(strings.ToLower(mem.Summary), query) {
			mem.AccessCount++
			result = append(result, mem)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}

	return result
}

// Delete removes a memory by ID
func (s *MemoryStore) Delete(memType MemoryType, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	memories := s.memories[string(memType)]
	for i, mem := range memories {
		if mem.ID == id {
			s.memories[string(memType)] = append(memories[:i], memories[i+1:]...)
			return s.save(memType)
		}
	}
	return nil
}

// GetStats returns statistics about stored memories
func (s *MemoryStore) GetStats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]int)
	total := 0
	for memType, memories := range s.memories {
		stats[memType] = len(memories)
		total += len(memories)
	}
	stats["total"] = total
	return stats
}

// Prune removes old memories based on access patterns
func (s *MemoryStore) Prune(memType MemoryType, keepCount int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	memories := s.memories[string(memType)]
	if len(memories) <= keepCount {
		return 0, nil
	}

	removed := len(memories) - keepCount
	s.memories[string(memType)] = memories[len(memories)-keepCount:]

	if err := s.save(memType); err != nil {
		return 0, err
	}

	return removed, nil
}

// ConversationMemory stores memories from a single conversation
type ConversationMemory struct {
	ConversationID string
	StartTime      time.Time
	EndTime        time.Time
	Summary        string
	Outcome        string // "success", "failure", "partial"
	Learnings      []string
	Tasks          []string
}

// RecordConversation records a completed conversation
func (s *MemoryStore) RecordConversation(conv *ConversationMemory) error {
	content := fmt.Sprintf(`## Conversation %s

**Duration**: %s - %s
**Outcome**: %s

### Tasks
%s

### Learnings
%s

### Summary
%s`,
		conv.ConversationID,
		conv.StartTime.Format(time.RFC3339),
		conv.EndTime.Format(time.RFC3339),
		conv.Outcome,
		strings.Join(conv.Tasks, "\n- "),
		strings.Join(conv.Learnings, "\n- "),
		conv.Summary,
	)

	metadata := map[string]interface{}{
		"outcome":   conv.Outcome,
		"startTime": conv.StartTime,
		"endTime":   conv.EndTime,
		"tasks":     conv.Tasks,
	}

	_, err := s.Add(MemoryTypeEpisodic, content, conv.Summary, []string{"conversation", conv.Outcome}, metadata)
	return err
}

// MemoryManager provides high-level memory operations
type MemoryManager struct {
	store *MemoryStore
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(basePath string) (*MemoryManager, error) {
	store, err := NewMemoryStore(basePath)
	if err != nil {
		return nil, err
	}
	return &MemoryManager{store: store}, nil
}

// RecordExperience records a new experience
func (m *MemoryManager) RecordExperience(conversationID, content, summary string, success bool, tasks []string) error {
	conv := &ConversationMemory{
		ConversationID: conversationID,
		StartTime:      time.Now().Add(-1 * time.Hour), // Approximate
		EndTime:        time.Now(),
		Summary:        summary,
		Outcome:        "success",
		Tasks:          tasks,
	}
	if !success {
		conv.Outcome = "failure"
	}
	return m.store.RecordConversation(conv)
}

// GetRelevantMemories returns memories relevant to a task
func (m *MemoryManager) GetRelevantMemories(task string, limit int) []*Memory {
	// Search episodic memory for related experiences
	episodic := m.store.Search(MemoryTypeEpisodic, task, limit)

	// Search procedural memory for related procedures
	procedural := m.store.Search(MemoryTypeProcedural, task, limit)

	// Combine and deduplicate
	seen := make(map[string]*Memory)
	for _, mem := range episodic {
		seen[mem.ID] = mem
	}
	for _, mem := range procedural {
		if _, exists := seen[mem.ID]; !exists {
			seen[mem.ID] = mem
		}
	}

	result := make([]*Memory, 0, len(seen))
	for _, mem := range seen {
		result = append(result, mem)
	}

	if limit > 0 && len(result) > limit {
		return result[:limit]
	}
	return result
}

// GenerateContextForTask generates relevant memory context for a task
func (m *MemoryManager) GenerateContextForTask(task string) string {
	memories := m.GetRelevantMemories(task, 5)
	if len(memories) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n## Relevant Past Experiences\n\n")

	for _, mem := range memories {
		sb.WriteString(fmt.Sprintf("### %s (%s)\n", mem.Type, mem.CreatedAt.Format("2006-01-02")))
		sb.WriteString(mem.Summary)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// GetStats returns memory statistics
func (m *MemoryManager) GetStats() map[string]int {
	return m.store.GetStats()
}
