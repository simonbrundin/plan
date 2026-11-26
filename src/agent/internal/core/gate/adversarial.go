package gate

import (
	"context"
	"fmt"
	"strings"

	"plan/agent/internal/core/checkpoint"
)

// AdversarialGate performs adversarial code review
type AdversarialGate struct {
	standardsFile string
	maxIterations int
	criticAgent   Agent
	fixerAgent    Agent
}

func NewAdversarialGate() *AdversarialGate {
	return &AdversarialGate{maxIterations: 3}
}

func NewAdversarialGateWithAgents(critic, fixer Agent, maxIterations int) *AdversarialGate {
	return &AdversarialGate{
		criticAgent:   critic,
		fixerAgent:    fixer,
		maxIterations: maxIterations,
	}
}

// Gate interface implementation
func (g *AdversarialGate) Type() GateType { return GateTypeAdversarial }
func (g *AdversarialGate) Name() string   { return "Adversarial Code Review Gate" }
func (g *AdversarialGate) Description() string {
	return "Verifies code quality through adversarial review"
}
func (g *AdversarialGate) Required() bool { return true }

// Run executes the adversarial gate check
func (g *AdversarialGate) Run(ctx context.Context, cp *checkpoint.Checkpoint) (*checkpoint.GateResult, error) {
	result := checkpoint.NewGateResult(string(GateTypeAdversarial))
	result.Status = checkpoint.GateStatusRunning

	if g.criticAgent == nil {
		result.Skip("Adversarial gate requires agent configuration")
		return result, nil
	}

	var allIssues []checkpoint.Issue

	for iteration := 0; iteration < g.maxIterations; iteration++ {
		criticReview, err := g.criticAgent.Run(ctx, buildAdversarialCriticPrompt(cp, g.standardsFile))
		if err != nil {
			result.Fail(fmt.Sprintf("Critic agent failed: %v", err), allIssues)
			return result, nil
		}

		issues := parseAdversarialIssues(criticReview)
		if len(issues) == 0 {
			result.Pass(fmt.Sprintf("Adversarial review passed after %d iteration(s)", iteration+1))
			return result, nil
		}

		allIssues = append(allIssues, issues...)

		if !hasBlockingIssues(issues) {
			result.Pass("Only non-blocking issues found")
			return result, nil
		}

		fixerResponse, err := g.fixerAgent.Run(ctx, buildAdversarialFixerPrompt(cp, issues))
		if err != nil {
			result.Fail(fmt.Sprintf("Fixer agent failed: %v", err), allIssues)
			return result, nil
		}

		fixResult := applyFixerPatches(fixerResponse)
		if len(fixResult.Failed) > 0 {
			result.Issues = append(result.Issues, checkpoint.Issue{
				Title:       "Fix application failed",
				Description: strings.Join(fixResult.Failed, "; "),
				Severity:    checkpoint.SeverityMedium,
			})
		}
		if len(fixResult.Applied) > 0 {
			result.Feedback = fmt.Sprintf("Applied fixes: %s", strings.Join(fixResult.Applied, ", "))
		}
	}

	result.Fail(fmt.Sprintf("Max iterations (%d) reached with unresolved issues", g.maxIterations), allIssues)
	return result, nil
}

func hasBlockingIssues(issues []checkpoint.Issue) bool {
	for _, issue := range issues {
		if issue.Severity == checkpoint.SeverityCritical || issue.Severity == checkpoint.SeverityHigh {
			return true
		}
	}
	return false
}
