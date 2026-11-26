package tools

import (
	"fmt"
	"os"
	"strings"

	"plan/agent/internal/core/messages"
	"plan/agent/internal/core/provider"
)

func readMultipleExecute(_ string, args map[string]any, signal <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	pathsVal, ok := args["paths"].([]any)
	if !ok {
		return errorResult("paths must be an array"), nil
	}

	select {
	case <-signal:
		return cancelledResult("ReadMultiple"), nil
	default:
	}

	var results []string
	for _, p := range pathsVal {
		path, ok := p.(string)
		if !ok {
			results = append(results, "Error: invalid path format")
			continue
		}

		if !isPathAllowed(path) {
			results = append(results, fmt.Sprintf("=== %s ===\nError: path outside allowed directories", path))
			continue
		}

		fileContent, err := os.ReadFile(path)
		if err != nil {
			results = append(results, fmt.Sprintf("=== %s (error) ===\n%v", path, err))
			continue
		}

		content := string(fileContent)
		if len(content) > 50000 {
			content = content[:50000] + "\n... (truncated)"
		}
		results = append(results, fmt.Sprintf("=== %s ===\n%s", path, content))
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: strings.Join(results, "\n\n")}},
	}, nil
}
