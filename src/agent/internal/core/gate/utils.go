package gate

import "strings"

// String helpers used across gate implementations

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

func startsWith(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func toLower(s string) string {
	return strings.ToLower(s)
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

// contains is a simple substring search (empty substr matches everything)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	return strings.Contains(s, substr)
}
