package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"plan-agent/internal/core/checkpoint"
)

// DefaultCheckpointPlanner creates checkpoints based on heuristics
type DefaultCheckpointPlanner struct{}

// NewDefaultCheckpointPlanner creates a default planner
func NewDefaultCheckpointPlanner() *DefaultCheckpointPlanner {
	return &DefaultCheckpointPlanner{}
}

// CreateCheckpoints generates checkpoints from a feature description
func (p *DefaultCheckpointPlanner) CreateCheckpoints(ctx context.Context, feature, description string) ([]*checkpoint.Checkpoint, error) {
	var checkpoints []*checkpoint.Checkpoint

	// Check for common patterns
	featureLower := strings.ToLower(feature + " " + description)

	// Database/Backend work
	if containsAny(featureLower, "database", "schema", "model", "table", "migration", "persist") {
		cp := checkpoint.NewCheckpoint(
			uuid.New().String(),
			"Database Schema",
			"Set up database schema and models",
			checkpoint.ComplexityLow,
		)
		cp.AddGate("behavior")
		cp.AddTask("Create/update database schema")
		cp.AddTask("Define models")
		cp.AddTask("Add migrations if needed")
		checkpoints = append(checkpoints, cp)
	}

	// API endpoints
	if containsAny(featureLower, "api", "endpoint", "route", "handler", "controller") {
		cp := checkpoint.NewCheckpoint(
			uuid.New().String(),
			"API Endpoints",
			"Implement API endpoints",
			checkpoint.ComplexityMedium,
		)
		cp.AddGate("behavior")
		cp.AddGate("adversarial")
		cp.AddTask("Design API contracts")
		cp.AddTask("Implement handlers")
		cp.AddTask("Add input validation")
		cp.AddTask("Write API tests")
		checkpoints = append(checkpoints, cp)
	}

	// Business logic
	if containsAny(featureLower, "service", "logic", "rule", "validation", "process") {
		cp := checkpoint.NewCheckpoint(
			uuid.New().String(),
			"Business Logic",
			"Implement business logic and validation",
			checkpoint.ComplexityMedium,
		)
		cp.AddGate("behavior")
		cp.AddGate("adversarial")
		cp.AddTask("Implement core logic")
		cp.AddTask("Add business validations")
		cp.AddTask("Handle edge cases")
		checkpoints = append(checkpoints, cp)
	}

	// UI/Frontend
	if containsAny(featureLower, "ui", "frontend", "component", "page", "screen", "view", "form", "button") {
		cp := checkpoint.NewCheckpoint(
			uuid.New().String(),
			"User Interface",
			"Implement user interface",
			checkpoint.ComplexityMedium,
		)
		cp.AddGate("behavior")
		cp.AddGate("ui")
		cp.AddGate("adversarial")
		cp.AddTask("Create component structure")
		cp.AddTask("Implement styling")
		cp.AddTask("Add interactions")
		cp.AddTask("Test responsive design")
		checkpoints = append(checkpoints, cp)
	}

	// Auth/Permissions
	if containsAny(featureLower, "auth", "permission", "role", "access", "security") {
		cp := checkpoint.NewCheckpoint(
			uuid.New().String(),
			"Authentication & Authorization",
			"Implement authentication and access control",
			checkpoint.ComplexityHigh,
		)
		cp.AddGate("behavior")
		cp.AddGate("adversarial")
		cp.AddTask("Implement authentication")
		cp.AddTask("Add authorization checks")
		cp.AddTask("Test permission boundaries")
		checkpoints = append(checkpoints, cp)
	}

	// Integration
	if containsAny(featureLower, "integration", "external", "api client", "webhook") {
		cp := checkpoint.NewCheckpoint(
			uuid.New().String(),
			"External Integrations",
			"Integrate with external services",
			checkpoint.ComplexityHigh,
		)
		cp.AddGate("behavior")
		cp.AddGate("adversarial")
		cp.AddTask("Implement integration client")
		cp.AddTask("Handle external errors")
		cp.AddTask("Add retry logic")
		checkpoints = append(checkpoints, cp)
	}

	// If no specific pattern matched, create a single generic checkpoint
	if len(checkpoints) == 0 {
		cp := checkpoint.NewCheckpoint(
			uuid.New().String(),
			"Implementation",
			fmt.Sprintf("Implement: %s", feature),
			checkpoint.ComplexityMedium,
		)
		cp.AddGate("behavior")
		cp.AddGate("adversarial")
		cp.AddGate("ui")
		cp.AddTask("Analyze requirements")
		cp.AddTask("Design solution")
		cp.AddTask("Implement feature")
		cp.AddTask("Test thoroughly")
		checkpoints = append(checkpoints, cp)
	}

	return checkpoints, nil
}

// LLMCheckpointPlanner uses an LLM to create checkpoints
type LLMCheckpointPlanner struct {
	agent Agent
}

// NewLLMCheckpointPlanner creates a planner that uses an LLM
func NewLLMCheckpointPlanner(agent Agent) *LLMCheckpointPlanner {
	return &LLMCheckpointPlanner{agent: agent}
}

// CreateCheckpoints uses the LLM to generate checkpoints
func (p *LLMCheckpointPlanner) CreateCheckpoints(ctx context.Context, feature, description string) ([]*checkpoint.Checkpoint, error) {
	prompt := p.buildPrompt(feature, description)

	response, err := p.agent.Run(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM planning failed: %w", err)
	}

	// Parse JSON response
	return p.parseResponse(response)
}

