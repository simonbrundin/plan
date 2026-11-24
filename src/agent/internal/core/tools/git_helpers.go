package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
)

// toJSON returns a JSON string representation of a value
func toJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

// CommandOutput represents the output of a shell command
type CommandOutput struct {
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exitCode"`
}

// captureCommand executes a command and returns its output
func captureCommand(ctx context.Context, name string, args ...string) (*CommandOutput, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return &CommandOutput{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
