package orchestrator

import (
	"fmt"
	"strings"

	"plan-agent/internal/core/checkpoint"
)

// ConsoleViewer displays checkpoints to console
type ConsoleViewer struct{}

// NewConsoleViewer creates a console viewer
func NewConsoleViewer() *ConsoleViewer {
	return &ConsoleViewer{}
}

// ShowPlan displays a plan for approval
func (v *ConsoleViewer) ShowPlan(plan *checkpoint.CheckpointPlan) {
	sep := strings.Repeat("=", 70)
	fmt.Printf("\n%s\n", sep)
	fmt.Printf("CHECKPOINT PLAN: %s\n", plan.Feature)
	fmt.Printf("%s\n\n", sep)
	fmt.Printf("Description: %s\n\n", plan.Description)
	fmt.Printf("Checkpoints (%d):\n", len(plan.Checkpoints))

	for i, cp := range plan.Checkpoints {
		fmt.Printf("\n  %d. %s [%s]\n", i+1, cp.Name, cp.Complexity)
		fmt.Printf("     %s\n", cp.Description)
		if len(cp.Tasks) > 0 {
			fmt.Printf("     Tasks:\n")
			for _, task := range cp.Tasks {
				fmt.Printf("       • %s\n", task)
			}
		}
		if len(cp.Gates) > 0 {
			fmt.Printf("     Gates: %s\n", strings.Join(cp.Gates, ", "))
		}
	}

	sepDash := strings.Repeat("-", 70)
	fmt.Printf("\n%s\n", sepDash)
	fmt.Printf("Review and approve this plan before execution.\n")
	fmt.Printf("%s\n\n", sep)
}

// ShowCheckpoint displays a checkpoint when it starts
func (v *ConsoleViewer) ShowCheckpoint(cp *checkpoint.Checkpoint) {
	sep := strings.Repeat("▶", 70)
	fmt.Printf("\n%s\n", sep)
	fmt.Printf("CHECKPOINT: %s\n", cp.Name)
	fmt.Printf("%s\n\n", sep)
	fmt.Printf("Description: %s\n", cp.Description)
	fmt.Printf("Complexity: %s\n", cp.Complexity)
	if len(cp.Tasks) > 0 {
		fmt.Printf("\nTasks:\n")
		for i, task := range cp.Tasks {
			fmt.Printf("  %d. %s\n", i+1, task)
		}
	}
	fmt.Printf("\n")
}

// UpdateCheckpoint displays updated checkpoint status
func (v *ConsoleViewer) UpdateCheckpoint(cp *checkpoint.Checkpoint) {
	// Status indicator
	statusIcon := "⏳"
	switch cp.Status {
	case checkpoint.CheckpointPassed:
		statusIcon = "✅"
	case checkpoint.CheckpointFailed:
		statusIcon = "❌"
	case checkpoint.CheckpointInProgress:
		statusIcon = "🔄"
	}

	fmt.Printf("\n  %s Checkpoint: %s - %s\n", statusIcon, cp.Name, cp.Status)

	// Gate status
	if len(cp.GateResults) > 0 {
		fmt.Printf("\n  Gate Status:\n")
		for _, gateType := range cp.Gates {
			result, exists := cp.GateResults[gateType]
			if !exists {
				fmt.Printf("    ○ %s: pending\n", gateType)
				continue
			}

			icon := "○"
			switch result.Status {
			case checkpoint.GateStatusPassed:
				icon = "✅"
			case checkpoint.GateStatusFailed:
				icon = "❌"
			case checkpoint.GateStatusRunning:
				icon = "🔄"
			case checkpoint.GateStatusSkipped:
				icon = "⏭️"
			}

			duration := ""
			if result.Duration != "" {
				duration = fmt.Sprintf(" (%s)", result.Duration)
			}

			fmt.Printf("    %s %s: %s%s\n", icon, gateType, result.Status, duration)
		}
	}
}

// HTMLViewer generates HTML for web display
type HTMLViewer struct {
	baseURL string
}

