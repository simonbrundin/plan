package goals

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Decomposer breaks down high-level goals into actionable sub-goals
type Decomposer struct {
	llmProvider LLMProvider
}

// LLMProvider interface for LLM calls
type LLMProvider interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// NewDecomposer creates a new goal decomposer
func NewDecomposer(llmProvider LLMProvider) *Decomposer {
	return &Decomposer{llmProvider: llmProvider}
}

// DecompositionResult holds the result of goal decomposition
type DecompositionResult struct {
	Goal      *DecomposedGoal
	SubGoals  []*DecomposedGoal
	Tasks     []*Task
	Timeline  string
	Resources []string
	Risks     []string
}

// Goal represents a decomposed goal (renamed to avoid conflict with store.go)
type DecomposedGoal struct {
	ID                string
	Title             string
	Description       string
	Priority          int
	Complexity        ComplexityLevel
	Dependencies      []string // Goal IDs this depends on
	Skills            []string // Required skills
	EstimatedDuration string
}

// Task represents an actionable task within a goal
type Task struct {
	ID           string
	GoalID       string
	Title        string
	Description  string
	Type         TaskType
	Priority     int
	Dependencies []string // Task IDs this depends on
}

// ComplexityLevel represents the complexity of a goal
type ComplexityLevel int

const (
	ComplexityLow ComplexityLevel = iota
	ComplexityMedium
	ComplexityHigh
)

func (c ComplexityLevel) String() string {
	switch c {
	case ComplexityLow:
		return "low"
	case ComplexityMedium:
		return "medium"
	case ComplexityHigh:
		return "high"
	default:
		return "unknown"
	}
}

// TaskType represents the type of task
type TaskType string

const (
	TaskTypeImplementation TaskType = "implementation"
	TaskTypeResearch       TaskType = "research"
	TaskTypeDesign         TaskType = "design"
	TaskTypeTesting        TaskType = "testing"
	TaskTypeReview         TaskType = "review"
	TaskTypeDeployment     TaskType = "deployment"
)

// Decompose takes a high-level goal description and breaks it down
func (d *Decomposer) Decompose(ctx context.Context, goalDescription string) (*DecompositionResult, error) {
	// Generate sub-goals using LLM
	subGoals, err := d.generateSubGoals(ctx, goalDescription)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sub-goals: %w", err)
	}

	// Generate tasks for each sub-goal
	var allTasks []*Task
	for _, sg := range subGoals {
		tasks, err := d.generateTasks(ctx, sg)
		if err != nil {
			continue // Log but continue
		}
		allTasks = append(allTasks, tasks...)
	}

	// Estimate timeline
	timeline := d.estimateTimeline(subGoals, allTasks)

	// Identify resources needed
	resources := d.identifyResources(subGoals, allTasks)

	// Identify risks
	risks := d.identifyRisks(subGoals)

	// Create main goal
	mainGoal := &DecomposedGoal{
		ID:          uuid.New().String(),
		Title:       goalDescription,
		Description: fmt.Sprintf("Decomposed goal: %s", goalDescription),
		Priority:    1,
		Complexity:  d.calculateComplexity(subGoals),
	}

	return &DecompositionResult{
		Goal:      mainGoal,
		SubGoals:  subGoals,
		Tasks:     allTasks,
		Timeline:  timeline,
		Resources: resources,
		Risks:     risks,
	}, nil
}

