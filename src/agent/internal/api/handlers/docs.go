package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetReadme returns the README.md content
// @Summary Get README
// @Description Returns the README.md content from the project root
// @Tags docs
// @Produce text/plain
// @Success 200 {string} string
// @Router /api/v1/readme [get]
func GetReadme(c *gin.Context) {
	content, err := getProjectFile("README.md")
	if err != nil {
		c.String(http.StatusNotFound, "README.md not found")
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", content)
}

// GetRoadmap returns the roadmap as markdown
// @Summary Get Roadmap
// @Description Returns the roadmap generated from goals.json (goalsen är roadmaps hem)
// @Tags docs
// @Produce text/plain
// @Success 200 {string} string
// @Router /api/v1/roadmap [get]
func GetRoadmap(c *gin.Context) {
	goals, err := loadGoalsFromFile()
	if err != nil {
		c.String(http.StatusInternalServerError, "Kunde inte läsa goals.json: %v", err)
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(renderRoadmap(goals)))
}

// renderRoadmap renders goals.json som markdown: en statusöversikt och en
// sektion per status. ROADMAP.md togs bort när målen flyttades till goals.json,
// så roadmaps hem är nu målen.
func renderRoadmap(goals *GoalsResponse) string {
	var b strings.Builder

	b.WriteString("# Roadmap — Agents\n\n")
	b.WriteString("> Genererad från `goals.json` — målen är roadmaps hem.\n\n")

	counts := map[string]int{}
	for _, g := range goals.Goals {
		counts[g.Status]++
	}

	b.WriteString("## Statusöversikt\n\n")
	b.WriteString("| Status | Antal |\n| --- | --- |\n")
	for _, s := range goals.Statuses {
		_, _ = fmt.Fprintf(&b, "| %s | %d |\n", s.Label, counts[s.Name])
	}
	b.WriteString("\n")

	for _, s := range goals.Statuses {
		section := make([]Goal, 0, len(goals.Goals))
		for _, g := range goals.Goals {
			if g.Status == s.Name {
				section = append(section, g)
			}
		}
		if len(section) == 0 {
			continue
		}

		// Högst prioritet först, sedan titeln som sorteringsnyckel.
		sort.SliceStable(section, func(i, j int) bool {
			if section[i].Priority != section[j].Priority {
				return section[i].Priority > section[j].Priority
			}
			return section[i].Title < section[j].Title
		})

		_, _ = fmt.Fprintf(&b, "## %s (%d)\n\n", s.Label, len(section))
		for _, g := range section {
			title := g.Title
			if g.Icon != "" {
				title = g.Icon + " " + title
			}
			if desc := firstLine(g.Description); desc != "" {
				title += " — " + desc
			}
			_, _ = fmt.Fprintf(&b, "- %s\n", title)
		}
		b.WriteString("\n")
	}

	return b.String()
}

// firstLine returns the first non-empty line of a text, trimmed.
func firstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// getProjectFile reads a file from the project root (agents repo).
//
// Sökvägen löses i den ordning:
//  1. AGENT_DOCS_DIR - explicit katalog med filerna
//  2. PROJECT_ROOT - repots rot (samma env som goals-handlern använder)
//  3. working directory - /app i containern
//  4. repo-roten härifrån (runtime.Caller) - lokal utveckling
func getProjectFile(filename string) ([]byte, error) {
	var searched []string

	for _, dir := range projectFileDirs() {
		path := filepath.Join(dir, filename)
		searched = append(searched, path)
		if content, err := os.ReadFile(path); err == nil {
			return content, nil
		}
	}

	return nil, fmt.Errorf("%s hittades inte (letade i %s)", filename, strings.Join(searched, ", "))
}

// projectFileDirs returns the directories to search for project files.
func projectFileDirs() []string {
	var dirs []string

	if dir := os.Getenv("AGENT_DOCS_DIR"); dir != "" {
		dirs = append(dirs, dir)
	}
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		dirs = append(dirs, root)
	}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}

	// Lokal utveckling: src/go/internal/api/handlers/docs.go -> repo-roten.
	_, currentFile, _, _ := runtime.Caller(0)
	handlersDir := filepath.Dir(currentFile)
	apiDir := filepath.Dir(handlersDir)
	agentDir := filepath.Dir(apiDir)
	dirs = append(dirs, filepath.Dir(agentDir))

	return dirs
}
