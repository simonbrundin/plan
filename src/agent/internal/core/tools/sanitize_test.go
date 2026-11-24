package tools

import (
	"testing"
)

func TestValidatePathStrict(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantOK  bool
		wantMsg string
	}{
		// Valid paths
		{"valid home dir", "/home/simon/projects", true, ""},
		{"valid tmp dir", "/tmp/test", true, ""},
		{"valid var dir", "/var/log/test", true, ""},

		// Invalid paths
		{"null byte injection", "/home/simon\x00/etc/passwd", false, "null byte"},
		{"path traversal", "../../../etc/passwd", false, "traversal"},
		{"denied shadow", "/etc/shadow", false, "denylist"},
		{"ssh key", "/root/.ssh/id_rsa", false, "denylist"},

		// Suspicious patterns
		{"git hooks", "/home/simon/.git/hooks", false, "suspicious"},
		{"github workflows", "/project/.github/workflows/main.yml", false, "suspicious"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePathStrict(tt.path)
			if result.Valid != tt.wantOK {
				t.Errorf("ValidatePathStrict(%q) valid = %v, want %v", tt.path, result.Valid, tt.wantOK)
			}
			if !tt.wantOK && tt.wantMsg != "" && !contains(result.Message, tt.wantMsg) {
				t.Errorf("ValidatePathStrict(%q) message = %q, want containing %q", tt.path, result.Message, tt.wantMsg)
			}
		})
	}
}

func TestValidateCommandStrict(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		wantOK  bool
		wantMsg string
	}{
		// Valid commands
		{"simple ls", "ls -la", true, ""},
		{"cat file", "cat /tmp/test.txt", true, ""},
		{"git status", "git status", true, ""},

		// Dangerous commands
		{"rm rf root", "rm -rf /", false, "blocklist"},
		{"dd if", "dd if=/dev/zero of=/tmp/test", false, "blocklist"},
		{"fork bomb", ":(){ :|:& };:", false, "blocklist"},
		{"pipe to sh", "echo test | sh", false, "pipe to shell"},
		{"command substitution", "ls $(pwd)", false, "substitution"},
		{"backtick substitution", "ls `whoami`", false, "substitution"},
		{"env var", "echo $HOME", false, "environment variable"},
		{"null byte", "ls\x00 -la", false, "null byte"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCommandStrict(tt.cmd)
			if result.Valid != tt.wantOK {
				t.Errorf("ValidateCommandStrict(%q) valid = %v, want %v, msg: %s", tt.cmd, result.Valid, tt.wantOK, result.Message)
			}
			if !tt.wantOK && tt.wantMsg != "" && !contains(result.Message, tt.wantMsg) {
				t.Errorf("ValidateCommandStrict(%q) message = %q, want containing %q", tt.cmd, result.Message, tt.wantMsg)
			}
		})
	}
}

func TestSanitizePathStrict(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  /home/simon  ", "/home/simon"},
		{"/home/simon/../etc/passwd", "/home/etc/passwd"}, // filepath.Clean normalizes
		{"/home/simon\x00bad", "/home/simonbad"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizePathStrict(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizePathStrict(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContentValidator(t *testing.T) {
	v := DefaultContentValidator()

	tests := []struct {
		name     string
		content  []byte
		filename string
		wantOK   bool
	}{
		{"text file", []byte("hello world"), "test.txt", true},
		{"empty file", []byte(""), "empty.txt", true},
		{"binary extension", []byte("binary"), "test.exe", false},
		{"dll extension", []byte("dll content"), "test.dll", false},
		{"shebang", []byte("#!/bin/bash\necho hi"), "test.sh", false},
		{"null bytes", []byte("hello\x00world"), "test.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateContent(tt.content, tt.filename)
			if result.Valid != tt.wantOK {
				t.Errorf("ValidateContent(%q, %q) valid = %v, want %v", tt.content, tt.filename, result.Valid, tt.wantOK)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