// generateSubGoals generates sub-goals using LLM
func (d *Decomposer) generateSubGoals(ctx context.Context, goalDescription string) ([]*DecomposedGoal, error) {
	prompt := fmt.Sprintf(`Analyze this goal and break it down into 3-7 logical sub-goals.

Goal: %s

Requirements:
1. Each sub-goal should be a distinct, meaningful chunk of work
2. Order them in a logical dependency order (foundational first)
3. Each sub-goal should be achievable in 1-4 hours

Output format (JSON array):
[
  {
    "title": "Clear, concise title",
    "description": "What this sub-goal accomplishes",
    "priority": 1-5 (1 is highest),
    "complexity": "low|medium|high",
    "skills": ["skill1", "skill2"],
    "estimated_duration": "1-4 hours"
  }
]

Only output JSON, no markdown formatting.`, goalDescription)

	response, err := d.llmProvider.Complete(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	return d.parseSubGoals(response)
}

// parseSubGoals parses the LLM response into sub-goals
func (d *Decomposer) parseSubGoals(response string) ([]*DecomposedGoal, error) {
	// Clean response
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Parse JSON (simplified - in production use proper JSON parsing)
	// For now, create goals from text response
	var goals []*DecomposedGoal

	lines := strings.Split(response, "\n")
	currentGoal := &DecomposedGoal{ID: uuid.New().String(), Priority: 1}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"title"`) {
			title := extractJSONValue(line, "title")
			if title != "" {
				if currentGoal.Title != "" {
					goals = append(goals, currentGoal)
				}
				currentGoal = &DecomposedGoal{ID: uuid.New().String(), Title: title, Priority: len(goals) + 1}
			}
		} else if strings.Contains(line, `"description"`) {
			currentGoal.Description = extractJSONValue(line, "description")
		} else if strings.Contains(line, `"complexity"`) {
			complexity := extractJSONValue(line, "complexity")
			currentGoal.Complexity = parseComplexity(complexity)
		} else if strings.Contains(line, `"estimated_duration"`) {
			currentGoal.EstimatedDuration = extractJSONValue(line, "estimated_duration")
		} else if strings.Contains(line, `"skills"`) {
			currentGoal.Skills = extractJSONArray(line, "skills")
		}
	}

	if currentGoal.Title != "" {
		goals = append(goals, currentGoal)
	}

	// Ensure we have at least some goals
	if len(goals) == 0 {
		goals = append(goals, &DecomposedGoal{
			ID:          uuid.New().String(),
			Title:       "Main Task",
			Description: "Primary task from goal",
			Priority:    1,
			Complexity:  ComplexityMedium,
		})
	}

	return goals, nil
}

// generateTasks generates tasks for a sub-goal
func (d *Decomposer) generateTasks(ctx context.Context, goal *DecomposedGoal) ([]*Task, error) {
	prompt := fmt.Sprintf(`Break down this sub-goal into 3-5 actionable tasks.

Goal: %s
Description: %s

Requirements:
1. Each task should be a single, actionable step
2. Order tasks in execution order
3. Each task should take 15-60 minutes

Output format (JSON array):
[
  {
    "title": "Clear task title",
    "description": "What to do",
    "type": "implementation|research|design|testing|review|deployment"
  }
]

Only output JSON, no markdown formatting.`, goal.Title, goal.Description)

	response, err := d.llmProvider.Complete(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return d.parseTasks(response, goal.ID)
}

// parseTasks parses the LLM response into tasks
func (d *Decomposer) parseTasks(response, goalID string) ([]*Task, error) {
	// Clean response
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var tasks []*Task

	lines := strings.Split(response, "\n")
	currentTask := &Task{ID: uuid.New().String(), GoalID: goalID}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, `"title"`) {
			title := extractJSONValue(line, "title")
			if title != "" {
				if currentTask.Title != "" {
					tasks = append(tasks, currentTask)
				}
				currentTask = &Task{ID: uuid.New().String(), GoalID: goalID, Priority: len(tasks) + 1}
				currentTask.Title = title
			}
		} else if strings.Contains(line, `"description"`) {
			currentTask.Description = extractJSONValue(line, "description")
		} else if strings.Contains(line, `"type"`) {
			currentTask.Type = TaskType(extractJSONValue(line, "type"))
		}
	}

	if currentTask.Title != "" {
		tasks = append(tasks, currentTask)
	}

	// Ensure we have at least some tasks
	if len(tasks) == 0 {
		tasks = append(tasks, &Task{
			ID:          uuid.New().String(),
			GoalID:      goalID,
			Title:       "Complete task",
			Description: "Work on the goal",
			Type:        TaskTypeImplementation,
			Priority:    1,
		})
	}

	return tasks, nil
}

// estimateTimeline estimates the timeline for completing all goals and tasks
func (d *Decomposer) estimateTimeline(subGoals []*DecomposedGoal, tasks []*Task) string {
	totalHours := 0

	for _, goal := range subGoals {
		if goal.EstimatedDuration != "" {
			// Parse duration string like "2-3 hours"
			hours := parseDuration(goal.EstimatedDuration)
			totalHours += hours
		} else {
			// Default based on complexity
			switch goal.Complexity {
			case ComplexityLow:
				totalHours += 1
			case ComplexityMedium:
				totalHours += 2
			case ComplexityHigh:
				totalHours += 4
			}
		}
	}

	if totalHours <= 4 {
		return "Half day"
	} else if totalHours <= 8 {
		return "1 day"
	} else if totalHours <= 16 {
		return "2 days"
	} else if totalHours <= 24 {
		return "3 days"
	} else {
		return "1 week"
	}
}

// identifyResources identifies resources needed for the goals
func (d *Decomposer) identifyResources(subGoals []*DecomposedGoal, tasks []*Task) []string {
	resourceSet := make(map[string]bool)

	for _, goal := range subGoals {
		for _, skill := range goal.Skills {
			resourceSet[skill] = true
		}
	}

	// Add common resources based on task types
	for _, task := range tasks {
		switch task.Type {
		case TaskTypeImplementation:
			resourceSet["code editor"] = true
			resourceSet["git"] = true
		case TaskTypeTesting:
			resourceSet["testing framework"] = true
			resourceSet["test environment"] = true
		case TaskTypeDeployment:
			resourceSet["deployment pipeline"] = true
			resourceSet["CI/CD"] = true
		case TaskTypeResearch:
			resourceSet["documentation"] = true
			resourceSet["internet access"] = true
		}
	}

	resources := make([]string, 0, len(resourceSet))
	for r := range resourceSet {
		resources = append(resources, r)
	}

	return resources
}

// identifyRisks identifies potential risks
func (d *Decomposer) identifyRisks(subGoals []*DecomposedGoal) []string {
	risks := []string{}

	highComplexity := 0
	for _, goal := range subGoals {
		if goal.Complexity == ComplexityHigh {
			highComplexity++
		}
	}

	if highComplexity > 2 {
		risks = append(risks, "Multiple high-complexity sub-goals may cause delays")
	}

	if len(subGoals) > 5 {
		risks = append(risks, "Large number of sub-goals increases coordination overhead")
	}

	// Check for potential circular dependencies
	hasCircularDeps := false
	for _, goal := range subGoals {
		for _, dep := range goal.Dependencies {
			if dep == goal.ID {
				hasCircularDeps = true
				break
			}
		}
	}
	if hasCircularDeps {
		risks = append(risks, "Circular dependencies detected - may cause deadlocks")
	}

	return risks
}

// calculateComplexity determines overall complexity
func (d *Decomposer) calculateComplexity(subGoals []*DecomposedGoal) ComplexityLevel {
	if len(subGoals) == 0 {
		return ComplexityLow
	}

	highCount := 0
	for _, goal := range subGoals {
		if goal.Complexity == ComplexityHigh {
			highCount++
		}
	}

	if highCount > len(subGoals)/2 {
		return ComplexityHigh
	}

	for _, goal := range subGoals {
		if goal.Complexity == ComplexityHigh {
			return ComplexityMedium
		}
	}

	return ComplexityLow
}

// Helper functions

func extractJSONValue(line, key string) string {
	// Simple JSON value extraction
	search := fmt.Sprintf(`"%s"`, key)
	idx := strings.Index(line, search)
	if idx == -1 {
		return ""
	}

	// Find the colon and opening quote
	start := idx + len(search) + 1
	for start < len(line) && (line[start] == ' ' || line[start] == ':' || line[start] == '"') {
		start++
	}

	// Find closing quote
	end := start
	for end < len(line) && line[end] != '"' && line[end] != ',' && line[end] != '}' {
		end++
	}

	return strings.TrimSpace(line[start:end])
}

func extractJSONArray(line, key string) []string {
	// Simple array extraction
	if !strings.Contains(line, key) {
		return nil
	}

	// Find array start
	start := strings.Index(line, "[")
	end := strings.Index(line, "]")
	if start == -1 || end == -1 {
		return nil
	}

	arrayStr := line[start+1 : end]
	items := strings.Split(arrayStr, ",")

	result := make([]string, 0)
	for _, item := range items {
		item = strings.TrimSpace(item)
		item = strings.Trim(item, `" `)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}

func parseComplexity(s string) ComplexityLevel {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "high":
		return ComplexityHigh
	case "medium":
		return ComplexityMedium
	default:
		return ComplexityLow
	}
}

func parseDuration(s string) int {
	s = strings.ToLower(s)

	// Handle ranges like "2-3 hours"
	parts := strings.Split(s, "-")
	if len(parts) == 2 {
		return (parseDuration(parts[0]) + parseDuration(parts[1])) / 2
	}

	// Extract number - om strängen inte innehåller ett tal blir num 0.
	var num int
	_, _ = fmt.Sscanf(s, "%d", &num)

	if strings.Contains(s, "hour") || strings.Contains(s, "h") {
		return num
	}
	if strings.Contains(s, "day") || strings.Contains(s, "d") {
		return num * 8
	}
	if strings.Contains(s, "week") || strings.Contains(s, "w") {
		return num * 40
	}

	return num
}

// DependencyGraph represents the dependency structure
type DependencyGraph struct {
	Nodes map[string]*GoalNode
	Edges []DependencyEdge
}

// GoalNode represents a node in the dependency graph
type GoalNode struct {
	Goal      *DecomposedGoal
	InDegree  int
	OutDegree int
}

// DependencyEdge represents an edge in the dependency graph
type DependencyEdge struct {
	From string
	To   string
}

// BuildGraph builds a dependency graph from goals
func BuildGraph(goals []*DecomposedGoal) *DependencyGraph {
	graph := &DependencyGraph{
		Nodes: make(map[string]*GoalNode),
		Edges: make([]DependencyEdge, 0),
	}

	for _, goal := range goals {
		graph.Nodes[goal.ID] = &GoalNode{Goal: goal}
	}

	for _, goal := range goals {
		for _, depID := range goal.Dependencies {
			if _, exists := graph.Nodes[depID]; exists {
				graph.Edges = append(graph.Edges, DependencyEdge{
					From: depID,
					To:   goal.ID,
				})
				graph.Nodes[goal.ID].InDegree++
				graph.Nodes[depID].OutDegree++
			}
		}
	}

	return graph
}

// TopologicalSort returns goals in execution order
func (g *DependencyGraph) TopologicalSort() []*DecomposedGoal {
	var result []*DecomposedGoal
	visited := make(map[string]bool)

	var visit func(id string)
	visit = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true

		// Visit dependencies first
		for _, edge := range g.Edges {
			if edge.To == id {
				visit(edge.From)
			}
		}

		result = append(result, g.Nodes[id].Goal)
	}

	for id := range g.Nodes {
		visit(id)
	}

	return result
}