// buildPrompt creates the prompt for the LLM
func (p *LLMCheckpointPlanner) buildPrompt(feature, description string) string {
	return fmt.Sprintf(`Analyze this feature request and create a detailed checkpoint plan.

Feature: %s
Description: %s

Requirements:
1. Break down into checkpoints ordered by increasing complexity
2. Each checkpoint should build on the previous one
3. Each checkpoint needs specific, testable tasks
4. Assign gates appropriately:
   - "behavior" for all checkpoints (automated tests)
   - "adversarial" for code-intensive checkpoints (security, logic)
   - "ui" only for checkpoints with UI components
   - "human" for final verification

Output JSON format:
{
  "checkpoints": [
    {
      "name": "Clear checkpoint name",
      "description": "What this checkpoint accomplishes",
      "complexity": "low|medium|high",
      "gates": ["behavior", "adversarial", "ui", "human"],
      "tasks": ["Specific task 1", "Specific task 2"]
    }
  ]
}

Important:
- Start with infrastructure/schema (low complexity)
- Progress to API/logic (medium complexity)
- End with integration/UI (high complexity)
- Include at least 3 checkpoints for features with multiple concerns
- Every checkpoint needs at least 3 concrete tasks
`, feature, description)
}

// parseResponse parses the LLM JSON response into checkpoints
func (p *LLMCheckpointPlanner) parseResponse(response string) ([]*checkpoint.Checkpoint, error) {
	// Try to extract JSON from markdown code blocks if present
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```") {
		// Extract JSON from code block
		lines := strings.Split(response, "\n")
		var jsonLines []string
		inBlock := false
		for _, line := range lines {
			if strings.HasPrefix(line, "```") {
				inBlock = !inBlock
				continue
			}
			if inBlock {
				jsonLines = append(jsonLines, line)
			}
		}
		response = strings.Join(jsonLines, "\n")
	}

	// Try to find JSON object
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")
	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}
	response = response[startIdx : endIdx+1]

	// Parse JSON
	var parsed struct {
		Checkpoints []struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Complexity  string   `json:"complexity"`
			Gates       []string `json:"gates"`
			Tasks       []string `json:"tasks"`
		} `json:"checkpoints"`
	}

	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(parsed.Checkpoints) == 0 {
		return nil, fmt.Errorf("no checkpoints found in response")
	}

	// Convert to checkpoint objects
	checkpoints := make([]*checkpoint.Checkpoint, len(parsed.Checkpoints))
	for i, cp := range parsed.Checkpoints {
		complexity := checkpoint.ComplexityMedium
		switch strings.ToLower(cp.Complexity) {
		case "low":
			complexity = checkpoint.ComplexityLow
		case "high":
			complexity = checkpoint.ComplexityHigh
		}

		// Default gates
		gates := cp.Gates
		if len(gates) == 0 {
			gates = []string{"behavior"}
		}

		// Default tasks
		tasks := cp.Tasks
		if len(tasks) == 0 {
			tasks = []string{"Implement feature", "Write tests", "Verify functionality"}
		}

		checkpoint := checkpoint.NewCheckpoint(
			uuid.New().String(),
			cp.Name,
			cp.Description,
			complexity,
		)

		for _, gate := range gates {
			checkpoint.AddGate(gate)
		}
		for _, task := range tasks {
			checkpoint.AddTask(task)
		}

		checkpoints[i] = checkpoint
	}

	// Explicitly sort by complexity: low -> medium -> high
	checkpoints = sortCheckpointsByComplexity(checkpoints)

	return checkpoints, nil
}

// sortCheckpointsByComplexity sorts checkpoints in increasing complexity order
// This ensures low complexity checkpoints (schema, models) come before high complexity (UI, integrations)
func sortCheckpointsByComplexity(checkpoints []*checkpoint.Checkpoint) []*checkpoint.Checkpoint {
	// Create a copy to avoid modifying original
	sorted := make([]*checkpoint.Checkpoint, len(checkpoints))
	copy(sorted, checkpoints)

	// Sort by complexity
	sort.Slice(sorted, func(i, j int) bool {
		return complexityOrder(sorted[i].Complexity) < complexityOrder(sorted[j].Complexity)
	})

	return sorted
}

// complexityOrder returns a numeric order for complexity levels
func complexityOrder(c checkpoint.ComplexityLevel) int {
	switch c {
	case checkpoint.ComplexityLow:
		return 0
	case checkpoint.ComplexityMedium:
		return 1
	case checkpoint.ComplexityHigh:
		return 2
	default:
		return 1 // medium as default
	}
}

// SmartCheckpointPlanner combines heuristics with LLM for better planning
type SmartCheckpointPlanner struct {
	llmAgent Agent
	fallback *DefaultCheckpointPlanner
}

// NewSmartCheckpointPlanner creates a planner that uses LLM with heuristic fallback
func NewSmartCheckpointPlanner(llmAgent Agent) *SmartCheckpointPlanner {
	return &SmartCheckpointPlanner{
		llmAgent: llmAgent,
		fallback: NewDefaultCheckpointPlanner(),
	}
}

// CreateCheckpoints first tries LLM, falls back to heuristics
func (p *SmartCheckpointPlanner) CreateCheckpoints(ctx context.Context, feature, description string) ([]*checkpoint.Checkpoint, error) {
	// Try LLM first
	llmPlanner := NewLLMCheckpointPlanner(p.llmAgent)
	checkpoints, err := llmPlanner.CreateCheckpoints(ctx, feature, description)
	if err == nil && len(checkpoints) > 0 {
		return checkpoints, nil
	}

	// Fall back to heuristics
	return p.fallback.CreateCheckpoints(ctx, feature, description)
}

// Helper functions

func containsAny(s string, substrs ...string) bool {
	s = strings.ToLower(s)
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
