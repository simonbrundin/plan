package gate

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"plan-agent/internal/core/checkpoint"
)

// BehaviorGate checks that the feature works correctly through automated tests
type BehaviorGate struct {
	testCommand string
	lintCommand string
}

// NewBehaviorGate creates a new behavior gate
func NewBehaviorGate() *BehaviorGate {
	return &BehaviorGate{
		testCommand: "go test ./...",
		lintCommand: "golangci-lint run",
	}
}

// NewBehaviorGateWithConfig creates a behavior gate with custom commands
func NewBehaviorGateWithConfig(testCmd, lintCmd string) *BehaviorGate {
	return &BehaviorGate{
		testCommand: testCmd,
		lintCommand: lintCmd,
	}
}

// Gate interface implementation
func (g *BehaviorGate) Type() GateType { return GateTypeBehavior }
func (g *BehaviorGate) Name() string   { return "Behavior Gate" }
func (g *BehaviorGate) Required() bool { return true }

func (g *BehaviorGate) Description() string {
	return "Verifies that the feature works correctly by running automated tests and linting"
}

// Run executes the behavior gate check
func (g *BehaviorGate) Run(ctx context.Context, cp *checkpoint.Checkpoint) (*checkpoint.GateResult, error) {
	result := checkpoint.NewGateResult(string(GateTypeBehavior))
	result.Status = checkpoint.GateStatusRunning

	var allIssues []checkpoint.Issue

	// Run tests
	if issues, err := g.runTests(ctx); err != nil {
		allIssues = append(allIssues, issues...)
		result.Fail(fmt.Sprintf("Tests failed: %v", err), allIssues)
		return result, nil
	}

	// Run linting
	if issues, err := g.runLint(ctx); err != nil {
		allIssues = append(allIssues, issues...)
		result.Fail(fmt.Sprintf("Linting failed: %v", err), allIssues)
		return result, nil
	}

	// Run vet
	if issues, err := g.runVet(ctx); err != nil {
		allIssues = append(allIssues, issues...)
		result.Fail(fmt.Sprintf("Vet failed: %v", err), allIssues)
		return result, nil
	}

	if len(allIssues) > 0 {
		result.Fail("Found issues in tests/lint/vet", allIssues)
	} else {
		result.Pass("All behavior checks passed: tests, lint, and vet")
	}

	return result, nil
}

func (g *BehaviorGate) runTests(ctx context.Context) ([]checkpoint.Issue, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", g.testCommand)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return parseTestIssues(string(output)), err
	}
	return nil, nil
}

func (g *BehaviorGate) runLint(ctx context.Context) ([]checkpoint.Issue, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", g.lintCommand)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return parseLintIssues(string(output)), err
	}
	return nil, nil
}

func (g *BehaviorGate) runVet(ctx context.Context) ([]checkpoint.Issue, error) {
	cmd := exec.CommandContext(ctx, "go", "vet", "./...")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return parseVetIssues(string(output)), err
	}
	return nil, nil
}

// --- Output Parsing ---

func parseTestIssues(output string) []checkpoint.Issue {
	var issues []checkpoint.Issue
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "FAIL") || strings.Contains(line, "--- FAIL") {
			issues = append(issues, checkpoint.Issue{
				Severity:    checkpoint.SeverityHigh,
				Category:    "test",
				Title:       "Test failure",
				Description: strings.TrimSpace(line),
			})
		}
	}
	if len(issues) == 0 {
		issues = []checkpoint.Issue{{
			Severity:    checkpoint.SeverityHigh,
			Category:    "test",
			Title:       "Test failure",
			Description: output,
		}}
	}
	return issues
}

func parseLintIssues(output string) []checkpoint.Issue {
	var issues []checkpoint.Issue
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// golangci-lint format: file:line:col: message
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 4)
			if len(parts) >= 3 {
				issues = append(issues, checkpoint.Issue{
					Severity:    checkpoint.SeverityMedium,
					Category:    "lint",
					Title:       "Linting issue",
					Description: getOrDefault(parts, 3, ""),
					Location:    parts[0] + ":" + parts[1],
				})
			}
		}
	}
	if len(issues) == 0 {
		issues = []checkpoint.Issue{{
			Severity:    checkpoint.SeverityMedium,
			Category:    "lint",
			Title:       "Linting issue",
			Description: output,
		}}
	}
	return issues
}

func parseVetIssues(output string) []checkpoint.Issue {
	var issues []checkpoint.Issue
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) >= 2 {
			issues = append(issues, checkpoint.Issue{
				Severity:    checkpoint.SeverityMedium,
				Category:    "vet",
				Title:       "Go vet issue",
				Description: strings.TrimSpace(parts[1]),
				Location:    strings.TrimSpace(parts[0]),
			})
		}
	}
	if len(issues) == 0 {
		issues = []checkpoint.Issue{{
			Severity:    checkpoint.SeverityMedium,
			Category:    "vet",
			Title:       "Go vet issue",
			Description: output,
		}}
	}
	return issues
}

func getOrDefault(parts []string, idx int, defaultVal string) string {
	if idx < len(parts) {
		return parts[idx]
	}
	return defaultVal
}
