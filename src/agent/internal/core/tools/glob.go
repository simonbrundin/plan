package tools

import (
	"path/filepath"
	"strings"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

func globExecute(_ string, args map[string]any, signal <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	pattern := getStringArg(args, "pattern", "")
	baseDir := getStringArg(args, "path", ".")

	if !isPathAllowed(baseDir) {
		return errorResult("path outside allowed directories"), nil
	}

	searchPattern := filepath.Join(baseDir, pattern)
	matches, err := filepath.Glob(searchPattern)
	if err != nil {
		return errorResult("invalid glob pattern: " + err.Error()), nil
	}

	var result []string
	for _, match := range matches {
		if isPathAllowed(match) {
			result = append(result, match)
		}
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: strings.Join(result, "\n")}},
	}, nil
}
