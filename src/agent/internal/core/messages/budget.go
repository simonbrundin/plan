package messages

import (
	"strings"
	"sync"
	"time"
)

// ContextBudget monitors context window usage and warns when approaching limits
type ContextBudget struct {
	mu sync.RWMutex

	// Model limits (tokens)
	modelMaxTokens    int
	warningThreshold  float64 // 0.0-1.0, warn when this percentage is reached
	criticalThreshold float64 // 0.0-1.0, critical alert threshold

	// Current usage
	currentTokens    int
	sessionStartTime time.Time

	// Callbacks
	onWarning        func(usage float64, tokens int)
	onCritical       func(usage float64, tokens int)
	onBudgetExceeded func(current int, max int)
}

// NewContextBudget creates a new context budget monitor
func NewContextBudget(modelMaxTokens int) *ContextBudget {
	return &ContextBudget{
		modelMaxTokens:    modelMaxTokens,
		warningThreshold:  0.75, // 75%
		criticalThreshold: 0.90, // 90%
		sessionStartTime:  time.Now(),
	}
}

// SetThresholds configures warning and critical thresholds
func (b *ContextBudget) SetThresholds(warning, critical float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.warningThreshold = warning
	b.criticalThreshold = critical
}

// SetWarningCallback sets the callback for warning events
func (b *ContextBudget) SetWarningCallback(fn func(usage float64, tokens int)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onWarning = fn
}

// SetCriticalCallback sets the callback for critical events
func (b *ContextBudget) SetCriticalCallback(fn func(usage float64, tokens int)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onCritical = fn
}

// SetBudgetExceededCallback sets the callback when budget is exceeded
func (b *ContextBudget) SetBudgetExceededCallback(fn func(current int, max int)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onBudgetExceeded = fn
}

// UpdateUsage updates the current token count and triggers callbacks if needed
func (b *ContextBudget) UpdateUsage(tokens int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.currentTokens = tokens
	usage := float64(tokens) / float64(b.modelMaxTokens)

	// Check critical threshold first
	if usage >= b.criticalThreshold {
		if b.onCritical != nil {
			b.mu.Unlock()
			b.onCritical(usage, tokens)
			b.mu.Lock()
		}
		return
	}

	// Check warning threshold
	if usage >= b.warningThreshold {
		if b.onWarning != nil {
			b.mu.Unlock()
			b.onWarning(usage, tokens)
			b.mu.Lock()
		}
	}
}

// AddUsage adds tokens to current usage
func (b *ContextBudget) AddUsage(tokens int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.currentTokens += tokens
	usage := float64(b.currentTokens) / float64(b.modelMaxTokens)

	if usage >= b.criticalThreshold {
		if b.onCritical != nil {
			b.mu.Unlock()
			b.onCritical(usage, b.currentTokens)
			b.mu.Lock()
		}
	} else if usage >= b.warningThreshold {
		if b.onWarning != nil {
			b.mu.Unlock()
			b.onWarning(usage, b.currentTokens)
			b.mu.Lock()
		}
	}
}

// GetUsage returns current usage stats
func (b *ContextBudget) GetUsage() (current int, max int, usage float64) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.currentTokens, b.modelMaxTokens, float64(b.currentTokens) / float64(b.modelMaxTokens)
}

// GetBudgetStatus returns the current budget status
func (b *ContextBudget) GetBudgetStatus() BudgetStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()

	usage := float64(b.currentTokens) / float64(b.modelMaxTokens)

	status := BudgetStatus{
		CurrentTokens:   b.currentTokens,
		MaxTokens:       b.modelMaxTokens,
		UsagePercentage: usage,
		SessionDuration: time.Since(b.sessionStartTime),
	}

	if usage >= b.criticalThreshold {
		status.Level = BudgetLevelCritical
		status.Message = "Context budget critical - summarization recommended"
	} else if usage >= b.warningThreshold {
		status.Level = BudgetLevelWarning
		status.Message = "Context budget approaching limit"
	} else {
		status.Level = BudgetLevelOK
		status.Message = "Context budget healthy"
	}

	return status
}

