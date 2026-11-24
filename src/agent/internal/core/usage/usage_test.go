package usage

import (
	"context"
	"testing"
	"time"
)

func TestCalculateCost(t *testing.T) {
	costs := DefaultCosts()

	tests := []struct {
		name         string
		provider     string
		model        string
		inputTokens  int
		outputTokens int
		minExpected  float64
		maxExpected  float64
	}{
		{
			name:         "MiniMax M2.7",
			provider:     "minimax",
			model:        "MiniMax-M2.7",
			inputTokens:  1000,
			outputTokens: 500,
			minExpected:  0.0004, // Should be around $0.0004
			maxExpected:  0.001,
		},
		{
			name:         "OpenAI gpt-4o-mini",
			provider:     "openai",
			model:        "gpt-4o-mini",
			inputTokens:  1000,
			outputTokens: 500,
			minExpected:  0.00045,
			maxExpected:  0.001,
		},
		{
			name:         "Anthropic Claude",
			provider:     "anthropic",
			model:        "claude-sonnet-4-7-25",
			inputTokens:  1000,
			outputTokens: 500,
			minExpected:  0.01,
			maxExpected:  0.03,
		},
		{
			name:         "Unknown provider uses default",
			provider:     "unknown",
			model:        "unknown-model",
			inputTokens:  1000,
			outputTokens: 500,
			minExpected:  0.00075,
			maxExpected:  0.002,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := CalculateCost(tt.provider, tt.model, tt.inputTokens, tt.outputTokens, costs)
			if cost < tt.minExpected || cost > tt.maxExpected {
				t.Errorf("CalculateCost() = %v, want between %v and %v", cost, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestInMemoryTracker(t *testing.T) {
	tracker := InMemoryTracker()

	// Record some usage
	if err := tracker.Record(context.TODO(), TokenUsage{
		ConversationID: "conv-1",
		Provider:       "minimax",
		Model:          "MiniMax-M2.7",
		InputTokens:    100,
		OutputTokens:   50,
		LatencyMs:      100,
		Timestamp:      time.Now(),
	}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	if err := tracker.Record(context.TODO(), TokenUsage{
		ConversationID: "conv-1",
		Provider:       "openai",
		Model:          "gpt-4o-mini",
		InputTokens:    200,
		OutputTokens:   100,
		LatencyMs:      150,
		Timestamp:      time.Now(),
	}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	// Get stats
	stats, err := tracker.GetStats(context.TODO())
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalRequests != 2 {
		t.Errorf("TotalRequests = %d, want 2", stats.TotalRequests)
	}

	if stats.TotalInputTokens != 300 {
		t.Errorf("TotalInputTokens = %d, want 300", stats.TotalInputTokens)
	}

	if stats.TotalOutputTokens != 150 {
		t.Errorf("TotalOutputTokens = %d, want 150", stats.TotalOutputTokens)
	}

	if stats.TotalCost <= 0 {
		t.Error("TotalCost should be greater than 0")
	}

	if stats.AverageLatencyMs != 125 {
		t.Errorf("AverageLatencyMs = %d, want 125", stats.AverageLatencyMs)
	}
}

func TestTracker_Record(t *testing.T) {
	tracker := InMemoryTracker()

	usage := TokenUsage{
		ConversationID: "test-conv",
		Provider:       "test",
		Model:          "test-model",
		InputTokens:    100,
		OutputTokens:   50,
		LatencyMs:      100,
		Timestamp:      time.Now(),
	}

	err := tracker.Record(context.TODO(), usage)
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	// Verify usage was recorded with cost calculated
	if len(tracker.usage) != 1 {
		t.Errorf("len(tracker.usage) = %d, want 1", len(tracker.usage))
	}

	recorded := tracker.usage[0]
	if recorded.ID == "" {
		t.Error("ID should be set")
	}

	if recorded.Cost <= 0 {
		t.Error("Cost should be calculated")
	}

	if recorded.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

func TestCostConfig(t *testing.T) {
	costs := DefaultCosts()

	if len(costs) == 0 {
		t.Error("DefaultCosts() should return at least one cost config")
	}

	// Verify all costs have valid values
	for _, c := range costs {
		if c.Provider == "" {
			t.Error("Provider should not be empty")
		}
		if c.Model == "" {
			t.Error("Model should not be empty")
		}
		if c.InputCostPer1M < 0 {
			t.Error("InputCostPer1M should not be negative")
		}
		if c.OutputCostPer1M < 0 {
			t.Error("OutputCostPer1M should not be negative")
		}
	}
}
