package orchestrator

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"plan-agent/internal/core/checkpoint"
)

// CLIAdapter adapts the orchestrator for CLI usage
type CLIAdapter struct {
	orchestrator *Orchestrator
	reader       *bufio.Reader
}

// NewCLIAdapter creates a new CLI adapter
func NewCLIAdapter(o *Orchestrator) *CLIAdapter {
	return &CLIAdapter{
		orchestrator: o,
		reader:       bufio.NewReader(os.Stdin),
	}
}

// RunInteractive runs the orchestrator interactively
func (a *CLIAdapter) RunInteractive(ctx context.Context) error {
	fmt.Println("🚀 Checkpoint-Gate Workflow Orchestrator")
	fmt.Println("========================================")

	// Get feature from user
	fmt.Print("Feature to implement: ")
	feature, _ := a.reader.ReadString('\n')
	feature = strings.TrimSpace(feature)
	if feature == "" {
		return fmt.Errorf("feature cannot be empty")
	}

	fmt.Print("Description (optional): ")
	description, _ := a.reader.ReadString('\n')
	description = strings.TrimSpace(description)

	// Create plan
	fmt.Println("\n📋 Creating checkpoint plan...")
	plan, err := a.orchestrator.CreatePlan(ctx, feature, description)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// Show plan
	a.orchestrator.viewer.ShowPlan(plan)

	// Get approval
	fmt.Print("\nDo you approve this plan? (y/n): ")
	response, _ := a.reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Plan rejected. Exiting.")
		return nil
	}

	// Approve plan
	if err := a.orchestrator.ApprovePlan(plan.ID, "user"); err != nil {
		return fmt.Errorf("failed to approve plan: %w", err)
	}

	// Execute
	fmt.Println("🚀 Starting execution...")
	result, err := a.orchestrator.executePlan(ctx, plan)
	if err != nil {
		fmt.Printf("\n❌ Execution failed: %v\n", err)
		return err
	}

	// Show result
	a.showResult(result)

	return nil
}

// showResult displays the execution result
func (a *CLIAdapter) showResult(result *ExecutionResult) {
	sep := strings.Repeat("=", 60)
	fmt.Printf("\n%s\n", sep)
	fmt.Println("EXECUTION RESULT")
	fmt.Printf("%s\n", sep)

	switch result.Status {
	case StatusCompleted:
		fmt.Println("\n✅ All checkpoints completed successfully!")
	case StatusFailed:
		fmt.Println("\n❌ Execution failed")
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
	case StatusPendingApproval:
		fmt.Println("\n⏳ Plan awaiting approval")
	}

	fmt.Printf("\nProgress: %d/%d checkpoints\n",
		result.CurrentCheckpoint, result.TotalCheckpoints)

	if len(result.CheckpointResults) > 0 {
		fmt.Println("\nCheckpoint Results:")
		for i, cpResult := range result.CheckpointResults {
			status := "❌"
			if cpResult.Passed {
				status = "✅"
			}
			fmt.Printf("  %d. %s %s\n", i+1, status, cpResult.CheckpointID)
		}
	}
}

// SubmitHumanDecision allows external submission of human decisions
func (a *CLIAdapter) SubmitHumanDecision(checkpointID string, approved bool, feedback string) {
	// This would be called from an HTTP handler
	// For now, this is a placeholder for the human gate mechanism
	_ = checkpointID
	_ = approved
	_ = feedback
}

// WaitForHumanGate waits for human input at a checkpoint
func (a *CLIAdapter) WaitForHumanGate(cp *checkpoint.Checkpoint) bool {
	sep := strings.Repeat("=", 60)
	fmt.Printf("\n%s\n", sep)
	fmt.Printf("👤 HUMAN REVIEW REQUIRED: %s\n", cp.Name)
	fmt.Printf("%s\n\n", sep)

	fmt.Printf("Description: %s\n\n", cp.Description)

	fmt.Println("Tasks:")
	for i, task := range cp.Tasks {
		fmt.Printf("  %d. %s\n", i+1, task)
	}

	sepDash := strings.Repeat("-", 60)
	fmt.Printf("\n%s\n", sepDash)
	fmt.Println("Please manually test the feature.")
	fmt.Println("When ready, enter your decision:")
	fmt.Println("  [y] - Approve")
	fmt.Println("  [n] - Request changes")
	fmt.Println("  [f] - View failed gates")
	fmt.Print("> ")

	response, _ := a.reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	switch response {
	case "y", "yes":
		return true
	case "f":
		// Show failed gates
		for gateType, result := range cp.GateResults {
			if result.Status == checkpoint.GateStatusFailed {
				fmt.Printf("\n❌ %s failed:\n", gateType)
				fmt.Printf("   %s\n", result.Feedback)
				if len(result.Issues) > 0 {
					fmt.Println("   Issues:")
					for _, issue := range result.Issues {
						fmt.Printf("   - [%s] %s: %s\n", issue.Severity, issue.Title, issue.Description)
					}
				}
			}
		}
		fmt.Print("\nPress Enter when ready to decide...")
		// EOF är normalt - användaren kan avsluta med Ctrl+D.
		_, _ = a.reader.ReadString('\n')
		return false
	default:
		return false
	}
}
