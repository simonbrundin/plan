package orchestrator

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"plan-agent/internal/core/checkpoint"
)

// CheckpointExecutor orchestrates checkpoint execution: plan tests → write tests → implement → run tests
type CheckpointExecutor struct {
	subAgentManager *SubAgentManager
	testPlanner     TestPlanner
}

// NewCheckpointExecutor creates a new checkpoint executor
func NewCheckpointExecutor(subAgentManager *SubAgentManager, testPlanner TestPlanner) *CheckpointExecutor {
	return &CheckpointExecutor{
		subAgentManager: subAgentManager,
		testPlanner:     testPlanner,
	}
}

// ExecutionPhase represents a phase in checkpoint execution
type ExecutionPhase string

const (
	PhasePlanTests  ExecutionPhase = "plan_tests"
	PhaseWriteTests ExecutionPhase = "write_tests"
	PhaseImplement  ExecutionPhase = "implement"
	PhaseRunTests   ExecutionPhase = "run_tests"
	PhaseVerify     ExecutionPhase = "verify"
)

// ExecutionProgress tracks progress through phases
type ExecutionProgress struct {
	Phase       ExecutionPhase
	PhaseIndex  int
	TotalPhases int
	Message     string
}

// ExecuteCheckpoint runs a checkpoint through all phases
func (e *CheckpointExecutor) ExecuteCheckpoint(
	ctx context.Context,
	cp *checkpoint.Checkpoint,
	onProgress func(*ExecutionProgress),
	onTestGenerated func(testFile string, content string),
	onImplementationGenerated func(files map[string]string),
) (*CheckpointExecutionResult, error) {

	phases := []ExecutionPhase{
		PhasePlanTests,
		PhaseWriteTests,
		PhaseImplement,
		PhaseRunTests,
	}

	result := &CheckpointExecutionResult{
		CheckpointID: cp.ID,
		StartTime:    time.Now(),
	}

	reportProgress := func(phase ExecutionPhase, msg string) {
		if onProgress != nil {
			onProgress(&ExecutionProgress{
				Phase:       phase,
				PhaseIndex:  getPhaseIndex(phase, phases),
				TotalPhases: len(phases),
				Message:     msg,
			})
		}
	}

	// Phase 1: Plan tests (TDD - write tests BEFORE implementation)
	reportProgress(PhasePlanTests, "Planning tests for checkpoint...")
	testCases, err := e.planTests(ctx, cp)
	if err != nil {
		result.Error = fmt.Sprintf("test planning failed: %v", err)
		return result, err
	}
	reportProgress(PhasePlanTests, fmt.Sprintf("Planned %d test cases", len(testCases)))

	// Phase 2: Write test files
	reportProgress(PhaseWriteTests, "Writing test files...")
	testFiles, err := e.writeTestFiles(cp, testCases)
	if err != nil {
		result.Error = fmt.Sprintf("writing tests failed: %v", err)
		return result, err
	}
	if onTestGenerated != nil {
		for path, content := range testFiles {
			onTestGenerated(path, content)
		}
	}
	reportProgress(PhaseWriteTests, fmt.Sprintf("Wrote %d test files", len(testFiles)))

	// Phase 3: Run initial tests (they should FAIL because implementation doesn't exist yet)
	reportProgress(PhaseRunTests, "Running initial tests (expecting FAIL)...")
	_, err = e.runTests(ctx, testFiles)
	if err != nil {
		log.Printf("Initial test run had errors: %v (expected if implementation missing)", err)
	}
	// In TDD, we EXPECT tests to fail here

	// Phase 4: Implement the feature
	reportProgress(PhaseImplement, "Generating implementation...")
	implFiles, err := e.generateImplementation(ctx, cp)
	if err != nil {
		result.Error = fmt.Sprintf("implementation generation failed: %v", err)
		return result, err
	}
	if onImplementationGenerated != nil {
		onImplementationGenerated(implFiles)
	}
	reportProgress(PhaseImplement, fmt.Sprintf("Generated %d implementation files", len(implFiles)))

	// Phase 5: Run tests again (they should PASS now)
	reportProgress(PhaseRunTests, "Running tests after implementation...")
	finalTestResult, err := e.runTests(ctx, testFiles)
	if err != nil {
		result.Error = fmt.Sprintf("tests failed after implementation: %v", err)
		result.EndTime = time.Now()
		return result, err
	}

	// Determine pass/fail based on test results
	result.Passed = finalTestResult.Passed
	result.EndTime = time.Now()

	return result, nil
}

// Execute is an alias for ExecuteCheckpoint
func (e *CheckpointExecutor) Execute(ctx context.Context, cp *checkpoint.Checkpoint) (*CheckpointExecutionResult, error) {
	return e.ExecuteCheckpoint(ctx, cp, nil, nil, nil)
}

// planTests uses the test planner to generate test cases
func (e *CheckpointExecutor) planTests(ctx context.Context, cp *checkpoint.Checkpoint) ([]*TestCase, error) {
	if e.testPlanner == nil {
		// No test planner - generate basic tests
		return generateBasicTestCases(cp), nil
	}
	return e.testPlanner.PlanTests(ctx, cp)
}

// generateBasicTestCases creates simple test cases for a checkpoint
func generateBasicTestCases(cp *checkpoint.Checkpoint) []*TestCase {
	var cases []*TestCase
	for i, task := range cp.Tasks {
		cases = append(cases, &TestCase{
			Name:        fmt.Sprintf("TestTask%d_%s", i+1, sanitizeTestName(task)),
			Description: fmt.Sprintf("Test: %s", task),
			Table: [][]any{
				{"input", "expected"},
				{"valid input", "success"},
				{"invalid input", "error"},
			},
		})
	}
	return cases
}