// FindCriticalPath finds the longest path through the dependency graph
func (g *DependencyGraph) FindCriticalPath() []string {
	// Simplified: just return nodes sorted by priority
	var path []string
	for _, node := range g.Nodes {
		path = append(path, node.Goal.ID)
	}
	return path
}

// ExecutionPlan represents a planned execution
type ExecutionPlan struct {
	ID         string
	Goal       *DecomposedGoal
	SubGoals   []*DecomposedGoal
	Tasks      []*Task
	Milestones []Milestone
	StartTime  time.Time
	EndTime    time.Time
	Status     PlanStatus
}

// PlanStatus represents the status of an execution plan
type PlanStatus string

const (
	PlanStatusDraft      PlanStatus = "draft"
	PlanStatusApproved   PlanStatus = "approved"
	PlanStatusInProgress PlanStatus = "in_progress"
	PlanStatusCompleted  PlanStatus = "completed"
	PlanStatusCancelled  PlanStatus = "cancelled"
)

// Milestone represents a checkpoint in the execution
type Milestone struct {
	ID          string
	Title       string
	GoalID      string
	Tasks       []string // Task IDs
	Completed   bool
	CompletedAt *time.Time
}

// CreateExecutionPlan creates an execution plan from decomposition
func (d *Decomposer) CreateExecutionPlan(result *DecompositionResult) *ExecutionPlan {
	plan := &ExecutionPlan{
		ID:         uuid.New().String(),
		Goal:       result.Goal,
		SubGoals:   result.SubGoals,
		Tasks:      result.Tasks,
		Status:     PlanStatusDraft,
		Milestones: make([]Milestone, 0),
	}

	// Create milestones for each sub-goal
	for i, goal := range result.SubGoals {
		var taskIDs []string
		for _, task := range result.Tasks {
			if task.GoalID == goal.ID {
				taskIDs = append(taskIDs, task.ID)
			}
		}

		plan.Milestones = append(plan.Milestones, Milestone{
			ID:     uuid.New().String(),
			Title:  fmt.Sprintf("Milestone %d: %s", i+1, goal.Title),
			GoalID: goal.ID,
			Tasks:  taskIDs,
		})
	}

	return plan
}

