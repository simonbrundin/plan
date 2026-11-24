package commands

import (
	"fmt"
	"strings"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
	"plan-agent/internal/core/tools"
)

// ListCommandsTool returns the list_commands tool definition
func ListCommandsTool() *tools.ToolDefinition {
	return &tools.ToolDefinition{
		Name:        "list_commands",
		Description: "List all available commands. Returns command names and descriptions.",
		InputSchema: nil,
		Execute:     listCommandsExecute,
	}
}

// RunCommandTool returns the run_command tool definition
func RunCommandTool() *tools.ToolDefinition {
	return &tools.ToolDefinition{
		Name:        "run_command",
		Description: "Load and execute a named command. Returns the command's workflow and instructions.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "The name of the command to run (e.g. 'push', 'debug-flux', 'ai-prime')",
				},
			},
			"required": []any{"name"},
		},
		Execute: runCommandExecute,
	}
}

func listCommandsExecute(
	_ string,
	_ map[string]any,
	_ <-chan struct{},
	_ func(partial *provider.ToolResult),
) (*provider.ToolResult, error) {
	// Load commands if not already loaded
	registry, err := Load(DefaultCommandsPath())
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load commands: %v", err)), nil
	}

	cmds := registry.List()
	if len(cmds) == 0 {
		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: "No commands available."}},
		}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Available commands (%d):\n\n", len(cmds)))
	for _, cmd := range cmds {
		sb.WriteString(fmt.Sprintf("• **%s**: %s\n", cmd.Name, cmd.Description))
	}
	sb.WriteString("\nUse run_command with the command name to execute one.")

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: sb.String()}},
	}, nil
}

func runCommandExecute(
	_ string,
	args map[string]any,
	_ <-chan struct{},
	_ func(partial *provider.ToolResult),
) (*provider.ToolResult, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return errorResult("command name is required"), nil
	}

	registry, err := Load(DefaultCommandsPath())
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load commands: %v", err)), nil
	}

	cmd := registry.Get(name)
	if cmd == nil {
		return errorResult(fmt.Sprintf("command '%s' not found. Use list_commands to see available commands.", name)), nil
	}

	// Build response: describe what the command does and the workflow
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Command: %s\n\n", cmd.Name))
	sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", cmd.Description))

	if len(cmd.AllowedTools) > 0 {
		sb.WriteString(fmt.Sprintf("**Allowed tools**: %s\n\n", strings.Join(cmd.AllowedTools, ", ")))
	}

	sb.WriteString("## Workflow\n\n")
	sb.WriteString("Follow the steps below to execute this command:\n\n")

	// Extract workflow section from the command file
	lines := strings.Split(cmd.Workflow, "\n")
	inWorkflow := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## Workflow") || strings.HasPrefix(trimmed, "### ") {
			inWorkflow = true
			if strings.HasPrefix(trimmed, "## Workflow") {
				continue
			}
			// Count step number
			step := 0
			for j := i - 1; j >= 0; j-- {
				if strings.HasPrefix(strings.TrimSpace(lines[j]), "### ") {
					step++
				}
			}
			if step > 0 {
				sb.WriteString(fmt.Sprintf("\n**Step %d**:\n", step))
			}
			// Strip the ### prefix
			stepText := strings.TrimPrefix(trimmed, "### ")
			stepText = strings.TrimPrefix(stepText, "## ")
			sb.WriteString(stepText + "\n")
			continue
		}

		if inWorkflow {
			// Stop at next ## heading (not ###)
			if strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "### ") {
				break
			}
			sb.WriteString(trimmed + "\n")
		}
	}

	sb.WriteString("\n---\nExecute the steps above using your available tools. Use bash to run shell commands, read to inspect files, and write/edit to make changes.")

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: sb.String()}},
	}, nil
}

func errorResult(msg string) *provider.ToolResult {
	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: msg}},
		IsError: true,
	}
}
