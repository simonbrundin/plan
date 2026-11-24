package orchestrator

import (
	"strings"
	"testing"
	"time"

	"plan-agent/internal/core/checkpoint"
)

func TestHTMLViewerV2ViewPlan(t *testing.T) {
	v := NewHTMLViewerV2()

	plan := &checkpoint.CheckpointPlan{
		ID:          "test-plan",
		Feature:     "Build Login Feature",
		Description: "Implement user authentication",
		Status:      checkpoint.PlanApproved,
		UpdatedAt:   time.Now(),
		Checkpoints: []*checkpoint.Checkpoint{
			{
				ID:          "cp1",
				Name:        "Database Schema",
				Description: "Create users table",
				Complexity:  checkpoint.ComplexityLow,
				Status:      checkpoint.CheckpointPassed,
				Tasks:       []string{"Create users table", "Add indexes"},
				Gates:       []string{"behavior"},
				GateResults: map[string]*checkpoint.GateResult{
					"behavior": {Status: checkpoint.GateStatusPassed, Feedback: "All tests pass"},
				},
			},
			{
				ID:          "cp2",
				Name:        "API Endpoints",
				Description: "Create login/logout endpoints",
				Complexity:  checkpoint.ComplexityMedium,
				Status:      checkpoint.CheckpointInProgress,
				Tasks:       []string{"POST /login", "POST /logout"},
				Gates:       []string{"behavior", "adversarial"},
				GateResults: map[string]*checkpoint.GateResult{
					"behavior": {Status: checkpoint.GateStatusPassed},
				},
			},
		},
	}

	html := v.ViewPlan(plan)

	// Check key elements are present
	if !strings.Contains(html, "Build Login Feature") {
		t.Error("Expected plan feature name in HTML")
	}
	if !strings.Contains(html, "Database Schema") {
		t.Error("Expected checkpoint name in HTML")
	}
	if !strings.Contains(html, "passed") {
		t.Error("Expected status 'passed' in HTML")
	}
	if !strings.Contains(html, "in_progress") {
		t.Error("Expected status 'in_progress' in HTML")
	}
	if !strings.Contains(html, "Create users table") {
		t.Error("Expected task in HTML")
	}
	if !strings.Contains(html, "behavior") {
		t.Error("Expected gate name in HTML")
	}
}

func TestHTMLViewerV2EmptyPlan(t *testing.T) {
	v := NewHTMLViewerV2()

	plan := &checkpoint.CheckpointPlan{
		ID:          "empty",
		Feature:     "Empty Plan",
		Description: "No checkpoints",
		Checkpoints: []*checkpoint.Checkpoint{},
	}

	html := v.ViewPlan(plan)

	if !strings.Contains(html, "Empty Plan") {
		t.Error("Expected feature name in HTML")
	}
	if !strings.Contains(html, "0 of 0 checkpoints") {
		t.Error("Expected progress text for empty plan")
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>", "&lt;script&gt;"},
		{"a & b", "a &amp; b"},
		{`quote "test"`, "quote &quot;test&quot;"},
		{"normal text", "normal text"},
		{"<>&\"'", "&lt;&gt;&amp;&quot;&#39;"},
	}

	for _, tt := range tests {
		result := escapeHTML(tt.input)
		if result != tt.expected {
			t.Errorf("escapeHTML(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestHTMLViewerV2WithFailedGate(t *testing.T) {
	v := NewHTMLViewerV2()

	plan := &checkpoint.CheckpointPlan{
		ID:      "failed",
		Feature: "Failed Feature",
		Status:  checkpoint.PlanApproved,
		Checkpoints: []*checkpoint.Checkpoint{
			{
				ID:     "cp1",
				Name:   "Implementation",
				Status: checkpoint.CheckpointInProgress,
				Gates:  []string{"behavior", "adversarial"},
				GateResults: map[string]*checkpoint.GateResult{
					"behavior":    {Status: checkpoint.GateStatusPassed},
					"adversarial": {Status: checkpoint.GateStatusFailed, Feedback: "Security issue found"},
				},
			},
		},
	}

	html := v.ViewPlan(plan)

	if !strings.Contains(html, "failed") {
		t.Error("Expected failed status in HTML")
	}
	if !strings.Contains(html, "Security issue found") {
		t.Error("Expected feedback in HTML")
	}
}

func TestHTMLViewerV2MultipleCheckpoints(t *testing.T) {
	v := NewHTMLViewerV2()

	// Create 5 checkpoints
	checkpoints := make([]*checkpoint.Checkpoint, 5)
	for i := 0; i < 5; i++ {
		status := checkpoint.CheckpointPending
		if i < 2 {
			status = checkpoint.CheckpointPassed
		} else if i == 2 {
			status = checkpoint.CheckpointInProgress
		}
		checkpoints[i] = &checkpoint.Checkpoint{
			ID:     "cp" + string(rune('0'+i)),
			Name:   "Checkpoint " + string(rune('A'+i)),
			Status: status,
		}
	}

	plan := &checkpoint.CheckpointPlan{
		ID:          "multi",
		Feature:     "Multi Feature",
		Checkpoints: checkpoints,
	}

	html := v.ViewPlan(plan)

	// Should show 2 of 5 complete (40%)
	if !strings.Contains(html, "2 of 5 checkpoints") {
		t.Error("Expected progress text")
	}

	// Should have 5 checkpoints (check by checkpoint class at start of div)
	count := strings.Count(html, "<div class=\"checkpoint ")
	if count != 5 {
		t.Errorf("Expected 5 checkpoint divs, got %d", count)
	}
}

func TestFormatTimestamp(t *testing.T) {
	// Test with zero time
	result := formatTimestamp(time.Time{})
	if result != "Just now" {
		t.Errorf("Expected 'Just now', got %s", result)
	}

	// Test with valid time
	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	result = formatTimestamp(timestamp)
	if !strings.Contains(result, "2024") {
		t.Errorf("Expected year 2024 in timestamp, got %s", result)
	}
}