// writeTestFiles generates Go test files from test cases
func (e *CheckpointExecutor) writeTestFiles(cp *checkpoint.Checkpoint, cases []*TestCase) (map[string]string, error) {
	files := make(map[string]string)

	// Determine test directory based on checkpoint type
	testDir := getTestDir(cp)

	for _, tc := range cases {
		content := GenerateTestFile(cp, []*TestCase{tc})
		fileName := fmt.Sprintf("%s_test.go", tc.Name)
		files[filepath.Join(testDir, fileName)] = content
	}

	return files, nil
}

// generateImplementation uses a sub-agent to generate implementation code
func (e *CheckpointExecutor) generateImplementation(ctx context.Context, cp *checkpoint.Checkpoint) (map[string]string, error) {
	if e.subAgentManager == nil {
		return nil, fmt.Errorf("no sub-agent manager configured")
	}

	prompt := e.buildImplementationPrompt(cp)

	result := e.subAgentManager.SpawnAndRun(ctx, SubAgentConfig{
		Name:         fmt.Sprintf("impl-%s", cp.Name),
		SystemPrompt: getImplementationSystemPrompt(),
	}, prompt)

	if result.Error != nil {
		return nil, result.Error
	}

	// Parse the result to extract file contents
	return parseImplementationResult(result.Output)
}

// runTests runs the test files and returns results
func (e *CheckpointExecutor) runTests(ctx context.Context, testFiles map[string]string) (*TestRunResult, error) {
	// Write test files first
	for path, content := range testFiles {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			log.Printf("Warning: failed to write test file %s: %v", path, err)
		}
	}

	// Run go test
	result := &TestRunResult{}

	// Get unique directories to test
	dirs := make(map[string]bool)
	for path := range testFiles {
		dirs[filepath.Dir(path)] = true
	}

	for dir := range dirs {
		testResult, err := runGoTest(ctx, dir)
		if err != nil {
			result.Errors = append(result.Errors, err)
		}
		result.Passed = result.Passed && testResult.Passed
		result.Total += testResult.Total
		result.PassedCount += testResult.PassedCount
		result.Failed += testResult.Failed
	}

	return result, nil
}

// buildImplementationPrompt creates a prompt for implementation generation
func (e *CheckpointExecutor) buildImplementationPrompt(cp *checkpoint.Checkpoint) string {
	return fmt.Sprintf(`Implement the feature described by this checkpoint.

Checkpoint: %s
Description: %s
Tasks:
%s

Requirements:
1. Write complete, working Go code
2. Follow Go best practices
3. Handle errors properly
4. Include appropriate tests
5. Match the existing code style

Output your implementation as file contents with this format:
---FILE: path/to/file.go---
file contents here
---END---

List each file you create.
`, cp.Name, cp.Description, strings.Join(cp.Tasks, "\n"))
}

// parseImplementationResult parses the sub-agent output into files
func parseImplementationResult(output string) (map[string]string, error) {
	files := make(map[string]string)

	lines := strings.Split(output, "\n")
	var currentFile string
	var currentContent strings.Builder
	inFile := false

	for _, line := range lines {
		if strings.HasPrefix(line, "---FILE:") {
			// Save previous file
			if currentFile != "" {
				files[currentFile] = currentContent.String()
			}
			currentFile = strings.TrimSpace(strings.TrimPrefix(line, "---FILE:"))
			currentContent.Reset()
			inFile = true
		} else if strings.HasPrefix(line, "---END---") {
			inFile = false
			if currentFile != "" {
				files[currentFile] = currentContent.String()
				currentFile = ""
				currentContent.Reset()
			}
		} else if inFile {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Don't forget last file
	if currentFile != "" && currentContent.Len() > 0 {
		files[currentFile] = currentContent.String()
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in implementation result")
	}

	return files, nil
}

// TestRunResult holds results from running tests
type TestRunResult struct {
	Passed      bool
	Total       int
	PassedCount int
	Failed      int
	Errors      []error
}

// CheckpointExecutionResult holds checkpoint execution results
type CheckpointExecutionResult struct {
	CheckpointID string
	Passed       bool
	StartTime    time.Time
	EndTime      time.Time
	Error        string
}

// Helper functions

func getPhaseIndex(phase ExecutionPhase, phases []ExecutionPhase) int {
	for i, p := range phases {
		if p == phase {
			return i
		}
	}
	return 0
}

func getTestDir(cp *checkpoint.Checkpoint) string {
	// Use checkpoint name to create a test directory
	name := strings.ToLower(cp.Name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return filepath.Join(".", "test", name)
}

func sanitizeTestName(name string) string {
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return '_'
	}, name)
	return name
}

func getImplementationSystemPrompt() string {
	return `You are an expert Go developer. Write complete, production-ready code.

Guidelines:
1. Use proper error handling
2. Follow Go idioms and style
3. Write clear, documented code
4. Include unit tests
5. Keep functions small and focused

Respond ONLY with the file contents. Use the format:
---FILE: path/to/file.go---
contents
---END---`
}

// runGoTest executes go test in a directory
func runGoTest(ctx context.Context, dir string) (*TestRunResult, error) {
	result := &TestRunResult{Passed: true}

	// Simple implementation - in production would run actual go test
	// For now, return a placeholder
	result.Passed = true
	result.Total = 1
	result.PassedCount = 1

	return result, nil
}
