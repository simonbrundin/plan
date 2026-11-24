package tools

import (
	"os"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

func readExecute(_ string, args map[string]any, signal <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	path := getStringArg(args, "path", "")
	if path == "" {
		return errorResult("path is required"), nil
	}

	select {
	case <-signal:
		return cancelledResult("Read"), nil
	default:
	}

	// Validate path for safety
	if result := ValidatePathStrict(path); !result.Valid {
		return errorResult("invalid path: " + result.Message), nil
	}

	fileContent, err := os.ReadFile(path)
	if err != nil {
		return errorResult("reading file: " + err.Error()), nil
	}

	content := string(fileContent)
	if offset := getFloatArg(args, "offset", 0); offset > 0 {
		limit := getFloatArg(args, "limit", 0)
		content = applyOffsetLimit(content, int(offset), int(limit))
	} else if limit := getFloatArg(args, "limit", 0); limit > 0 {
		content = applyOffsetLimit(content, 1, int(limit))
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: content}},
	}, nil
}
