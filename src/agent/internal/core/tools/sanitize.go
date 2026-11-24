package tools

import (
	"path/filepath"
	"regexp"
	"strings"
)

// PathAllowlist defines allowed paths for file operations
var PathAllowlist = []string{
	"/home/simon",
	"/tmp",
	"/var",
}

// PathDenylist defines paths that are always forbidden
var PathDenylist = []string{
	"/etc/shadow",
	"/etc/passwd",
	"/etc/sudoers",
	"/root/.ssh",
	"/home/*/.ssh",
	"/.git/objects",
	"/.git/config",
}

// CommandBlocklist patterns that are always forbidden
var CommandBlocklist = []string{
	`rm\s+-rf\s+/`,
	`dd\s+if=`,
	`:(){ :|:& };:`,
	`fork\(\)`,
	`/proc/self`,
	`curl.*-X.*--data`,
	`wget.*--post-data`,
	`nc\s+-e`,
	`bash\s+-i`,
	`python.*-c\s+import\s+os`,
	`perl.*-e.*system`,
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Valid   bool
	Message string
}

// ValidatePathStrict checks if a path is allowed for file operations
func ValidatePathStrict(path string) ValidationResult {
	// Check for null bytes (null byte injection)
	if strings.Contains(path, "\x00") {
		return ValidationResult{Valid: false, Message: "path contains null byte"}
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// Check for traversal attempts
	if strings.Contains(path, "..") {
		return ValidationResult{Valid: false, Message: "path traversal not allowed"}
	}

	// Check for suspicious patterns (always forbidden)
	suspicious := []string{
		"/.git/hooks",
		"/.github/workflows",
		"/__pycache__",
		"/node_modules/.bin",
	}
	for _, s := range suspicious {
		if strings.Contains(cleaned, s) {
			return ValidationResult{Valid: false, Message: "suspicious path pattern: " + s}
		}
	}

	// Check denylist (always forbidden)
	for _, denied := range PathDenylist {
		if strings.HasPrefix(cleaned, denied) || cleaned == denied {
			return ValidationResult{Valid: false, Message: "path is in denylist"}
		}
	}

	// Check allowlist (if not empty)
	if len(PathAllowlist) > 0 {
		allowed := false
		for _, allowedPath := range PathAllowlist {
			if strings.HasPrefix(cleaned, allowedPath) {
				allowed = true
				break
			}
		}
		if !allowed {
			return ValidationResult{Valid: false, Message: "path not in allowlist"}
		}
	}

	return ValidationResult{Valid: true}
}

// ValidateCommandStrict checks if a command is safe to execute
func ValidateCommandStrict(command string) ValidationResult {
	// Check for null bytes
	if strings.Contains(command, "\x00") {
		return ValidationResult{Valid: false, Message: "command contains null byte"}
	}

	// Check blocklist patterns
	for _, blocked := range CommandBlocklist {
		re := regexp.MustCompile(blocked)
		if re.MatchString(command) {
			return ValidationResult{Valid: false, Message: "command matches blocklist pattern"}
		}
	}

	// Check for dangerous redirections
	dangerousRedirects := []string{
		">/etc/passwd",
		">/etc/shadow",
		">~/.ssh",
		">/dev/sda",
		">/dev/null",
		"2>&1", // Could be used to hide errors
	}
	for _, d := range dangerousRedirects {
		if strings.Contains(command, d) {
			return ValidationResult{Valid: false, Message: "dangerous redirection: " + d}
		}
	}

	// Check for pipe to shell
	if matched, _ := regexp.MatchString(`\|\s*sh\b`, command); matched {
		return ValidationResult{Valid: false, Message: "pipe to shell not allowed"}
	}

	// Check for command substitution
	if strings.Contains(command, "$(") || strings.Contains(command, "`") {
		return ValidationResult{Valid: false, Message: "command substitution not allowed"}
	}

	// Check for environment variable expansion
	if matched, _ := regexp.MatchString(`\$\{?[A-Za-z_][A-Za-z0-9_]*\}?`, command); matched {
		return ValidationResult{Valid: false, Message: "environment variable expansion not allowed"}
	}

	return ValidationResult{Valid: true}
}

// SanitizePathStrict sanitizes a path for safe use
func SanitizePathStrict(path string) string {
	// Remove null bytes
	path = strings.ReplaceAll(path, "\x00", "")

	// Clean the path
	path = filepath.Clean(path)

	// Remove leading/trailing whitespace
	path = strings.TrimSpace(path)

	return path
}

// SanitizeCommandStrict sanitizes a command for safe use
func SanitizeCommandStrict(command string) string {
	// Remove null bytes
	command = strings.ReplaceAll(command, "\x00", "")

	// Remove backticks
	command = strings.ReplaceAll(command, "`", "")

	// Remove command substitution
	command = regexp.MustCompile(`\$\([^)]*\)`).ReplaceAllString(command, "")

	return command
}

// ContentValidator validates file content
type ContentValidator struct {
	MaxSize       int64 // Maximum file size in bytes
	BlockedExts   []string
	BlockedPrefix []string
}

// DefaultContentValidator returns a default content validator
func DefaultContentValidator() *ContentValidator {
	return &ContentValidator{
		MaxSize:     10 * 1024 * 1024, // 10MB
		BlockedExts: []string{".exe", ".dll", ".so", ".dylib"},
	}
}

// ValidateContent checks if content is safe to write
func (v *ContentValidator) ValidateContent(content []byte, filename string) ValidationResult {
	// Check size
	if int64(len(content)) > v.MaxSize {
		return ValidationResult{Valid: false, Message: "content exceeds maximum size"}
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(filename))
	for _, blocked := range v.BlockedExts {
		if ext == blocked {
			return ValidationResult{Valid: false, Message: "file extension not allowed: " + ext}
		}
	}

	// Check for shebang manipulation
	if len(content) > 2 && content[0] == '#' && content[1] == '!' {
		return ValidationResult{Valid: false, Message: "shebang not allowed"}
	}

	// Check for null bytes in content
	if strings.Contains(string(content), "\x00") {
		return ValidationResult{Valid: false, Message: "content contains null bytes"}
	}

	return ValidationResult{Valid: true}
}

// SetAllowlist sets the path allowlist
func SetAllowlist(paths []string) {
	PathAllowlist = paths
}

// AddToAllowlist adds a path to the allowlist
func AddToAllowlist(path string) {
	PathAllowlist = append(PathAllowlist, path)
}
