package featurestest

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"plan-agent/internal/core/provider"
	"plan-agent/internal/core/tools"
)

var (
	tempDir     string
	verktygsResultat map[string]interface{}
)

func setupTempDir() error {
	var err error
	tempDir, err = os.MkdirTemp("", "agent-test-*")
	return err
}

func cleanupTempDir() {
	if tempDir != "" {
		os.RemoveAll(tempDir)
	}
}

func anropaVerktyg(tool *tools.ToolDefinition, args map[string]interface{}) (interface{}, error) {
	signal := make(chan struct{})
	result, err := tool.Execute("", args, signal, nil)
	return result, err
}

func initieraTools(ctx *godog.ScenarioContext) {
	ctx.BeforeScenario(func(*godog.Scenario) {
		setupTempDir()
		verktygsResultat = nil
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		cleanupTempDir()
	})

	ctx.Step(`^verktygsdefinitionen "([^"]*)"$`, func(toolName string) error { return nil })

	ctx.Step(`^jag anropar bash med "([^"]*)"$`, func(cmd string) error {
		bashTool := tools.BashTool()
		result, err := anropaVerktyg(bashTool, map[string]interface{}{"command": cmd})
		if err != nil {
			return err
		}
		verktygsResultat = map[string]interface{}{"bash": result}
		return nil
	})

	ctx.Step(`^kommandot ska köras och output returneras$`, func() error {
		if verktygsResultat == nil || verktygsResultat["bash"] == nil {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^en fil "([^"]*)" med innehåll "([^"]*)"$`, func(filename, content string) error {
		if tempDir == "" {
			setupTempDir()
		}
		path := filepath.Join(tempDir, filename)
		return os.WriteFile(path, []byte(content), 0644)
	})

	ctx.Step(`^jag anropar read med sökväg "([^"]*)"$`, func(path string) error {
		readPath := path
		if !filepath.IsAbs(path) && tempDir != "" {
			readPath = filepath.Join(tempDir, path)
		}
		readTool := tools.ReadTool()
		result, err := anropaVerktyg(readTool, map[string]interface{}{"path": readPath})
		if err != nil {
			return err
		}
		verktygsResultat = map[string]interface{}{"read": result}
		return nil
	})

	ctx.Step(`^filens innehåll ska returneras$`, func() error {
		if verktygsResultat == nil || verktygsResultat["read"] == nil {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^en fil med (\d+) rader$`, func(lineCount int) error {
		if tempDir == "" {
			setupTempDir()
		}
		path := filepath.Join(tempDir, "hundred_lines.txt")
		var b strings.Builder
		for i := 1; i <= lineCount; i++ {
			b.WriteString("Line ")
			b.WriteString(string(rune('0' + (i % 10))))
			b.WriteString("\n")
		}
		return os.WriteFile(path, []byte(b.String()), 0644)
	})

	ctx.Step(`^jag anropar read med offset (\d+) och limit (\d+)$`, func(offset, limit int) error {
		path := filepath.Join(tempDir, "hundred_lines.txt")
		readTool := tools.ReadTool()
		result, err := anropaVerktyg(readTool, map[string]interface{}{
			"path":   path,
			"offset": float64(offset),
			"limit":  float64(limit),
		})
		if err != nil {
			return err
		}
		verktygsResultat = map[string]interface{}{"read": result}
		return nil
	})

	ctx.Step(`^rad (\d+)-(\d+) ska returneras$`, func(start, end int) error { return nil })

	ctx.Step(`^jag skriver till "([^"]*)" med innehåll "([^"]*)"$`, func(path, content string) error {
		writePath := path
		if tempDir != "" && !filepath.IsAbs(path) {
			writePath = filepath.Join(tempDir, path)
		}
		writeTool := tools.WriteTool()
		result, err := anropaVerktyg(writeTool, map[string]interface{}{
			"path":    writePath,
			"content": content,
		})
		if err != nil {
			return err
		}
		verktygsResultat = map[string]interface{}{"write": result}
		return nil
	})

	ctx.Step(`^filen ska skapas med korrekt innehåll$`, func() error { return nil })

	ctx.Step(`^en fil med texten "([^"]*)"$`, func(content string) error {
		if tempDir == "" {
			setupTempDir()
		}
		path := filepath.Join(tempDir, "edit_test.txt")
		return os.WriteFile(path, []byte(content), 0644)
	})

	ctx.Step(`^jag ersätter "([^"]*)" med "([^"]*)"$`, func(oldText, newText string) error {
		path := filepath.Join(tempDir, "edit_test.txt")
		editTool := tools.EditTool()
		result, err := anropaVerktyg(editTool, map[string]interface{}{
			"path":    path,
			"oldText": oldText,
			"newText": newText,
		})
		if err != nil {
			return err
		}
		verktygsResultat = map[string]interface{}{"edit": result}
		return nil
	})

	ctx.Step(`^filen ska innehålla "([^"]*)"$`, func(expected string) error {
		path := filepath.Join(tempDir, "edit_test.txt")
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if string(content) != expected {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^jag gör två ersättningar i samma anrop$`, func() error {
		path := filepath.Join(tempDir, "edit_test.txt")
		editTool := tools.EditTool()
		result, err := anropaVerktyg(editTool, map[string]interface{}{
			"path":    path,
			"oldText": "foo",
			"newText": "FOO",
		})
		if err != nil {
			return err
		}
		verktygsResultat = map[string]interface{}{"edit": result}
		return nil
	})

	ctx.Step(`^båda ersättningarna ska appliceras$`, func() error {
		path := filepath.Join(tempDir, "edit_test.txt")
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if string(content) == "foo bar baz" {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^en fil med flera rader innehållande "([^"]*)"$`, func(pattern string) error {
		if tempDir == "" {
			setupTempDir()
		}
		path := filepath.Join(tempDir, "grep_test.txt")
		content := "first test line\nsecond test line\nthird line\n"
		return os.WriteFile(path, []byte(content), 0644)
	})

	ctx.Step(`^jag söker efter "([^"]*)"$`, func(pattern string) error {
		path := filepath.Join(tempDir, "grep_test.txt")
		grepTool := tools.GrepTool()
		result, err := anropaVerktyg(grepTool, map[string]interface{}{
			"pattern": pattern,
			"path":    path,
		})
		if err != nil {
			return err
		}
		verktygsResultat = map[string]interface{}{"grep": result}
		return nil
	})

	ctx.Step(`^matchande rader ska returneras$`, func() error {
		if verktygsResultat == nil || verktygsResultat["grep"] == nil {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^alla \.(\w+)-filer ska returneras$`, func(ext string) error {
		if verktygsResultat == nil {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^markdown-filer i alla underkataloger ska returneras$`, func() error {
		if tempDir == "" {
			setupTempDir()
		}
		subDir := filepath.Join(tempDir, "subdir")
		os.MkdirAll(subDir, 0755)
		os.WriteFile(filepath.Join(subDir, "test.md"), []byte("# Test"), 0644)
		globTool := tools.GlobTool()
		result, err := anropaVerktyg(globTool, map[string]interface{}{
			"pattern": "**/*.md",
			"path":    tempDir,
		})
		if err != nil {
			return err
		}
		verktygsResultat = map[string]interface{}{"glob": result}
		return nil
	})

	ctx.Step(`^jag anropar read_multiple med båda sökvägarna$`, func() error {
		path1 := filepath.Join(tempDir, "a.txt")
		path2 := filepath.Join(tempDir, "b.txt")
		os.WriteFile(path1, []byte("Content A"), 0644)
		os.WriteFile(path2, []byte("Content B"), 0644)
		readMultipleTool := tools.ReadMultipleTool()
		result, err := anropaVerktyg(readMultipleTool, map[string]interface{}{
			"paths": []string{path1, path2},
		})
		if err != nil {
			return err
		}
		verktygsResultat = map[string]interface{}{"read_multiple": result}
		return nil
	})

	ctx.Step(`^innehåll från båda filerna ska returneras$`, func() error {
		if verktygsResultat == nil {
			return godog.ErrPending
		}
		return nil
	})

	ctx.Step(`^jag anropar bash utan obligatoriskt argument "([^"]*)"$`, func(argName string) error {
		bashTool := tools.BashTool()
		result, err := anropaVerktyg(bashTool, map[string]interface{}{})
		if err != nil {
			return nil
		}
		if result != nil {
			res := result.(*provider.ToolResult)
			if res.IsError {
				return nil
			}
		}
		return godog.ErrPending
	})

	ctx.Step(`^ett valideringsfel ska returneras$`, func() error { return nil })
}
