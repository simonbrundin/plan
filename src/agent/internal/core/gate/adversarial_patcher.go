package gate

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Patch represents a unified diff patch
type Patch struct {
	File  string
	Hunks []Hunk
}

// Hunk represents a single hunk in a patch
type Hunk struct {
	OldStart, OldCount int
	NewStart, NewCount int
	Lines              []string
}

// FixApplicationResult contains the result of applying fixes
type FixApplicationResult struct {
	Applied  []string
	Failed   []string
	Verified bool
}

func applyFixerPatches(fixerResponse string) *FixApplicationResult {
	result := &FixApplicationResult{Applied: []string{}, Failed: []string{}}

	patches := parsePatches(fixerResponse)
	if len(patches) == 0 {
		result.Failed = append(result.Failed, "No patches found in fixer response")
		return result
	}

	for _, patch := range patches {
		if err := applyPatch(patch); err != nil {
			result.Failed = append(result.Failed, fmt.Sprintf("%s: %v", patch.File, err))
		} else {
			result.Applied = append(result.Applied, patch.File)
			if err := verifyFile(patch.File); err != nil {
				result.Failed = append(result.Failed, fmt.Sprintf("verification failed for %s: %v", patch.File, err))
			}
		}
	}

	result.Verified = len(result.Failed) == 0
	return result
}

func parsePatches(response string) []Patch {
	var patches []Patch
	lines := strings.Split(response, "\n")
	inCodeBlock := false
	var currentPatch Patch
	var currentHunk []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				// Skip the code block marker
				bl := trimmed[3:]
				if strings.HasPrefix(bl, "diff") || strings.HasPrefix(bl, "patch") {
					continue
				}
			} else {
				inCodeBlock = false
				if len(currentHunk) > 0 {
					if hunk, ok := parseHunk(currentHunk); ok {
						currentPatch.Hunks = append(currentPatch.Hunks, hunk)
					}
					currentHunk = nil
				}
				if len(currentPatch.Hunks) > 0 {
					patches = append(patches, currentPatch)
					currentPatch = Patch{}
				}
				continue
			}
		}

		if !inCodeBlock {
			continue
		}

		// Parse diff header
		if strings.HasPrefix(trimmed, "---") {
			if len(currentPatch.Hunks) > 0 {
				patches = append(patches, currentPatch)
				currentPatch = Patch{}
			}
			continue
		}
		if strings.HasPrefix(trimmed, "+++") || strings.HasPrefix(trimmed, "---") {
			parts := strings.Split(trimmed, "\t")
			if len(parts) >= 2 {
				file := stripPathPrefix(parts[1])
				currentPatch.File = file
			}
			continue
		}

		// Parse hunk header
		if strings.HasPrefix(trimmed, "@@") {
			if len(currentHunk) > 0 {
				if hunk, ok := parseHunk(currentHunk); ok {
					currentPatch.Hunks = append(currentPatch.Hunks, hunk)
				}
				currentHunk = nil
			}
			currentHunk = append(currentHunk, trimmed)
			continue
		}

		if len(currentHunk) > 0 {
			currentHunk = append(currentHunk, line)
		}
	}

	// Don't forget the last hunk
	if len(currentHunk) > 0 {
		if hunk, ok := parseHunk(currentHunk); ok {
			currentPatch.Hunks = append(currentPatch.Hunks, hunk)
		}
	}
	if len(currentPatch.Hunks) > 0 {
		patches = append(patches, currentPatch)
	}

	return patches
}

func stripPathPrefix(file string) string {
	file = strings.TrimPrefix(file, "a/")
	file = strings.TrimPrefix(file, "b/")
	return file
}

func parseHunk(lines []string) (Hunk, bool) {
	if len(lines) == 0 {
		return Hunk{}, false
	}

	var hunk Hunk
	header := lines[0]

	// Parse @@ -old +new @@ format
	n, err := fmt.Sscanf(header, "@@ -%d,%d +%d,%d @@", &hunk.OldStart, &hunk.OldCount, &hunk.NewStart, &hunk.NewCount)
	if err != nil {
		_, err = fmt.Sscanf(header, "@@ -%d +%d @@", &hunk.OldStart, &hunk.NewStart)
		hunk.OldCount = 1
		hunk.NewCount = 1
	}
	if err != nil || n == 0 {
		return Hunk{}, false
	}

	hunk.Lines = lines[1:]
	return hunk, true
}

func applyPatch(patch Patch) error {
	if patch.File == "" {
		return fmt.Errorf("no file specified in patch")
	}

	original, err := os.ReadFile(patch.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	lines := strings.Split(string(original), "\n")

	for _, hunk := range patch.Hunks {
		lines, err = applyHunkToLines(lines, hunk)
		if err != nil {
			return fmt.Errorf("failed to apply hunk: %w", err)
		}
	}

	modified := strings.Join(lines, "\n")
	if !strings.HasSuffix(modified, "\n") && strings.HasSuffix(string(original), "\n") {
		modified += "\n"
	}

	if err := os.WriteFile(patch.File, []byte(modified), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func applyHunkToLines(lines []string, hunk Hunk) ([]string, error) {
	start := hunk.OldStart - 1
	if start < 0 {
		start = 0
	}

	var newLines []string
	oldIdx := 0

	// Copy lines before the hunk
	for oldIdx < start && oldIdx < len(lines) {
		newLines = append(newLines, lines[oldIdx])
		oldIdx++
	}

	// Process hunk lines
	for _, line := range hunk.Lines {
		if len(line) == 0 {
			newLines = append(newLines, "")
			oldIdx++
			continue
		}

		prefix := line[0]
		content := line[1:]

		switch prefix {
		case ' ', '+':
			newLines = append(newLines, content)
			oldIdx++
		case '-':
			oldIdx++
		default:
			newLines = append(newLines, line)
			oldIdx++
		}
	}

	// Add remaining lines
	for oldIdx < len(lines) {
		newLines = append(newLines, lines[oldIdx])
		oldIdx++
	}

	return newLines, nil
}

func verifyFile(path string) error {
	if !strings.HasSuffix(path, ".go") {
		return nil
	}

	if err := runCommand("gofmt", "-l", path); err != nil {
		return fmt.Errorf("gofmt failed: %w", err)
	}

	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("%s: %s", err, string(output))
		}
		return err
	}
	return nil
}
