package gate

import (
	"fmt"
	"strings"

	"plan-agent/internal/core/checkpoint"
)

func buildAdversarialCriticPrompt(cp *checkpoint.Checkpoint, standardsFile string) string {
	standards := ""
	if standardsFile != "" {
		standards = fmt.Sprintf("\n\nCode standards to verify against:\nFile: %s", standardsFile)
	}

	return fmt.Sprintf(`You are an adversarial code reviewer. Your job is to find problems in the code.

Checkpoint: %s
Description: %s
Tasks: %s

Review the code thoroughly and list any issues you find. Be critical.

Check for:
- Security vulnerabilities
- Performance problems
- Error handling issues
- Code smells
- Violations of best practices
- Missing edge case handling
- API design issues
- Data consistency problems
%s

Respond in this format:
ISSUES FOUND: [yes/no]
VERDICT: PASS/FAIL

If issues found:
- [severity] [file:line]: [issue title]
  [description]
  [suggestion]
`, cp.Name, cp.Description, strings.Join(cp.Tasks, "\n- "), standards)
}

func buildAdversarialFixerPrompt(cp *checkpoint.Checkpoint, issues []checkpoint.Issue) string {
	var issuesList strings.Builder
	for _, issue := range issues {
		issuesList.WriteString(fmt.Sprintf("- [%s] %s\n  Location: %s\n  %s\n  Suggestion: %s\n\n",
			issue.Severity, issue.Title, issue.Location, issue.Description, issue.Suggestion))
	}

	return fmt.Sprintf(`You are a code fixer agent. Your job is to fix the issues found by the critic.

Checkpoint: %s

Issues to fix:
%s

For each issue:
1. Understand the problem
2. Apply the fix
3. Verify the fix makes sense

Respond with:
FIXES APPLIED:
- [issue]: [what was changed]

If you cannot fix an issue, explain why.
`, cp.Name, issuesList.String())
}

func parseAdversarialIssues(output string) []checkpoint.Issue {
	if !containsIgnoreCase(output, "ISSUES FOUND: yes") && !containsIgnoreCase(output, "VERDICT: FAIL") {
		return nil
	}

	var issues []checkpoint.Issue
	for i, line := range splitLines(output) {
		line = trimSpace(line)
		if startsWith(line, "-") && containsSeverity(line) {
			if issue := parseAdversarialIssue(line, splitLines(output), i); issue.Title != "" {
				issues = append(issues, issue)
			}
		}
	}
	return issues
}

func containsSeverity(line string) bool {
	lower := toLower(line)
	return contains(lower, "critical") || contains(lower, "high") ||
		contains(lower, "medium") || contains(lower, "low")
}

func parseAdversarialIssue(line string, lines []string, index int) checkpoint.Issue {
	issue := checkpoint.Issue{Description: line}

	// Determine severity
	lower := toLower(line)
	switch {
	case contains(lower, "critical"):
		issue.Severity = checkpoint.SeverityCritical
	case contains(lower, "high"):
		issue.Severity = checkpoint.SeverityHigh
	case contains(lower, "medium"):
		issue.Severity = checkpoint.SeverityMedium
	default:
		issue.Severity = checkpoint.SeverityLow
	}

	// Extract location (file:line)
	if colonIdx := strings.Index(line, ":"); colonIdx > 0 {
		issue.Location = trimSpace(line[:colonIdx])
		issue.Title = trimSpace(line[colonIdx+1:])
	} else {
		issue.Title = stripSeverityPrefix(line)
	}

	// Look for suggestion in following lines
	if index+1 < len(lines) {
		nextLine := trimSpace(lines[index+1])
		if !startsWith(nextLine, "-") && !contains(nextLine, ":") && len(nextLine) > 0 {
			issue.Suggestion = nextLine
		}
	}

	return issue
}

func stripSeverityPrefix(line string) string {
	title := line
	for _, prefix := range []string{"critical", "high", "medium", "low"} {
		title = strings.TrimPrefix(toLower(title), prefix+" ")
		title = strings.TrimPrefix(title, prefix+":")
		title = strings.TrimPrefix(title, prefix+" ")
	}
	return trimSpace(title)
}
