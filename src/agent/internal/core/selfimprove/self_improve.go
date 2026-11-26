package selfimprove

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"plan/agent/internal/core/memory"
)

// SelfAnalyzer analyzes code for improvement opportunities
type SelfAnalyzer struct {
	store *memory.MemoryStore
}

// NewSelfAnalyzer creates a new self-analyzer
func NewSelfAnalyzer(store *memory.MemoryStore) *SelfAnalyzer {
	return &SelfAnalyzer{store: store}
}

// AnalysisResult represents the result of analyzing code
type AnalysisResult struct {
	FilePath    string
	Issues      []CodeIssue
	Suggestions []ImprovementSuggestion
	Metrics     CodeMetrics
}

// CodeIssue represents a code issue found
type CodeIssue struct {
	Type        IssueType
	Severity    int // 1-5, 5 being most severe
	Line        int
	Description string
	Pattern     string
}

// IssueType represents the type of issue
type IssueType string

const (
	IssueTypeComplexity      IssueType = "complexity"
	IssueTypeDuplication     IssueType = "duplication"
	IssueTypePerformance     IssueType = "performance"
	IssueTypeMaintainability IssueType = "maintainability"
	IssueTypeSecurity        IssueType = "security"
)

// ImprovementSuggestion represents a suggested improvement
type ImprovementSuggestion struct {
	Type        string
	Priority    int // 1-5, 5 being highest
	Description string
	Before      string
	After       string
	Effort      string // "low", "medium", "high"
	Impact      string // "low", "medium", "high"
}

// CodeMetrics holds metrics about the code
type CodeMetrics struct {
	LinesOfCode       int
	FunctionCount     int
	AverageComplexity float64
	MaxComplexity     int
	CommentRatio      float64
	DuplicationScore  float64
	TestCoverage      float64
}

// AnalyzeFile analyzes a single Go file for issues
func (a *SelfAnalyzer) AnalyzeFile(filePath string) (*AnalysisResult, error) {
	result := &AnalysisResult{
		FilePath:    filePath,
		Issues:      make([]CodeIssue, 0),
		Suggestions: make([]ImprovementSuggestion, 0),
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Calculate metrics
	result.Metrics = a.calculateMetrics(node, filePath)

	// Find issues
	result.Issues = a.findIssues(node, fset)

	// Generate suggestions
	result.Suggestions = a.generateSuggestions(result.Issues, result.Metrics)

	return result, nil
}

// calculateMetrics calculates code metrics
func (a *SelfAnalyzer) calculateMetrics(node *ast.File, filePath string) CodeMetrics {
	metrics := CodeMetrics{}

	// Count lines of code
	if data, err := os.ReadFile(filePath); err == nil {
		lines := strings.Split(string(data), "\n")
		metrics.LinesOfCode = len(lines)

		// Count comment lines
		commentLines := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
				commentLines++
			}
		}
		metrics.CommentRatio = float64(commentLines) / float64(metrics.LinesOfCode)
	}

	// Count functions
	ast.Inspect(node, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.FuncDecl:
			metrics.FunctionCount++
		}
		return true
	})

	return metrics
}

// findIssues finds code issues using AST analysis
func (a *SelfAnalyzer) findIssues(node *ast.File, fset *token.FileSet) []CodeIssue {
	var issues []CodeIssue

	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncDecl:
			// Check for long functions (complexity)
			if numStmts := len(n.Body.List); numStmts > 50 {
				issues = append(issues, CodeIssue{
					Type:        IssueTypeComplexity,
					Severity:    3,
					Line:        fset.Position(n.Pos()).Line,
					Description: fmt.Sprintf("Function '%s' has %d statements (high complexity)", n.Name.Name, numStmts),
					Pattern:     "long-function",
				})
			}

			// Check for deeply nested ifs
			if depth := nestedIfDepth(n.Body); depth > 4 {
				issues = append(issues, CodeIssue{
					Type:        IssueTypeComplexity,
					Severity:    4,
					Line:        fset.Position(n.Pos()).Line,
					Description: fmt.Sprintf("Function '%s' has deeply nested ifs (depth: %d)", n.Name.Name, depth),
					Pattern:     "deep-nesting",
				})
			}

		case *ast.IfStmt:
			// Check for empty if bodies
			if len(n.Body.List) == 0 {
				issues = append(issues, CodeIssue{
					Type:        IssueTypeMaintainability,
					Severity:    2,
					Line:        fset.Position(n.Pos()).Line,
					Description: "Empty if statement",
					Pattern:     "empty-if",
				})
			}
		}

		return true
	})

	return issues
}

