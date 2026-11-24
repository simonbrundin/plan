package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LogEntry är en rad i loggen.
type LogEntry struct {
	Date  string
	Op    string
	Title string
}

// String renderar en loggrad i det parsebara formatet:
// "## [2026-04-02] ingest | Rubrik".
func (e LogEntry) String() string {
	return fmt.Sprintf("## [%s] %s | %s", e.Date, e.Op, e.Title)
}

// AppendLog lägger till poster i loggen. Loggen är append-only.
// Formatet är medvetet grep-bart: `grep "^## \\[" log.md | tail -5`.
func AppendLog(dir string, entries []LogEntry, details []string) error {
	path := filepath.Join(dir, LogFile)

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("kunde inte läsa %s: %w", path, err)
	}

	var b strings.Builder
	b.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		b.WriteString("\n")
	}
	if len(existing) > 0 {
		b.WriteString("\n")
	}

	for i, entry := range entries {
		b.WriteString(entry.String())
		b.WriteString("\n")
		if i == 0 {
			for _, detail := range details {
				b.WriteString("- " + detail + "\n")
			}
		}
	}
	b.WriteString("\n")

	return writeFile(path, b.String())
}
