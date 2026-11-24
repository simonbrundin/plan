package tools

import (
	"fmt"
	"os"
	"strings"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

// ToolFunc is the function signature for tool implementations
type ToolFunc func(
	toolCallID string,
	args map[string]any,
	signal <-chan struct{},
	onUpdate func(partial *provider.ToolResult),
) (*provider.ToolResult, error)

// ToolDefinition defines a tool available to the agent
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema map[string]any
	Execute     ToolFunc
}

// ValidateArguments validates tool arguments against the schema.
// Uses full JSON Schema validation including type, required, minLength,
// maxLength, minimum, maximum, pattern, enum, and array items.
func ValidateArguments(args map[string]any, schema map[string]any) error {
	return Validate(args, schema)
}

// ToProviderFormat converts to provider.ToolDefinition format
func (t *ToolDefinition) ToProviderFormat() provider.ToolDefinition {
	return provider.ToolDefinition{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: t.InputSchema,
	}
}

// ============================================================================
// Tool Definitions
// ============================================================================

// BashTool creates a bash/shell execution tool
func BashTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "bash",
		Description: "Execute a bash command. Returns stdout and stderr.",
		InputSchema: bashInputSchema(),
		Execute:     bashExecute,
	}
}

// ReadTool creates a file reading tool
func ReadTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "read",
		Description: "Read the contents of a file. Supports text files.",
		InputSchema: fileInputSchema(),
		Execute:     readExecute,
	}
}

// WriteTool creates a file writing tool
func WriteTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "write",
		Description: "Write content to a file. Creates the file if it doesn't exist, overwrites if it does.",
		InputSchema: writeInputSchema(),
		Execute:     writeExecute,
	}
}

// EditTool creates a file editing tool
func EditTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "edit",
		Description: "Edit specific sections of a file using exact text replacement.",
		InputSchema: editInputSchema(),
		Execute:     editExecute,
	}
}

// GlobTool creates a file globbing tool
func GlobTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "glob",
		Description: "Find files matching a pattern using glob syntax.",
		InputSchema: globInputSchema(),
		Execute:     globExecute,
	}
}

// GrepTool creates a text search tool
func GrepTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "grep",
		Description: "Search for lines matching a pattern in a file.",
		InputSchema: grepInputSchema(),
		Execute:     grepExecute,
	}
}

// ReadMultipleTool creates a tool for reading multiple files at once
func ReadMultipleTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "read_multiple",
		Description: "Read multiple files in a single call.",
		InputSchema: readMultipleInputSchema(),
		Execute:     readMultipleExecute,
	}
}

// ============================================================================
// Tool Implementations

func editExecute(_ string, args map[string]any, signal <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	path := getStringArg(args, "path", "")
	oldText := getStringArg(args, "oldText", "")
	newText := getStringArg(args, "newText", "")

	select {
	case <-signal:
		return cancelledResult("Edit"), nil
	default:
	}

	if !isPathAllowed(path) {
		return errorResult("path outside allowed directories"), nil
	}

	fileContent, err := os.ReadFile(path)
	if err != nil {
		return errorResult(fmt.Sprintf("reading file: %v", err)), nil
	}

	content := string(fileContent)
	if !strings.Contains(content, oldText) {
		return errorResult(fmt.Sprintf("oldText not found in file\nSearched for: %q", oldText)), nil
	}

	newContent := strings.Replace(content, oldText, newText, 1)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return errorResult(fmt.Sprintf("writing file: %v", err)), nil
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: fmt.Sprintf("Successfully edited %s", path)}},
	}, nil
}

// ============================================================================
// Default Tool Set
// ============================================================================

// GetDefaultTools returns the default set of tools available to the agent
func GetDefaultTools() []*ToolDefinition {
	return []*ToolDefinition{
		BashTool(),
		ReadTool(),
		WriteTool(),
		EditTool(),
		GlobTool(),
		GrepTool(),
		ReadMultipleTool(),
	}
}

// DefaultSystemPrompt returns the default system prompt for the agent
func DefaultSystemPrompt() string {
	return `You are a helpful AI coding assistant.

You have access to the following tools:
- bash: Execute bash commands
- read: Read file contents
- write: Write content to files
- edit: Edit specific parts of files
- glob: Find files matching patterns
- grep: Search for text in files
- read_multiple: Read multiple files at once

When using tools:
- Always validate arguments before execution
- Use appropriate timeouts
- Handle errors gracefully
- Provide clear feedback

For file operations, prefer edit over write when making small changes.`
}
