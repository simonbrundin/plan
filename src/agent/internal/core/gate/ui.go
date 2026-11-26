package gate

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"plan/agent/internal/core/checkpoint"
)

// UIGate checks that the UI matches the design/prototype
type UIGate struct {
	prototypePath string // Path to HTML prototype
	designFile    string // Path to design specification
	agentFactory  AgentFactory
}

// AgentFactory creates agents for sub-tasks
type AgentFactory interface {
	CreateAgent(config AgentConfig) Agent
}

// AgentConfig contains configuration for creating an agent
type AgentConfig struct {
	Name         string
	SystemPrompt string
	Tools        []string
}

// Agent represents an agent that can be spawned
type Agent interface {
	Run(ctx context.Context, prompt string) (string, error)
}

// NewUIGate creates a new UI gate
func NewUIGate(prototypePath, designFile string) *UIGate {
	return &UIGate{
		prototypePath: prototypePath,
		designFile:    designFile,
	}
}

// NewUIGateWithFactory creates a UI gate with a custom agent factory
func NewUIGateWithFactory(prototypePath, designFile string, factory AgentFactory) *UIGate {
	return &UIGate{
		prototypePath: prototypePath,
		designFile:    designFile,
		agentFactory:  factory,
	}
}

// Type returns the gate type
func (g *UIGate) Type() GateType { return GateTypeUI }

// Name returns a human-readable name
func (g *UIGate) Name() string { return "UI Gate" }

// Description returns a description of what this gate checks
func (g *UIGate) Description() string {
	return "Verifies that the UI matches the design and prototype specifications"
}

// Required returns true if this gate is required for all checkpoints
func (g *UIGate) Required() bool { return false }

// Run executes the UI gate check
func (g *UIGate) Run(ctx context.Context, cp *checkpoint.Checkpoint) (*checkpoint.GateResult, error) {
	result := checkpoint.NewGateResult(string(GateTypeUI))
	result.Status = checkpoint.GateStatusRunning

	if !g.hasUIComponent(cp) {
		result.Skip("No UI components in this checkpoint")
		return result, nil
	}

	prototypePath := g.getPrototypePath(cp)
	designFile := g.getDesignFile(cp)

	if prototypePath == "" && designFile == "" {
		result.Skip("No prototype or design file defined - UI gate skipped")
		return result, nil
	}

	if g.agentFactory != nil {
		return g.runWithAgents(ctx, cp, result, prototypePath, designFile)
	}

	result.Skip("UI gate requires agent factory for automated comparison")
	return result, nil
}

// getPrototypePath returns prototype path from checkpoint metadata or default
func (g *UIGate) getPrototypePath(cp *checkpoint.Checkpoint) string {
	if cp.Metadata != nil {
		if path, ok := cp.Metadata["prototypePath"]; ok {
			return path
		}
	}
	return g.prototypePath
}

// getDesignFile returns design file path from checkpoint metadata or default
func (g *UIGate) getDesignFile(cp *checkpoint.Checkpoint) string {
	if cp.Metadata != nil {
		if path, ok := cp.Metadata["designFile"]; ok {
			return path
		}
	}
	return g.designFile
}

// hasUIComponent checks if the checkpoint involves UI work
func (g *UIGate) hasUIComponent(cp *checkpoint.Checkpoint) bool {
	uiKeywords := []string{"ui", "frontend", "component", "page", "view", "screen", "button", "form", "input", "display"}
	desc := cp.Description + " " + cp.Name

	for _, keyword := range uiKeywords {
		if containsIgnoreCase(desc, keyword) {
			return true
		}
	}
	return false
}

// runWithAgents runs look and behavior review agents in parallel
func (g *UIGate) runWithAgents(ctx context.Context, cp *checkpoint.Checkpoint, result *checkpoint.GateResult, prototypePath, designFile string) (*checkpoint.GateResult, error) {
	var (
		lookIssues     []checkpoint.Issue
		behaviorIssues []checkpoint.Issue
		lookErr        error
		behaviorErr    error
		mu             sync.Mutex
		wg             sync.WaitGroup
	)

	lookAgent := g.agentFactory.CreateAgent(AgentConfig{
		Name:         "look-reviewer",
		SystemPrompt: getLookReviewSystemPrompt(),
	})

	behaviorAgent := g.agentFactory.CreateAgent(AgentConfig{
		Name:         "behavior-reviewer",
		SystemPrompt: getBehaviorReviewSystemPrompt(),
	})

	// Run look review
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := lookAgent.Run(ctx, buildLookReviewPrompt(cp, designFile))
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			lookErr = err
		} else {
			lookIssues = parseReviewOutput(output, "ui-look")
		}
	}()

	// Run behavior review
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := behaviorAgent.Run(ctx, buildBehaviorReviewPrompt(cp, prototypePath))
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			behaviorErr = err
		} else {
			behaviorIssues = parseReviewOutput(output, "ui-behavior")
		}
	}()

	wg.Wait()

	allIssues := append(lookIssues, behaviorIssues...)

	if (lookErr != nil || behaviorErr != nil) && len(allIssues) == 0 {
		result.Fail(buildErrorMessage(lookErr, behaviorErr), nil)
		return result, nil
	}

	return evaluateGateResult(result, allIssues), nil
}

