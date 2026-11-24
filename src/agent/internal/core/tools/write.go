package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

func writeExecute(_ string, args map[string]any, signal <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	path := getStringArg(args, "path", "")
	content := getStringArg(args, "content", "")

	select {
	case <-signal:
		return cancelledResult("Write"), nil
	default:
	}

	// Validate path for safety
	if result := ValidatePathStrict(path); !result.Valid {
		return errorResult("invalid path: " + result.Message), nil
	}

	// Validate content for safety
	validator := DefaultContentValidator()
	if result := validator.ValidateContent([]byte(content), path); !result.Valid {
		return errorResult("invalid content: " + result.Message), nil
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errorResult("creating directory: " + err.Error()), nil
		}
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return errorResult("writing file: " + err.Error()), nil
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path)}},
	}, nil
}
