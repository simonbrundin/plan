package tools

import (
	"path/filepath"
	"strings"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

// AllowedPaths restricts file operations to these directories (empty = all allowed)
var AllowedPaths []string

func isPathAllowed(path string) bool {
	if len(AllowedPaths) == 0 {
		return true
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, allowed := range AllowedPaths {
		absAllowed, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, absAllowed) {
			return true
		}
	}
	return false
}

func applyOffsetLimit(content string, offset, limit int) string {
	lines := strings.Split(content, "\n")
	start := offset - 1
	if start < 0 {
		start = 0
	}
	if start >= len(lines) {
		return ""
	}
	lines = lines[start:]
	if limit > 0 && limit < len(lines) {
		lines = lines[:limit]
	}
	return strings.Join(lines, "\n")
}

func getStringArg(args map[string]any, name, fallback string) string {
	if v, ok := args[name].(string); ok {
		return v
	}
	return fallback
}

func getFloatArg(args map[string]any, name string, fallback float64) float64 {
	if v, ok := args[name].(float64); ok {
		return v
	}
	return fallback
}

func getBoolArg(args map[string]any, name string, fallback bool) bool {
	if v, ok := args[name].(bool); ok {
		return v
	}
	return fallback
}

func errorResult(msg string) *provider.ToolResult {
	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: "Error: " + msg}},
		IsError: true,
	}
}

func cancelledResult(op string) *provider.ToolResult {
	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: op + " cancelled"}},
		IsError: true,
	}
}
