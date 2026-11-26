package commands

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Command represents a loaded command from src/agent/commands/
type Command struct {
	Name         string   // filename without extension
	Description  string   // from frontmatter description or body
	AllowedTools []string // from frontmatter allowed-tools
	Workflow     string   // the full command file content
}

// Registry holds all available commands
type Registry struct {
	commands map[string]*Command
}

// DefaultCommandsPath returns the default path to the commands directory.
//
// Kommandona ligger i src/agent/commands. Sökvägen hittas genom att gå upp
// från binären eller arbetskatalogen tills en kandidat matchar - agenten kan
// köras både från repot (go run) och från imagen (/app).
func DefaultCommandsPath() string {
	// Kandidater i prioritetsordning. Den första är den aktuella placeringen,
	// den andra håller gamla installationer (kommandon under src/commands)
	// fungerande.
	candidates := [][]string{
		{"src", "agent", "commands"},
		{"src", "commands"},
	}

	// Try relative to executable
	if exe, err := os.Executable(); err == nil {
		// bin/agent -> agent/ -> repo/
		base := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
		for _, parts := range candidates {
			cmdDir := filepath.Join(append([]string{base}, parts...)...)
			if _, err := os.Stat(cmdDir); err == nil {
				return cmdDir
			}
		}
	}

	// Fallback: try relative to working directory
	wd, err := os.Getwd()
	if err == nil {
		// wd/src/agent/internal/core/commands -> repo/src/agent/commands
		for i := 0; i < 6; i++ {
			for _, parts := range candidates {
				cmdDir := filepath.Join(append([]string{wd}, parts...)...)
				if _, err := os.Stat(cmdDir); err == nil {
					return cmdDir
				}
			}
			parent := filepath.Dir(wd)
			if parent == wd {
				break
			}
			wd = parent
		}
	}

	return "" // no commands found
}

var frontmatterRE = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n`)
var descriptionRE = regexp.MustCompile(`(?m)^description:\s*(.*)$`)
var allowedToolsRE = regexp.MustCompile(`(?m)^allowed-tools:\s*(.*)$`)

func parseFrontmatter(content string) (description string, allowedTools []string) {
	matches := frontmatterRE.FindStringSubmatch(content)
	if len(matches) < 2 {
		return "", nil
	}
	frontmatter := matches[1]

	descMatches := descriptionRE.FindStringSubmatch(frontmatter)
	if len(descMatches) > 1 {
		desc := strings.TrimSpace(descMatches[1])
		// Handle YAML multiline: description: >-
		// Content continues on next indented line
		if strings.HasPrefix(desc, ">-") || strings.HasPrefix(desc, ">") || strings.HasPrefix(desc, "|") {
			lines := strings.Split(frontmatter, "\n")
			foundEnd := false
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				if foundEnd {
					desc = strings.TrimSpace(line)
					break
				}
				if strings.Contains(line, ">-") || strings.Contains(line, "|") || strings.Contains(line, ">") {
					foundEnd = true
				}
			}
		} else {
			desc = strings.Trim(desc, "<>\n ")
		}
		description = desc
	}

	toolMatches := allowedToolsRE.FindStringSubmatch(frontmatter)
	if len(toolMatches) > 1 {
		tools := strings.TrimSpace(toolMatches[1])
		if tools != "" {
			parts := strings.Split(tools, ",")
			for _, p := range parts {
				t := strings.TrimSpace(p)
				if t != "" {
					allowedTools = append(allowedTools, t)
				}
			}
		}
	}

	return description, allowedTools
}

// Load reads all .md files from the given directory
func Load(dir string) (*Registry, error) {
	if dir == "" {
		return &Registry{commands: map[string]*Command{}}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{commands: map[string]*Command{}}, nil
		}
		return nil, err
	}

	registry := &Registry{commands: map[string]*Command{}}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		path := filepath.Join(dir, entry.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)

		description, allowedTools := parseFrontmatter(content)
		if description == "" {
			description = strings.SplitN(content, "\n", 2)[0]
			if strings.HasPrefix(description, "# ") {
				description = strings.TrimPrefix(description, "# ")
				description = strings.TrimSpace(description)
			}
		}

		registry.commands[name] = &Command{
			Name:         name,
			Description:  description,
			AllowedTools: allowedTools,
			Workflow:     content,
		}
	}

	return registry, nil
}

// Get returns a command by name
func (r *Registry) Get(name string) *Command {
	return r.commands[name]
}

// List returns all available commands
func (r *Registry) List() []*Command {
	result := make([]*Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		result = append(result, cmd)
	}
	return result
}

// Names returns all command names
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	return names
}
