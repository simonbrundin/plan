package orchestrator

import (
	"fmt"
	"strings"
	"time"

	"plan-agent/internal/core/checkpoint"
)

// HTMLViewerV2 generates enhanced HTML visualization of checkpoint plans
type HTMLViewerV2 struct{}

// NewHTMLViewerV2 creates a new HTML viewer v2
func NewHTMLViewerV2() *HTMLViewerV2 {
	return &HTMLViewerV2{}
}

// ViewPlan generates an HTML page for a checkpoint plan
func (v *HTMLViewerV2) ViewPlan(plan *checkpoint.CheckpointPlan) string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Checkpoint Plan: ` + escapeHTML(plan.Feature) + `</title>
	<style>
		* { box-sizing: border-box; margin: 0; padding: 0; }
		body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; background: #0f172a; color: #e2e8f0; padding: 2rem; }
		.container { max-width: 1200px; margin: 0 auto; }
		h1 { font-size: 2rem; margin-bottom: 0.5rem; color: #f1f5f9; }
		.meta { color: #94a3b8; margin-bottom: 2rem; }
		.checkpoint { background: #1e293b; border-radius: 12px; padding: 1.5rem; margin-bottom: 1.5rem; border-left: 4px solid #475569; }
		.checkpoint.pending { border-left-color: #94a3b8; }
		.checkpoint.in_progress { border-left-color: #f59e0b; }
		.checkpoint.passed { border-left-color: #22c55e; }
		.checkpoint.failed { border-left-color: #ef4444; }
		.checkpoint-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
		.checkpoint-name { font-size: 1.25rem; font-weight: 600; }
		.status-badge { padding: 0.25rem 0.75rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; text-transform: uppercase; }
		.status-badge.pending { background: #475569; color: #cbd5e1; }
		.status-badge.in_progress { background: #f59e0b33; color: #fbbf24; }
		.status-badge.passed { background: #22c55e33; color: #4ade80; }
		.status-badge.failed { background: #ef444433; color: #f87171; }
		.description { color: #94a3b8; margin-bottom: 1rem; }
		.complexity { display: inline-block; padding: 0.125rem 0.5rem; border-radius: 4px; font-size: 0.75rem; background: #334155; color: #94a3b8; }
		.tasks, .gates { margin-top: 1rem; }
		.tasks h3, .gates h3 { font-size: 0.875rem; color: #64748b; margin-bottom: 0.5rem; text-transform: uppercase; letter-spacing: 0.05em; }
		.task, .gate { padding: 0.5rem 0.75rem; background: #0f172a; border-radius: 6px; margin-bottom: 0.5rem; display: flex; align-items: center; gap: 0.5rem; }
		.gate { justify-content: space-between; }
		.gate-status { font-size: 0.875rem; }
		.gate-status.passed { color: #4ade80; }
		.gate-status.failed { color: #f87171; }
		.gate-status.pending { color: #fbbf24; }
		.gate-status.running { color: #60a5fa; }
		.feedback { font-size: 0.875rem; color: #f87171; margin-top: 0.5rem; padding: 0.5rem; background: #ef444433; border-radius: 6px; }
		.progress-bar { height: 8px; background: #334155; border-radius: 4px; overflow: hidden; margin: 1rem 0; }
		.progress-fill { height: 100%; background: linear-gradient(90deg, #22c55e, #4ade80); transition: width 0.3s ease; }
		.progress-text { text-align: center; color: #94a3b8; font-size: 0.875rem; margin-bottom: 2rem; }
		.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; padding-bottom: 1rem; border-bottom: 1px solid #334155; }
		.logo { font-weight: 700; color: #22c55e; }
		.timestamp { color: #64748b; font-size: 0.875rem; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<div class="logo">Checkpoint Plan</div>
			<div class="timestamp">Updated: ` + formatTimestamp(plan.UpdatedAt) + `</div>
		</div>
		<h1>` + escapeHTML(plan.Feature) + `</h1>
		<div class="meta">` + escapeHTML(plan.Description) + `</div>
`)

	// Progress bar
	passed := 0
	total := len(plan.Checkpoints)
	for _, cp := range plan.Checkpoints {
		if cp.Status == checkpoint.CheckpointPassed {
			passed++
		}
	}
	percent := 0
	if total > 0 {
		percent = (passed * 100) / total
	}
	_, _ = fmt.Fprintf(&sb, `
		<div class="progress-bar"><div class="progress-fill" style="width: %d%%"></div></div>
		<div class="progress-text">%d of %d checkpoints complete</div>
`, percent, passed, total)

	// Checkpoints
	for i, cp := range plan.Checkpoints {
		sb.WriteString(v.viewCheckpoint(cp, i+1))
	}

	sb.WriteString(`
	</div>
</body>
</html>`)

	return sb.String()
}

// viewCheckpoint generates HTML for a single checkpoint
func (v *HTMLViewerV2) viewCheckpoint(cp *checkpoint.Checkpoint, number int) string {
	status := strings.ToLower(string(cp.Status))
	description := escapeHTML(cp.Description)
	if description == "" {
		description = "No description"
	}

	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, `
		<div class="checkpoint %s">
			<div class="checkpoint-header">
				<div>
					<span style="color: #64748b; margin-right: 0.5rem;">#%d</span>
					<span class="checkpoint-name">%s</span>
					<span class="complexity">%s</span>
				</div>
				<span class="status-badge %s">%s</span>
			</div>
			<div class="description">%s</div>
`, status, number, escapeHTML(cp.Name), cp.Complexity, status, status, description)

	// Tasks
	if len(cp.Tasks) > 0 {
		sb.WriteString(`<div class="tasks"><h3>Tasks</h3>`)
		for _, task := range cp.Tasks {
			_, _ = fmt.Fprintf(&sb, `<div class="task">✓ %s</div>`, escapeHTML(task))
		}
		sb.WriteString(`</div>`)
	}

	// Gates
	if len(cp.Gates) > 0 {
		sb.WriteString(`<div class="gates"><h3>Gates</h3>`)
		for _, gate := range cp.Gates {
			gateStatus := "pending"
			feedback := ""
			if result, ok := cp.GateResults[gate]; ok {
				gateStatus = strings.ToLower(string(result.Status))
				feedback = result.Feedback
			}
			_, _ = fmt.Fprintf(&sb, `
				<div class="gate">
					<span>%s</span>
					<span class="gate-status %s">%s</span>
				</div>`, escapeHTML(gate), gateStatus, gateStatus)
			if feedback != "" {
				_, _ = fmt.Fprintf(&sb, `<div class="feedback">%s</div>`, escapeHTML(feedback))
			}
		}
		sb.WriteString(`</div>`)
	}

	sb.WriteString(`</div>`)
	return sb.String()
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// formatTimestamp formats a timestamp
func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "Just now"
	}
	return t.Format(time.RFC1123)
}