func buildErrorMessage(lookErr, behaviorErr error) string {
	msg := "UI review failed:"
	if lookErr != nil {
		msg += fmt.Sprintf(" look: %v", lookErr)
	}
	if behaviorErr != nil {
		msg += fmt.Sprintf(" behavior: %v", behaviorErr)
	}
	return msg
}

func evaluateGateResult(result *checkpoint.GateResult, allIssues []checkpoint.Issue) *checkpoint.GateResult {
	var blocking []checkpoint.Issue
	for _, issue := range allIssues {
		if issue.Severity == checkpoint.SeverityHigh || issue.Severity == checkpoint.SeverityCritical {
			blocking = append(blocking, issue)
		}
	}

	if len(blocking) > 0 {
		result.Fail("UI has blocking issues", blocking)
	} else if len(allIssues) > 0 {
		result.Pass(fmt.Sprintf("UI matches with %d minor issues", len(allIssues)))
		result.Issues = allIssues
	} else {
		result.Pass("UI matches prototype specification")
	}
	return result
}

func buildLookReviewPrompt(cp *checkpoint.Checkpoint, designFile string) string {
	designInfo := ""
	if designFile != "" {
		designInfo = fmt.Sprintf("\nDesign file: %s", designFile)
	}
	return fmt.Sprintf(`Review the UI implementation for checkpoint "%s" against the design specification.

Description: %s
Tasks: %s
%s

Compare these visual aspects:
1. Layout and structure
2. Spacing and sizing (padding, margins, widths, heights)
3. Colors and backgrounds
4. Typography (font sizes, weights, families)
5. Alignment and positioning
6. Visual hierarchy

Be extremely detail-oriented. Even small differences matter for visual consistency.
`, cp.Name, cp.Description, strings.Join(cp.Tasks, ", "), designInfo)
}

func buildBehaviorReviewPrompt(cp *checkpoint.Checkpoint, prototypePath string) string {
	prototypeInfo := ""
	if prototypePath != "" {
		prototypeInfo = fmt.Sprintf("\nPrototype: %s", prototypePath)
	}
	return fmt.Sprintf(`Review the UI implementation for checkpoint "%s" against the prototype.

Description: %s
Tasks: %s
%s

Compare these behavioral aspects:
1. Click/tap interactions
2. Hover states
3. Focus states
4. Transitions and animations
5. Form validation behavior
6. Loading states
7. Error states
8. State changes (before/after interactions)

IMPORTANT: Verify that both the prototype AND the app are in the SAME STATE before comparing.
`, cp.Name, cp.Description, strings.Join(cp.Tasks, ", "), prototypeInfo)
}

func getLookReviewSystemPrompt() string {
	return `You are a visual design perfectionist. You review UI implementations with extreme attention to detail.

You check for: spacing differences, sizing differences, position differences, color differences, font differences, visual hierarchy differences.

Respond in this format:
VERDICT: PASS/FAIL

If FAIL, list issues:
ISSUES:
- [high/critical] [location]: [specific visual difference found]
- ...

Be specific. "Spacing is off" is not helpful. "Button padding is 8px but should be 12px" is helpful.`
}

func getBehaviorReviewSystemPrompt() string {
	return `You are an interaction design expert. You verify that UI behavior matches prototypes.

You check for: click/tap handlers, hover states, focus states, transitions, form validation, error messages, loading indicators, state transitions.

IMPORTANT: First verify both prototype and app are in the SAME STATE.

Respond in this format:
VERDICT: PASS/FAIL

If FAIL, list issues:
ISSUES:
- [high/critical] [element]: [expected behavior] vs [actual behavior]
- ...

Be specific. "Behavior is wrong" is not helpful. "Submit button does not show loading spinner when clicked" is helpful.`
}

// parseReviewOutput parses review output for both look and behavior reviews
func parseReviewOutput(output string, category string) []checkpoint.Issue {
	if containsIgnoreCase(output, "VERDICT: PASS") {
		return nil
	}

	var issues []checkpoint.Issue
	lines := strings.Split(output, "\n")
	inIssues := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if containsIgnoreCase(line, "ISSUES:") {
			inIssues = true
			continue
		}

		if inIssues && (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "•")) {
			issues = append(issues, parseIssueLine(line, category))
		}
	}
	return issues
}

func parseIssueLine(line string, category string) checkpoint.Issue {
	severity := checkpoint.SeverityMedium
	if containsIgnoreCase(line, "critical") {
		severity = checkpoint.SeverityCritical
	} else if containsIgnoreCase(line, "high") {
		severity = checkpoint.SeverityHigh
	}

	return checkpoint.Issue{
		Severity:    severity,
		Category:    category,
		Title:       strings.TrimSpace(line[1:]),
		Description: line,
	}
}