// nestedIfDepth calculates the nesting depth of if statements
func nestedIfDepth(n ast.Node) int {
	maxDepth := 0
	var walk func(n ast.Node, depth int)
	walk = func(n ast.Node, depth int) {
		if depth > maxDepth {
			maxDepth = depth
		}
		ast.Inspect(n, func(child ast.Node) bool {
			if ifStmt, ok := child.(*ast.IfStmt); ok {
				walk(ifStmt.Body, depth+1)
				if ifStmt.Else != nil {
					walk(ifStmt.Else.(ast.Node), depth+1)
				}
				return false
			}
			return true
		})
	}
	walk(n, 0)
	return maxDepth
}

// generateSuggestions generates improvement suggestions based on issues
func (a *SelfAnalyzer) generateSuggestions(issues []CodeIssue, metrics CodeMetrics) []ImprovementSuggestion {
	var suggestions []ImprovementSuggestion

	// Group issues by type
	issueCounts := make(map[IssueType]int)
	for _, issue := range issues {
		issueCounts[issue.Type]++
	}

	// Generate suggestions based on issue counts
	if count := issueCounts[IssueTypeComplexity]; count > 3 {
		suggestions = append(suggestions, ImprovementSuggestion{
			Type:        "refactor",
			Priority:    4,
			Description: fmt.Sprintf("Found %d complexity issues. Consider extracting smaller functions.", count),
			Effort:      "medium",
			Impact:      "high",
		})
	}

	if metrics.AverageComplexity > 10 {
		suggestions = append(suggestions, ImprovementSuggestion{
			Type:        "simplify",
			Priority:    5,
			Description: fmt.Sprintf("Average complexity is %.1f. High complexity makes code hard to maintain.", metrics.AverageComplexity),
			Effort:      "high",
			Impact:      "high",
		})
	}

	if metrics.CommentRatio < 0.1 {
		suggestions = append(suggestions, ImprovementSuggestion{
			Type:        "document",
			Priority:    3,
			Description: fmt.Sprintf("Low comment ratio (%d%%). Add documentation.", int(metrics.CommentRatio*100)),
			Effort:      "medium",
			Impact:      "medium",
		})
	}

	return suggestions
}

// AnalyzeDirectory analyzes all Go files in a directory
func (a *SelfAnalyzer) AnalyzeDirectory(dirPath string) ([]*AnalysisResult, error) {
	var results []*AnalysisResult

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Skip test files and generated files
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}

			result, err := a.AnalyzeFile(path)
			if err != nil {
				// Log but continue
				return nil
			}
			results = append(results, result)
		}
		return nil
	})

	return results, err
}

// ImprovementEngine generates and prioritizes improvements
type ImprovementEngine struct {
	analyzer *SelfAnalyzer
	store    *memory.MemoryStore
}

// NewImprovementEngine creates a new improvement engine
func NewImprovementEngine(store *memory.MemoryStore) *ImprovementEngine {
	return &ImprovementEngine{
		analyzer: NewSelfAnalyzer(store),
		store:    store,
	}
}

// Improvement represents a planned improvement
type Improvement struct {
	ID          string
	Type        string
	Priority    int
	Description string
	Changes     []CodeChange
	Status      string // "planned", "approved", "applied", "verified"
	CreatedAt   time.Time
	AppliedAt   time.Time
}

// CodeChange represents a change to be made
type CodeChange struct {
	FilePath string
	OldCode  string
	NewCode  string
	Reason   string
}

// GenerateImprovements analyzes code and generates improvement suggestions
func (e *ImprovementEngine) GenerateImprovements(ctx context.Context, dirPath string) ([]*Improvement, error) {
	results, err := e.analyzer.AnalyzeDirectory(dirPath)
	if err != nil {
		return nil, err
	}

	var improvements []*Improvement

	for _, result := range results {
		for _, suggestion := range result.Suggestions {
			imp := &Improvement{
				ID:          fmt.Sprintf("imp-%d", time.Now().UnixNano()),
				Type:        suggestion.Type,
				Priority:    suggestion.Priority,
				Description: suggestion.Description,
				Changes:     e.planChanges(result, suggestion),
				Status:      "planned",
				CreatedAt:   time.Now(),
			}
			improvements = append(improvements, imp)
		}
	}

	// Sort by priority
	sort.Slice(improvements, func(i, j int) bool {
		return improvements[i].Priority > improvements[j].Priority
	})

	return improvements, nil
}

// planChanges plans the code changes for an improvement
func (e *ImprovementEngine) planChanges(result *AnalysisResult, suggestion ImprovementSuggestion) []CodeChange {
	var changes []CodeChange

	switch suggestion.Type {
	case "refactor":
		// Add refactoring changes
		changes = append(changes, CodeChange{
			FilePath: result.FilePath,
			OldCode:  "func (", // Placeholder
			NewCode:  "func (", // Placeholder
			Reason:   suggestion.Description,
		})
	case "document":
		changes = append(changes, CodeChange{
			FilePath: result.FilePath,
			OldCode:  "",
			NewCode:  "// TODO: Add documentation",
			Reason:   suggestion.Description,
		})
	}

	return changes
}