// PrettyPrint prints the decomposition result in a readable format
func (r *DecompositionResult) PrettyPrint() string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString("                    GOAL DECOMPOSITION                             \n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	_, _ = fmt.Fprintf(&sb, "Main Goal: %s\n\n", r.Goal.Title)

	_, _ = fmt.Fprintf(&sb, "Timeline: %s\n", r.Timeline)
	_, _ = fmt.Fprintf(&sb, "Sub-Goals: %d\n", len(r.SubGoals))
	_, _ = fmt.Fprintf(&sb, "Total Tasks: %d\n\n", len(r.Tasks))

	if len(r.Resources) > 0 {
		sb.WriteString("───────────────────────────────────────────────────────────────\n")
		sb.WriteString("Resources Needed\n")
		sb.WriteString("───────────────────────────────────────────────────────────────\n")
		for _, res := range r.Resources {
			_, _ = fmt.Fprintf(&sb, "  • %s\n", res)
		}
		sb.WriteString("\n")
	}

	if len(r.Risks) > 0 {
		sb.WriteString("───────────────────────────────────────────────────────────────\n")
		sb.WriteString("Potential Risks\n")
		sb.WriteString("───────────────────────────────────────────────────────────────\n")
		for _, risk := range r.Risks {
			_, _ = fmt.Fprintf(&sb, "  ⚠ %s\n", risk)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("───────────────────────────────────────────────────────────────\n")
	sb.WriteString("Sub-Goals\n")
	sb.WriteString("───────────────────────────────────────────────────────────────\n")

	for i, goal := range r.SubGoals {
		_, _ = fmt.Fprintf(&sb, "\n[%d] %s (%s)\n", i+1, goal.Title, goal.Complexity)
		_, _ = fmt.Fprintf(&sb, "    %s\n", goal.Description)

		// Find tasks for this goal
		for _, task := range r.Tasks {
			if task.GoalID == goal.ID {
				_, _ = fmt.Fprintf(&sb, "    ├── [%s] %s\n", strings.ToUpper(string(task.Type)[:4]), task.Title)
			}
		}
	}

	sb.WriteString("\n═══════════════════════════════════════════════════════════════\n")

	return sb.String()
}