// NewHTMLViewer creates an HTML viewer
func NewHTMLViewer(baseURL string) *HTMLViewer {
	return &HTMLViewer{baseURL: baseURL}
}

// ShowPlan generates HTML for plan display
func (v *HTMLViewer) ShowPlan(plan *checkpoint.CheckpointPlan) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html>
<html>
<head>
  <title>Checkpoint Plan</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
    .checkpoint { border: 1px solid #ddd; border-radius: 8px; padding: 16px; margin: 12px 0; }
    .complexity-low { border-left: 4px solid #10b981; }
    .complexity-medium { border-left: 4px solid #f59e0b; }
    .complexity-high { border-left: 4px solid #ef4444; }
    .gate { display: inline-block; padding: 4px 8px; margin: 4px; border-radius: 4px; background: #f3f4f6; }
    .gate-passed { background: #d1fae5; }
    .gate-failed { background: #fee2e2; }
    .btn { padding: 8px 16px; border-radius: 6px; border: none; cursor: pointer; font-size: 14px; }
    .btn-approve { background: #10b981; color: white; }
    .btn-reject { background: #ef4444; color: white; }
  </style>
</head>
<body>
  <h1>` + plan.Feature + `</h1>
  <p>` + plan.Description + `</p>
`)

	for i, cp := range plan.Checkpoints {
		_, _ = fmt.Fprintf(&b, `  <div class="checkpoint complexity-%s">
    <h3>%d. %s</h3>
    <p>%s</p>
`, cp.Complexity, i+1, cp.Name, cp.Description)

		if len(cp.Gates) > 0 {
			b.WriteString("    <div>Gates: ")
			for _, g := range cp.Gates {
				_, _ = fmt.Fprintf(&b, `<span class="gate">%s</span>`, g)
			}
			b.WriteString("</div>\n")
		}

		if len(cp.Tasks) > 0 {
			b.WriteString("    <ul>")
			for _, task := range cp.Tasks {
				b.WriteString("<li>" + task + "</li>")
			}
			b.WriteString("</ul>\n")
		}

		b.WriteString("  </div>\n")
	}

	b.WriteString(`
  <div style="margin-top: 24px;">
    <button class="btn btn-approve" onclick="approve()">Approve Plan</button>
    <button class="btn btn-reject" onclick="reject()">Reject Plan</button>
  </div>
  <script>
    function approve() { window.location.href = '/api/v1/plans/` + plan.ID + `/approve'; }
    function reject() { window.location.href = '/api/v1/plans/` + plan.ID + `/reject'; }
  </script>
</body>
</html>`)

	return b.String()
}

// UpdateCheckpoint generates HTML for checkpoint update
func (v *HTMLViewer) UpdateCheckpoint(cp *checkpoint.Checkpoint) string {
	// Simplified - in production would generate just the delta
	var b strings.Builder

	statusIcon := "⏳"
	switch cp.Status {
	case checkpoint.CheckpointPassed:
		statusIcon = "✅"
	case checkpoint.CheckpointFailed:
		statusIcon = "❌"
	case checkpoint.CheckpointInProgress:
		statusIcon = "🔄"
	}

	_, _ = fmt.Fprintf(&b, `<div class="checkpoint" id="cp-%s">
  <h3>%s %s</h3>
`, cp.ID, statusIcon, cp.Name)

	for _, gateType := range cp.Gates {
		result, exists := cp.GateResults[gateType]
		class := ""
		icon := "○"

		if exists {
			switch result.Status {
			case checkpoint.GateStatusPassed:
				icon = "✅"
				class = "gate-passed"
			case checkpoint.GateStatusFailed:
				icon = "❌"
				class = "gate-failed"
			case checkpoint.GateStatusRunning:
				icon = "🔄"
			}
		}

		_, _ = fmt.Fprintf(&b, `  <span class="gate %s">%s %s</span>`, class, icon, gateType)
	}

	b.WriteString("</div>\n")
	return b.String()
}
