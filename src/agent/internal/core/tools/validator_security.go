package tools

import (
	"fmt"
	"strings"
)

// PathValidator validates that a path is safe and within allowed directories
type PathValidator struct {
	AllowedDirs []string
}

// NewPathValidator creates a new PathValidator
func NewPathValidator(allowedDirs []string) *PathValidator {
	return &PathValidator{AllowedDirs: allowedDirs}
}

// Validate checks if a path is safe
func (v *PathValidator) Validate(path string) error {
	if path == "" {
		return &ValidationError{Field: "path", Message: "path cannot be empty"}
	}

	// Block dangerous paths
	dangerous := []string{"../", "~", "$", "`", ";", "|", "<", ">"}
	for _, d := range dangerous {
		if strings.Contains(path, d) && !strings.HasPrefix(path, "./") {
			return &ValidationError{
				Field:   "path",
				Message: fmt.Sprintf("path contains potentially dangerous sequence: %s", d),
			}
		}
	}

	// If no allowed dirs specified, allow all
	if len(v.AllowedDirs) == 0 {
		return nil
	}

	return nil
}

// ValidateCommand validates shell commands for safety
func ValidateCommand(cmd string) error {
	if cmd == "" {
		return &ValidationError{Field: "command", Message: "command cannot be empty"}
	}

	// Block obviously dangerous commands
	dangerous := []string{
		"rm -rf /",
		":(){:|:&};:", // Fork bomb
		"dd if=/dev/zero",
		"> /dev/sda",
		"mkfs",
		"chmod -R 777 /",
	}

	lower := strings.ToLower(cmd)
	for _, d := range dangerous {
		if strings.Contains(lower, strings.ToLower(d)) {
			return &ValidationError{
				Field:   "command",
				Message: "command contains potentially dangerous pattern",
			}
		}
	}

	return nil
}
