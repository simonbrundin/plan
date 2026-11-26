package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"plan/agent/internal/core/checkpoint"
)

// TestPlanner generates tests for checkpoints
type TestPlanner interface {
	// PlanTests creates test cases for a checkpoint
	PlanTests(ctx context.Context, cp *checkpoint.Checkpoint) ([]*TestCase, error)
}

// TestCase represents a single test case
type TestCase struct {
	Name        string  // e.g., "TestUserCreation"
	Description string  // e.g., "should create user with valid data"
	Table       [][]any // Table-driven test data
}

// LLMTestPlanner uses an LLM agent to generate tests
type LLMTestPlanner struct {
	agent Agent
}

// NewLLMTestPlanner creates a test planner that uses an LLM
func NewLLMTestPlanner(agent Agent) *LLMTestPlanner {
	return &LLMTestPlanner{agent: agent}
}

// PlanTests generates test cases for a checkpoint using LLM
func (p *LLMTestPlanner) PlanTests(ctx context.Context, cp *checkpoint.Checkpoint) ([]*TestCase, error) {
	prompt := p.buildTestPrompt(cp)

	response, err := p.agent.Run(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("test planning failed: %w", err)
	}

	return p.parseTestResponse(response)
}

// buildTestPrompt creates the prompt for test generation
func (p *LLMTestPlanner) buildTestPrompt(cp *checkpoint.Checkpoint) string {
	return fmt.Sprintf(`You are a test planning agent. Generate comprehensive tests for this checkpoint.

Checkpoint: %s
Description: %s
Tasks:
%s

Requirements:
1. Generate table-driven Go tests using testing package
2. Cover happy path AND edge cases
3. Include error cases
4. Use descriptive test names: Test[Feature]_[Scenario]_[ExpectedResult]

Output format (JSON array):
{
  "tests": [
    {
      "name": "TestUserCreation_ValidInput_Success",
      "description": "should create user when input is valid",
      "table": [
        ["input", "expected"],
        ["{\"name\": \"John\"}", "success"],
        ["{\"name\": \"\"}", "error: name required"]
      ]
    }
  ]
}

Generate tests that will FAIL initially (because the feature doesn't exist yet).
This is TDD - write the tests first!
`, cp.Name, cp.Description, strings.Join(cp.Tasks, "\n- "))
}

