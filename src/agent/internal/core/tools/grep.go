package tools

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

func grepExecute(_ string, args map[string]any, signal <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	pattern := getStringArg(args, "pattern", "")
	path := getStringArg(args, "path", "")
	ignoreCase := getBoolArg(args, "ignoreCase", false)

	if !isPathAllowed(path) {
		return errorResult("path outside allowed directories"), nil
	}

	select {
	case <-signal:
		return cancelledResult("Search"), nil
	default:
	}

	file, err := os.Open(path)
	if err != nil {
		return errorResult("opening file: " + err.Error()), nil
	}
	defer func() { _ = file.Close() }()

	regexPattern := pattern
	if ignoreCase {
		regexPattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return errorResult("invalid regex pattern: " + err.Error()), nil
	}

	var results []string
	lineNum := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNum++
		if re.MatchString(scanner.Text()) {
			results = append(results, fmt.Sprintf("%d:%s", lineNum, scanner.Text()))
		}
	}

	if len(results) == 0 {
		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: "No matches found"}},
		}, nil
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: strings.Join(results, "\n")}},
	}, nil
}
