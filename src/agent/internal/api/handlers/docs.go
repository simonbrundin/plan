package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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

	// Lokal utveckling: src/agent/internal/api/handlers/docs.go -> repo-nära kataloger.
	_, currentFile, _, _ := runtime.Caller(0)
	handlersDir := filepath.Dir(currentFile)
	apiDir := filepath.Dir(handlersDir)
	agentDir := filepath.Dir(apiDir)
	dirs = append(dirs, filepath.Dir(agentDir))

	return dirs
}