// parseTestResponse parses the LLM response into test cases
func (p *LLMTestPlanner) parseTestResponse(response string) ([]*TestCase, error) {
	// Try to extract JSON from markdown code blocks if present
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```") {
		lines := strings.Split(response, "\n")
		var jsonLines []string
		inBlock := false
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inBlock = !inBlock
				continue
			}
			if inBlock {
				jsonLines = append(jsonLines, line)
			}
		}
		response = strings.Join(jsonLines, "\n")
	}

	// Find JSON object
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")
	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}
	response = response[startIdx : endIdx+1]

	// Simple JSON parsing for the test structure
	// Look for test entries
	var tests []*TestCase

	// Extract name, description, and table from the response
	lines := strings.Split(response, "\n")
	var currentTest *TestCase
	var inTable bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for test name
		if strings.Contains(line, "\"name\"") {
			name := extractJSONValue(line, "name")
			if name != "" {
				if currentTest != nil {
					tests = append(tests, currentTest)
				}
				currentTest = &TestCase{Name: name}
				inTable = false
			}
		}

		// Look for description
		if currentTest != nil && strings.Contains(line, "\"description\"") {
			currentTest.Description = extractJSONValue(line, "description")
		}

		// Start of table
		if strings.Contains(line, "\"table\"") || strings.Contains(line, "\"data\"") {
			inTable = true
			continue
		}

		// End of table (array closing)
		if inTable && strings.HasPrefix(line, "]") {
			inTable = false
		}

		// Parse table rows
		if inTable && currentTest != nil && strings.Contains(line, "[") && strings.Contains(line, "]") {
			// This is a table row
			row := parseTableRow(line)
			if len(row) > 0 {
				currentTest.Table = append(currentTest.Table, row)
			}
		}
	}

	if currentTest != nil {
		tests = append(tests, currentTest)
	}

	if len(tests) == 0 {
		return nil, fmt.Errorf("no tests found in response")
	}

	return tests, nil
}

// extractJSONValue extracts a string value from a JSON key-value line
func extractJSONValue(line, key string) string {
	search := fmt.Sprintf(`"%s"`, key)
	idx := strings.Index(line, search)
	if idx == -1 {
		return ""
	}

	// Find the colon and opening quote
	rest := line[idx+len(search):]
	colonIdx := strings.Index(rest, ":")
	if colonIdx == -1 {
		return ""
	}

	rest = rest[colonIdx+1:]
	rest = strings.TrimSpace(rest)

	// Check for string value
	if !strings.HasPrefix(rest, "\"") {
		return ""
	}

	rest = rest[1:] // Remove opening quote
	endIdx := strings.Index(rest, "\"")
	if endIdx == -1 {
		return ""
	}

	return rest[:endIdx]
}

// parseTableRow parses a table row like ["input", "expected"]
func parseTableRow(line string) []any {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return nil
	}

	// Simple parsing - extract quoted strings
	var values []any
	rest := strings.TrimPrefix(strings.TrimSuffix(line, "]"), "[")

	for _, part := range strings.Split(rest, ",") {
		part = strings.TrimSpace(part)
		part = strings.TrimPrefix(part, "\"")
		part = strings.TrimSuffix(part, "\"")
		if part != "" {
			values = append(values, part)
		}
	}

	return values
}

// GenerateTestFile generates a Go test file content from test cases
func GenerateTestFile(cp *checkpoint.Checkpoint, tests []*TestCase) string {
	var sb strings.Builder

	// Package and imports
	_, _ = fmt.Fprintf(&sb, "package %s\n\n", getPackageName(cp))
	sb.WriteString("import (\n")
	sb.WriteString("\t\"testing\"\n")
	sb.WriteString(")\n\n")

	// Generate test functions
	for _, tc := range tests {
		_, _ = fmt.Fprintf(&sb, "// %s\n", tc.Description)
		_, _ = fmt.Fprintf(&sb, "func %s(t *testing.T) {\n", tc.Name)

		if len(tc.Table) > 0 {
			// Table-driven test
			sb.WriteString("\ttestCases := []struct {\n")

			// Headers from first row
			if len(tc.Table) > 0 {
				for _, cell := range tc.Table[0] {
					_, _ = fmt.Fprintf(&sb, "\t\t%s string\n", sanitizeName(fmt.Sprintf("%v", cell)))
				}
			}
			sb.WriteString("\t}{\n")

			// Test data rows (skip header)
			for i, row := range tc.Table {
				if i == 0 {
					continue // Skip header
				}
				sb.WriteString("\t\t{")
				for j, cell := range row {
					if j > 0 {
						sb.WriteString(", ")
					}
					_, _ = fmt.Fprintf(&sb, "%s: %q", sanitizeName(fmt.Sprintf("%v", tc.Table[0][j])), fmt.Sprintf("%v", cell))
				}
				sb.WriteString("},\n")
			}
			sb.WriteString("\t}\n\n")

			sb.WriteString("\tfor _, tc := range testCases {\n")
			sb.WriteString("\t\ttc := tc // capture range variable\n")
			sb.WriteString("\t\tt.Run(tc.Description, func(t *testing.T) {\n")
			sb.WriteString("\t\t\t// TODO: implement test logic\n")
			sb.WriteString("\t\t\tt.Skip(\"not implemented yet\")\n")
			sb.WriteString("\t\t})\n")
			sb.WriteString("\t}\n")
		} else {
			sb.WriteString("\t// TODO: implement test\n")
			sb.WriteString("\tt.Skip(\"not implemented yet\")\n")
		}

		sb.WriteString("}\n\n")
	}

	return sb.String()
}

// getPackageName derives a package name from checkpoint name
func getPackageName(cp *checkpoint.Checkpoint) string {
	name := strings.ToLower(cp.Name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

// sanitizeName makes a string a valid Go identifier
func sanitizeName(s string) string {
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return '_'
	}, s)

	// Ensure it starts with a letter
	if len(s) > 0 && s[0] >= '0' && s[0] <= '9' {
		s = "_" + s
	}

	return s
}

// DefaultTestPlanner is a fallback planner that generates basic tests
type DefaultTestPlanner struct{}

// NewDefaultTestPlanner creates a default test planner
func NewDefaultTestPlanner() *DefaultTestPlanner {
	return &DefaultTestPlanner{}
}

// PlanTests generates basic tests for a checkpoint
func (p *DefaultTestPlanner) PlanTests(ctx context.Context, cp *checkpoint.Checkpoint) ([]*TestCase, error) {
	// Generate a basic test case for each task
	var tests []*TestCase

	for i, task := range cp.Tasks {
		test := &TestCase{
			Name:        fmt.Sprintf("TestTask%d", i+1),
			Description: task,
			Table: [][]any{
				{"input", "expected"},
				{"valid input", "success"},
				{"invalid input", "error"},
			},
		}
		tests = append(tests, test)
	}

	_ = ctx // not used

	return tests, nil
}

// Placeholder time for compatibility
var _ = time.Now
