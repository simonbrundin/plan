package tools

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"plan/agent/internal/core/messages"
	"plan/agent/internal/core/provider"
)

// planGetter is a function that returns the current plan
// This avoids import cycles between tools and workflow packages
var (
	planGetter   func() interface{}
	planGetterMu sync.RWMutex
)

// RegisterPlanGetter registers a function to get the current plan
// This is called during workflow initialization
func RegisterPlanGetter(getter func() interface{}) {
	planGetterMu.Lock()
	defer planGetterMu.Unlock()
	planGetter = getter
}

// AskHumanApprovalTool creates a tool that asks the user for approval
// This is used when a checkpoint has a human gate that needs input
func AskHumanApprovalTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "ask_human_approval",
		Description: "Ask the human user for approval on a checkpoint. Use this when you've completed a checkpoint that requires human review. Returns the human's decision.",
		InputSchema: askHumanApprovalSchema(),
		Execute:     askHumanApprovalExecute,
	}
}

func askHumanApprovalSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"checkpoint_name": map[string]any{
				"type":        "string",
				"description": "Name of the checkpoint you're asking approval for",
			},
			"summary": map[string]any{
				"type":        "string",
				"description": "Brief summary of what was accomplished",
			},
			"feedback_request": map[string]any{
				"type":        "string",
				"description": "Specific question or area you want feedback on",
			},
		},
		"required": []any{"checkpoint_name", "summary"},
	}
}

func askHumanApprovalExecute(
	toolCallID string,
	args map[string]any,
	signal <-chan struct{},
	onUpdate func(partial *provider.ToolResult),
) (*provider.ToolResult, error) {

	checkpointName := getStringArg(args, "checkpoint_name", "unknown")
	summary := getStringArg(args, "summary", "")
	feedbackRequest := getStringArg(args, "feedback_request", "")

	// Get plan via registered getter to avoid import cycle
	plan := getPlan()
	if plan == nil {
		return errorResult("No active plan"), nil
	}

	current := getCurrentCheckpoint(plan)
	if current == nil {
		return errorResult("No current checkpoint"), nil
	}

	// Print separator
	sep := strings.Repeat("=", 60)
	fmt.Printf("\n%s\n", sep)
	fmt.Printf("🔔 HUMAN APPROVAL REQUEST\n")
	fmt.Printf("%s\n\n", sep)

	fmt.Printf("Checkpoint: %s\n", checkpointName)
	if summary != "" {
		fmt.Printf("Summary: %s\n", summary)
	}
	if feedbackRequest != "" {
		fmt.Printf("Questions: %s\n", feedbackRequest)
	}

	// Show tasks
	tasks := getCheckpointTasks(current)
	if len(tasks) > 0 {
		fmt.Printf("\nTasks completed:\n")
		for i, task := range tasks {
			fmt.Printf("  ✅ %d. %s\n", i+1, task)
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("-", 60))
	fmt.Printf("Options:\n")
	fmt.Printf("  [y/Y] - Approve and continue\n")
	fmt.Printf("  [n/N] - Request changes (provide feedback)\n")
	fmt.Printf("  [s/S] - Skip this checkpoint\n")
	fmt.Printf("> ")

	// Non-blocking check
	select {
	case <-signal:
		return errorResult("Approval request cancelled"), nil
	default:
	}

	// Read user input with timeout
	scanner := bufio.NewScanner(os.Stdin)
	inputCh := make(chan string, 1)
	go func() {
		if scanner.Scan() {
			inputCh <- scanner.Text()
		} else {
			inputCh <- ""
		}
		close(inputCh)
	}()

	// Wait for input with timeout
	var response string
	select {
	case <-signal:
		return errorResult("Approval request cancelled"), nil
	case resp, ok := <-inputCh:
		if !ok || resp == "" {
			return errorResult("No input received"), nil
		}
		response = strings.TrimSpace(strings.ToLower(resp))
	case <-time.After(30 * time.Minute):
		return errorResult("Approval request timed out after 30 minutes"), nil
	}

	response = strings.TrimSpace(response)

	// Process response
	switch response {
	case "y", "yes":
		// Approve
		if err := updateCheckpointStatus(current, true); err != nil {
			fmt.Printf("Warning: Failed to update checkpoint: %v\n", err)
		}
		return &provider.ToolResult{
			Content: []messages.ContentBlock{
				&messages.TextBlock{
					Text: fmt.Sprintf("✅ Human approved checkpoint '%s'. You may now proceed to the next checkpoint.", checkpointName),
				},
			},
		}, nil

	case "n", "no":
		// Request changes - need feedback
		fmt.Printf("\nPlease provide your feedback:\n")
		fmt.Printf("> ")
		scanner2 := bufio.NewScanner(os.Stdin)
		var feedback strings.Builder
		for scanner2.Scan() {
			line := scanner2.Text()
			if line == "" {
				break
			}
			feedback.WriteString(line)
			feedback.WriteString("\n")
		}
		feedbackStr := strings.TrimSpace(feedback.String())

		if err := updateCheckpointStatus(current, false); err != nil {
			fmt.Printf("Warning: Failed to update checkpoint: %v\n", err)
		}

		return &provider.ToolResult{
			Content: []messages.ContentBlock{
				&messages.TextBlock{
					Text: fmt.Sprintf("⚠️ Human requested changes for checkpoint '%s'.\nFeedback: %s\n\nPlease address the feedback and try again.", checkpointName, feedbackStr),
				},
			},
		}, nil

	case "s", "skip":
		// Skip this checkpoint
		fmt.Printf("Skipping checkpoint...\n")
		return &provider.ToolResult{
			Content: []messages.ContentBlock{
				&messages.TextBlock{
					Text: fmt.Sprintf("⏭️ Human chose to skip checkpoint '%s'. Moving to next checkpoint.", checkpointName),
				},
			},
		}, nil

	default:
		return errorResult(fmt.Sprintf("Invalid input '%s'. Please enter y, n, or s.", response)), nil
	}
}

// getPlan returns the current plan via registered getter
func getPlan() interface{} {
	planGetterMu.RLock()
	defer planGetterMu.RUnlock()
	if planGetter == nil {
		return nil
	}
	return planGetter()
}

// getCurrentCheckpoint gets the current checkpoint from a plan
func getCurrentCheckpoint(plan interface{}) interface{} {
	type withGetCurrent interface {
		GetCurrent() interface{}
	}
	if p, ok := plan.(withGetCurrent); ok {
		return p.GetCurrent()
	}
	return nil
}

// getCheckpointTasks gets the tasks from a checkpoint
func getCheckpointTasks(cp interface{}) []string {
	type withTasks interface {
		GetTasks() []string
	}
	if c, ok := cp.(withTasks); ok {
		return c.GetTasks()
	}
	return nil
}

// updateCheckpointStatus updates the checkpoint status
func updateCheckpointStatus(cp interface{}, approved bool) error {
	type approvable interface {
		MarkPassed()
		MarkFailed()
	}
	if c, ok := cp.(approvable); ok {
		if approved {
			c.MarkPassed()
		} else {
			c.MarkFailed()
		}
		return nil
	}
	return fmt.Errorf("checkpoint does not support approval")
}
