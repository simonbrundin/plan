package tools

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"
	"time"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

func bashExecute(_ string, args map[string]any, signal <-chan struct{}, onUpdate func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	command, ok := args["command"].(string)
	if !ok {
		return errorResult("command must be a string"), nil
	}

	// Validate command for safety
	if result := ValidateCommandStrict(command); !result.Valid {
		return errorResult("invalid command: " + result.Message), nil
	}

	timeout := getFloatArg(args, "timeout", 60)
	cwd := getStringArg(args, "cwd", ".")
	stream := getBoolArg(args, "stream", true) // Enable streaming by default

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the timeout in a goroutine
	timeoutCh := time.After(time.Duration(timeout) * time.Second)

	select {
	case <-signal:
		return cancelledResult("Command cancelled"), nil
	case <-timeoutCh:
		return errorResult("command timed out"), nil
	default:
	}

	shell := "bash"
	cmd := exec.CommandContext(ctx, shell, "-c", command)
	if cwd != "." {
		cmd.Dir = cwd
	}

	// Create pipes for streaming output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errorResult("failed to create stdout pipe: " + err.Error()), nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errorResult("failed to create stderr pipe: " + err.Error()), nil
	}

	// Start command
	if err := cmd.Start(); err != nil {
		// Fallback: try with sh
		shell = "sh"
		cmd = exec.CommandContext(ctx, shell, "-c", command)
		if cwd != "." {
			cmd.Dir = cwd
		}

		stdout, err = cmd.StdoutPipe()
		if err != nil {
			return errorResult("failed to create stdout pipe: " + err.Error()), nil
		}

		stderr, err = cmd.StderrPipe()
		if err != nil {
			return errorResult("failed to create stderr pipe: " + err.Error()), nil
		}

		if err := cmd.Start(); err != nil {
			return errorResult("failed to start command: " + err.Error()), nil
		}
	}

	// Combined output buffer
	var combinedOutput strings.Builder

	// Stream reader that sends updates
	streamOutput := func(reader io.Reader, isError bool) {
		if !stream || onUpdate == nil {
			// Just buffer
			buf := make([]byte, 1024)
			for {
				n, err := reader.Read(buf)
				if n > 0 {
					combinedOutput.Write(buf[:n])
				}
				if err != nil {
					break
				}
			}
			return
		}

		// Stream with updates
		reader = bufio.NewReaderSize(reader, 4096)
		lineBuf := &strings.Builder{}

		for {
			select {
			case <-signal:
				return
			default:
			}

			// Read a chunk
			buf := make([]byte, 512)
			n, err := reader.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				combinedOutput.WriteString(chunk)
				lineBuf.WriteString(chunk)

				// Send update for each line or every ~1KB
				if lineBuf.Len() >= 1024 || strings.Contains(chunk, "\n") {
					lines := lineBuf.String()
					lineBuf.Reset()

					// Send partial update
					onUpdate(&provider.ToolResult{
						Content: []messages.ContentBlock{&messages.TextBlock{
							Text: lines,
						}},
						IsError: isError,
					})
				}
			}

			if err != nil {
				// Flush remaining
				if lineBuf.Len() > 0 {
					onUpdate(&provider.ToolResult{
						Content: []messages.ContentBlock{&messages.TextBlock{
							Text: lineBuf.String(),
						}},
						IsError: isError,
					})
				}
				break
			}
		}
	}

	// Merge stdout and stderr for streaming
	stdoutCh := make(chan struct{})
	stderrCh := make(chan struct{})

	go func() {
		streamOutput(stdout, false)
		close(stdoutCh)
	}()

	go func() {
		streamOutput(stderr, true)
		close(stderrCh)
	}()

	// Wait for command completion or cancellation/timeout
	var cmdErr error
	resultCh := make(chan struct{})

	go func() {
		cmdErr = cmd.Wait()
		close(resultCh)
	}()

	select {
	case <-signal:
		cancel()
		// Drain remaining output
		<-stdoutCh
		<-stderrCh
		return cancelledResult("Command cancelled"), nil

	case <-timeoutCh:
		cancel()
		// Drain remaining output
		<-stdoutCh
		<-stderrCh
		return errorResult("command timed out"), nil

	case <-resultCh:
		// Wait for output to finish draining
		<-stdoutCh
		<-stderrCh
	}

	output := combinedOutput.String()

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: output}},
		IsError: cmdErr != nil,
		Details: cmdErr,
	}, nil
}