// ApplyImprovement applies a planned improvement
func (e *ImprovementEngine) ApplyImprovement(imp *Improvement) error {
	for _, change := range imp.Changes {
		if change.OldCode == "" {
			// Insert new code
			if err := e.insertCode(change.FilePath, change.NewCode); err != nil {
				return err
			}
		} else {
			// Replace existing code
			if err := e.replaceCode(change.FilePath, change.OldCode, change.NewCode); err != nil {
				return err
			}
		}
	}

	imp.Status = "applied"
	imp.AppliedAt = time.Now()

	// Record the improvement in memory
	if _, err := e.store.AddProcedural(
		imp.Description,
		fmt.Sprintf("Applied improvement: %s", imp.Description),
		fmt.Sprintf("Priority: %d, Type: %s", imp.Priority, imp.Type),
		[]string{"improvement", imp.Type, "applied"},
	); err != nil {
		log.Printf("failed to record improvement: %v", err)
	}

	return nil
}

// insertCode inserts new code at the end of a file
func (e *ImprovementEngine) insertCode(filePath, newCode string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	newContent := string(data) + "\n" + newCode + "\n"
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// replaceCode replaces old code with new code in a file
func (e *ImprovementEngine) replaceCode(filePath, oldCode, newCode string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	newContent := strings.Replace(string(data), oldCode, newCode, 1)
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// SelfImprover orchestrates the self-improvement process
type SelfImprover struct {
	engine *ImprovementEngine
	store  *memory.MemoryStore
}

// NewSelfImprover creates a new self-improver
func NewSelfImprover(store *memory.MemoryStore) *SelfImprover {
	return &SelfImprover{
		engine: NewImprovementEngine(store),
		store:  store,
	}
}

// RunSelfCheck runs a self-check and returns findings
func (s *SelfImprover) RunSelfCheck(ctx context.Context, codePaths []string) (*SelfCheckReport, error) {
	report := &SelfCheckReport{
		Timestamp:    time.Now(),
		Findings:     make([]Finding, 0),
		Improvements: make([]*Improvement, 0),
	}

	for _, path := range codePaths {
		improvements, err := s.engine.GenerateImprovements(ctx, path)
		if err != nil {
			continue
		}

		for _, imp := range improvements {
			report.Improvements = append(report.Improvements, imp)
			report.Findings = append(report.Findings, Finding{
				Type:        imp.Type,
				Description: imp.Description,
				Priority:    imp.Priority,
				Status:      imp.Status,
			})
		}
	}

	// Sort findings by priority
	sort.Slice(report.Findings, func(i, j int) bool {
		return report.Findings[i].Priority > report.Findings[j].Priority
	})

	// Sort improvements by priority
	sort.Slice(report.Improvements, func(i, j int) bool {
		return report.Improvements[i].Priority > report.Improvements[j].Priority
	})

	return report, nil
}

// SelfCheckReport represents the report from a self-check
type SelfCheckReport struct {
	Timestamp    time.Time
	Findings     []Finding
	Improvements []*Improvement
	TotalIssues  int
	HighPriority int
}

// Finding represents a single finding
type Finding struct {
	Type        string
	Description string
	Priority    int
	Status      string
}

// LearnFromExperience learns from past experiences to improve future performance
func (s *SelfImprover) LearnFromExperience(task string, success bool, lesson string) error {
	_, err := s.store.AddEpisodic(
		fmt.Sprintf("task-%d", time.Now().UnixNano()),
		fmt.Sprintf("Task: %s, Success: %v, Lesson: %s", task, success, lesson),
		lesson,
		map[string]interface{}{
			"task":    task,
			"success": success,
		},
	)
	return err
}

// GetImprovementHistory returns history of improvements
func (s *SelfImprover) GetImprovementHistory() []*Improvement {
	// Get recent procedural memories that are improvements
	memories := s.store.GetRecent(memory.MemoryTypeProcedural, 20)

	var improvements []*Improvement
	for _, mem := range memories {
		if containsString(mem.Tags, "improvement") {
			improvements = append(improvements, &Improvement{
				ID:          mem.ID,
				Description: mem.Summary,
				Type:        getTagValue(mem.Tags, "improvement"),
				Status:      "applied",
				CreatedAt:   mem.CreatedAt,
			})
		}
	}

	return improvements
}

// containsString checks if a string slice contains a value
func containsString(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// getTagValue extracts a value from tags (format: "key:value")
func getTagValue(tags []string, key string) string {
	for _, tag := range tags {
		parts := strings.SplitN(tag, ":", 2)
		if len(parts) == 2 && parts[0] == key {
			return parts[1]
		}
	}
	return ""
}

// PatternDetector detects common patterns that can be improved
type PatternDetector struct{}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector() *PatternDetector {
	return &PatternDetector{}
}

// DetectedPattern represents a detected pattern
type DetectedPattern struct {
	Name        string
	Description string
	Occurrences int
	Suggestion  string
}

// DetectPatterns scans code for common patterns
func (d *PatternDetector) DetectPatterns(dirPath string) []DetectedPattern {
	var patterns []DetectedPattern

	// Detect error handling patterns
	errorPatterns := d.detectErrorHandling(dirPath)
	patterns = append(patterns, errorPatterns...)

	// Detect logging patterns
	logPatterns := d.detectLogging(dirPath)
	patterns = append(patterns, logPatterns...)

	// Detect context usage patterns
	contextPatterns := d.detectContextUsage(dirPath)
	patterns = append(patterns, contextPatterns...)

	return patterns
}

// detectErrorHandling detects error handling patterns
func (d *PatternDetector) detectErrorHandling(dirPath string) []DetectedPattern {
	var patterns []DetectedPattern

	// Pattern: ignoring errors with _
	ignoringErrPattern := regexp.MustCompile(`_ = .*Err`)
	matches := d.searchFiles(dirPath, ignoringErrPattern)
	if len(matches) > 0 {
		patterns = append(patterns, DetectedPattern{
			Name:        "ignored-error",
			Description: "Error values are being ignored",
			Occurrences: len(matches),
			Suggestion:  "Handle errors properly instead of ignoring them",
		})
	}

	return patterns
}

// detectLogging detects logging patterns
func (d *PatternDetector) detectLogging(dirPath string) []DetectedPattern {
	var patterns []DetectedPattern

	// Pattern: missing error logging
	noLogErrPattern := regexp.MustCompile(`if err != nil \{\s*return`)
	matches := d.searchFiles(dirPath, noLogErrPattern)
	if len(matches) > 2 {
		patterns = append(patterns, DetectedPattern{
			Name:        "silent-errors",
			Description: "Errors returned without logging",
			Occurrences: len(matches),
			Suggestion:  "Add logging before returning errors",
		})
	}

	return patterns
}

// detectContextUsage detects context usage patterns
func (d *PatternDetector) detectContextUsage(dirPath string) []DetectedPattern {
	var patterns []DetectedPattern

	// Pattern: context.Background() in production code
	bgPattern := regexp.MustCompile(`context\.Background\(\)`)
	matches := d.searchFiles(dirPath, bgPattern)
	if len(matches) > 0 {
		patterns = append(patterns, DetectedPattern{
			Name:        "context-background",
			Description: "Using context.Background() instead of request context",
			Occurrences: len(matches),
			Suggestion:  "Pass context from the request instead of using Background()",
		})
	}

	return patterns
}

// searchFiles searches all Go files in a directory for a pattern
func (d *PatternDetector) searchFiles(dirPath string, pattern *regexp.Regexp) []string {
	var matches []string

	_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		if data, err := os.ReadFile(path); err == nil {
			if pattern.Match(data) {
				matches = append(matches, path)
			}
		}
		return nil
	})

	return matches
}

// PrettyPrint prints a report in a readable format
func (r *SelfCheckReport) PrettyPrint() string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString("                    SELF-IMPROVEMENT REPORT                        \n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	_, _ = fmt.Fprintf(&sb, "Timestamp: %s\n\n", r.Timestamp.Format(time.RFC822))

	_, _ = fmt.Fprintf(&sb, "Total Issues Found: %d\n", len(r.Findings))
	_, _ = fmt.Fprintf(&sb, "High Priority: %d\n\n", r.HighPriority)

	if len(r.Findings) > 0 {
		sb.WriteString("───────────────────────────────────────────────────────────────\n")
		sb.WriteString("TOP FINDINGS\n")
		sb.WriteString("───────────────────────────────────────────────────────────────\n\n")

		for i, f := range r.Findings {
			if i >= 10 { // Show top 10
				_, _ = fmt.Fprintf(&sb, "\n... and %d more findings\n", len(r.Findings)-10)
				break
			}

			priority := "LOW"
			if f.Priority >= 4 {
				priority = "HIGH"
			} else if f.Priority >= 3 {
				priority = "MED"
			}

			_, _ = fmt.Fprintf(&sb, "[%s] #%d: %s\n", priority, i+1, f.Description)
		}
	}

	sb.WriteString("\n═══════════════════════════════════════════════════════════════\n")

	return sb.String()
}
