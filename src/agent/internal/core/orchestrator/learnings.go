package orchestrator

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// LearningsManager manages the learnings file
type LearningsManager struct {
	filePath string
}

// NewLearningsManager creates a new learnings manager
func NewLearningsManager(filePath string) *LearningsManager {
	return &LearningsManager{filePath: filePath}
}

// AddLearning adds a learning entry
func (m *LearningsManager) AddLearning(category, title, description, lesson string) error {
	entry := fmt.Sprintf(`## %s: %s

- **Date**: %s
- **Category**: %s
- **Description**: %s
- **Lesson**: %s

`,
		time.Now().Format("2006-01-02"),
		title,
		time.Now().Format("2006-01-02"),
		category,
		description,
		lesson,
	)

	// Read existing content
	var content string
	if data, err := os.ReadFile(m.filePath); err == nil {
		content = string(data)
	}

	// Prepend new entry
	content = "# Learnings\n\nThis file contains lessons learned from checkpoint execution.\n\n" + entry + "\n" + content

	return os.WriteFile(m.filePath, []byte(content), 0644)
}

// AddCheckpointLearning adds a learning from a checkpoint
func (m *LearningsManager) AddCheckpointLearning(cpName, issue, fix, lesson string) error {
	return m.AddLearning("checkpoint", cpName, issue, lesson)
}

// GetLearnings returns all learnings
func (m *LearningsManager) GetLearnings() (string, error) {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "# Learnings\n\nNo learnings recorded yet.\n", nil
		}
		return "", err
	}
	return string(data), nil
}

// GeneratePromptContext generates context for agent prompts
// It filters learnings by the given categories if provided
func (m *LearningsManager) GeneratePromptContext(categories ...string) string {
	content, err := m.GetLearnings()
	if err != nil {
		return ""
	}

	// Only include recent learnings (last 10)
	lines := strings.Split(content, "\n")
	var relevant []string
	count := 0
	includeCurrent := false

	for _, line := range lines {
		// Track category from Date headers
		if strings.HasPrefix(line, "## 20") {
			count++
			if count > 10 {
				break
			}
			includeCurrent = false
			// Extract category if available
			for _, cat := range categories {
				if strings.Contains(line, cat) {
					includeCurrent = true
					break
				}
			}
		}

		// Track category from Category lines
		if strings.HasPrefix(line, "- **Category**:") {
			for _, cat := range categories {
				if strings.Contains(line, cat) {
					includeCurrent = true
					break
				}
			}
		}

		// If no categories specified, include all
		if len(categories) == 0 {
			includeCurrent = true
		}

		if includeCurrent {
			relevant = append(relevant, line)
		}
	}

	if len(relevant) == 0 {
		return ""
	}

	return "\n\n## Relevant Learnings\n\n" + strings.Join(relevant, "\n")
}

// GetLearningsByCategory returns learnings filtered by category
func (m *LearningsManager) GetLearningsByCategory(category string) (string, error) {
	all, err := m.GetLearnings()
	if err != nil {
		return "", err
	}

	lines := strings.Split(all, "\n")
	var filtered []string
	includeCurrent := false

	for _, line := range lines {
		if strings.HasPrefix(line, "## 20") || strings.HasPrefix(line, "# Learnings") {
			includeCurrent = strings.Contains(line, category)
		}
		if includeCurrent || strings.HasPrefix(line, "# Learnings") {
			filtered = append(filtered, line)
		}
	}

	if len(filtered) == 0 {
		return "", nil
	}

	return strings.Join(filtered, "\n"), nil
}