// BudgetStatus represents the current state of context budget
type BudgetStatus struct {
	Level           BudgetLevel
	Message         string
	CurrentTokens   int
	MaxTokens       int
	UsagePercentage float64
	SessionDuration time.Duration
}

// BudgetLevel represents the severity of budget usage
type BudgetLevel int

const (
	BudgetLevelOK BudgetLevel = iota
	BudgetLevelWarning
	BudgetLevelCritical
)

// String returns a string representation of the budget level
func (l BudgetLevel) String() string {
	switch l {
	case BudgetLevelOK:
		return "OK"
	case BudgetLevelWarning:
		return "WARNING"
	case BudgetLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// EstimateMessageTokens estimates tokens in a message (rough approximation)
// In production, use tiktoken or similar
func EstimateMessageTokens(msg Message) int {
	// Rough estimate: ~4 chars per token for English
	content := extractTextContent(msg)
	return len(content) / 4
}

// extractTextContent extracts text from a message
func extractTextContent(msg Message) string {
	var sb strings.Builder
	for _, block := range msg.GetContent() {
		switch b := block.(type) {
		case *TextBlock:
			sb.WriteString(b.Text)
		case *ImageBlock:
			// Images are expensive - estimate ~1000 tokens
			sb.WriteString(" [image placeholder] ")
		}
	}
	return sb.String()
}

// CountMessagesTokens counts total tokens in a slice of messages
func CountMessagesTokens(msgs []Message) int {
	total := 0
	for _, msg := range msgs {
		total += EstimateMessageTokens(msg)
	}
	return total
}

// ContextBudgetManager manages budgets for multiple conversations
type ContextBudgetManager struct {
	mu               sync.RWMutex
	budgets          map[string]*ContextBudget
	defaultMaxTokens int
}

// NewContextBudgetManager creates a manager for multiple conversation budgets
func NewContextBudgetManager(defaultMaxTokens int) *ContextBudgetManager {
	return &ContextBudgetManager{
		budgets:          make(map[string]*ContextBudget),
		defaultMaxTokens: defaultMaxTokens,
	}
}

// GetOrCreateBudget gets or creates a budget for a conversation
func (m *ContextBudgetManager) GetOrCreateBudget(conversationID string) *ContextBudget {
	m.mu.Lock()
	defer m.mu.Unlock()

	if budget, exists := m.budgets[conversationID]; exists {
		return budget
	}

	budget := NewContextBudget(m.defaultMaxTokens)
	m.budgets[conversationID] = budget
	return budget
}

// RemoveBudget removes a conversation's budget
func (m *ContextBudgetManager) RemoveBudget(conversationID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.budgets, conversationID)
}

// GetAllStatuses returns budget status for all conversations
func (m *ContextBudgetManager) GetAllStatuses() map[string]BudgetStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make(map[string]BudgetStatus)
	for id, budget := range m.budgets {
		statuses[id] = budget.GetBudgetStatus()
	}
	return statuses
}

// SuggestPruning analyzes a conversation and suggests what to prune
func SuggestPruning(messages []Message, targetReduction float64) []PruningSuggestion {
	var suggestions []PruningSuggestion

	if len(messages) < 4 {
		return suggestions // Not enough to prune
	}

	// Analyze message age and content
	keepCount := int(float64(len(messages)) * (1 - targetReduction))

	for i := 0; i < len(messages)-keepCount; i++ {
		msg := messages[i]
		tokens := EstimateMessageTokens(msg)

		suggestions = append(suggestions, PruningSuggestion{
			MessageIndex: i,
			MessageRole:  msg.GetRole(),
			Tokens:       tokens,
			Reason:       getPruningReason(msg),
		})
	}

	return suggestions
}

// PruningSuggestion represents a suggested message to prune
type PruningSuggestion struct {
	MessageIndex int
	MessageRole  string
	Tokens       int
	Reason       string
}

func getPruningReason(msg Message) string {
	// System messages keep
	if msg.GetRole() == string(RoleSystem) {
		return "Cannot prune system messages"
	}

	// Old user messages might be less relevant
	if msg.GetRole() == string(RoleUser) {
		return "Older user message - context still recoverable from later messages"
	}

	// Assistant messages can be summarized
	return "Can be summarized or removed"
}
