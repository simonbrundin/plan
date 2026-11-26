package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Goal struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Status      string `json:"status"`
	ParentID    string `json:"parent_id,omitempty"`
	Priority    int    `json:"priority,omitempty"`
	Created     string `json:"created"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type GoalStatus struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// UpdateGoalRequest is the request body for updating a goal
type UpdateGoalRequest struct {
	Status      string `json:"status,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// StartAgentRequest is the request body for starting an agent on a goal
type StartAgentRequest struct {
	GoalID       string `json:"goal_id"`
	Instructions string `json:"instructions,omitempty"`
}

type GoalsResponse struct {
	Goals    []Goal       `json:"goals"`
	Statuses []GoalStatus `json:"statuses"`
}

var (
	goalsCache     *GoalsResponse
	goalsCacheOnce sync.Once
	goalsCacheErr  error
)

// getGoalsPath returns the path to goals.json
func getGoalsPath() string {
	// Check for custom path first
	if path := os.Getenv("GOALS_FILE"); path != "" {
		return path
	}
	// Default to goals.json in the project root
	return filepath.Join(getProjectRoot(), "goals.json")
}

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Try environment variable first
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		return root
	}
	// Fallback: look in parent directories for goals.json
	// The binary is in src/agent/ so we need to go up
	cwd, _ := os.Getwd()

	// Check if goals.json exists in current dir
	if _, err := os.Stat(filepath.Join(cwd, "goals.json")); err == nil {
		return cwd
	}

	// Check parent directory (for when running from src/agent)
	if _, err := os.Stat(filepath.Join(cwd, "..", "goals.json")); err == nil {
		return filepath.Join(cwd, "..")
	}

	return cwd
}

// loadGoals loads goals from the JSON file
func loadGoals() (*GoalsResponse, error) {
	goalsCacheOnce.Do(func() {
		path := getGoalsPath()

		data, err := os.ReadFile(path)
		if err != nil {
			goalsCacheErr = err
			return
		}

		var resp GoalsResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			goalsCacheErr = err
			return
		}

		goalsCache = &resp
	})

	return goalsCache, goalsCacheErr
}

// loadGoalsFromFile loads goals directly from disk (bypasses cache)
func loadGoalsFromFile() (*GoalsResponse, error) {
	path := getGoalsPath()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read goals file: %w", err)
	}

	var resp GoalsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse goals file: %w", err)
	}

	return &resp, nil
}

// invalidateGoalsCache clears the goals cache so the next load reads fresh data
func invalidateGoalsCache() {
	goalsCacheOnce = sync.Once{} // Reset the once
	goalsCache = nil
	goalsCacheErr = nil
}

// saveGoals writes goals to the JSON file and invalidates cache
func saveGoals(goals []Goal) error {
	path := getGoalsPath()

	// Read existing file to get statuses
	existing, err := loadGoalsFromFile()
	if err != nil {
		return fmt.Errorf("failed to load existing goals: %w", err)
	}

	// Merge: update existing goals, keep statuses
	saveData := map[string]interface{}{
		"goals":    goals,
		"statuses": existing.Statuses,
	}

	data, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal goals: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write goals file: %w", err)
	}

	// Invalidate cache so next read gets fresh data
	invalidateGoalsCache()
	log.Printf("Goals saved to %s", path)
	return nil
}

// ListGoals returns all goals
// @Summary List all goals
// @Description Returns all goals with their statuses from the goals.json file
// @Tags goals
// @Produce json
// @Success 200 {object} GoalsResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/goals [get]
func ListGoals(c *gin.Context) {
	goals, err := loadGoals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load goals",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, goals)
}

// GetGoal returns a specific goal by ID
// @Summary Get a goal by ID
// @Description Returns a specific goal with all its children
// @Tags goals
// @Produce json
// @Param id path string true "Goal ID"
// @Success 200 {object} Goal
// @Failure 404 {object} map[string]string
// @Router /api/v1/goals/{id} [get]
func GetGoal(c *gin.Context) {
	id := c.Param("id")

	goals, err := loadGoals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load goals",
			"message": err.Error(),
		})
		return
	}

	for _, goal := range goals.Goals {
		if goal.ID == id {
			// Get children
			var children []Goal
			for _, child := range goals.Goals {
				if child.ParentID == id {
					children = append(children, child)
				}
			}

			goalMap := map[string]interface{}{
				"goal":     goal,
				"children": children,
			}
			c.JSON(http.StatusOK, goalMap)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error":   "Goal not found",
		"message": "No goal found with ID: " + id,
	})
}

// GetGoalStats returns statistics about goals
// @Summary Get goal statistics
// @Description Returns statistics about all goals
// @Tags goals
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/goals/stats [get]
func GetGoalStats(c *gin.Context) {
	goals, err := loadGoals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load goals",
			"message": err.Error(),
		})
		return
	}

	stats := make(map[string]interface{})

	// Count by status
	statusCounts := make(map[string]int)
	for _, goal := range goals.Goals {
		statusCounts[goal.Status]++
	}
	stats["byStatus"] = statusCounts
	stats["total"] = len(goals.Goals)

	// Count completed
	completed := 0
	for _, goal := range goals.Goals {
		if goal.Status == "completed" {
			completed++
		}
	}
	stats["completed"] = completed

	if len(goals.Goals) > 0 {
		stats["completionRate"] = float64(completed) / float64(len(goals.Goals)) * 100
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateGoal updates a goal's status
// @Summary Update a goal
// @Description Updates a goal's status, title, or description
// @Tags goals
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Param goal body UpdateGoalRequest true "Goal update"
// @Success 200 {object} Goal
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/goals/{id} [patch]
func UpdateGoal(c *gin.Context) {
	id := c.Param("id")

	var req UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Load goals from file
	goals, err := loadGoalsFromFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load goals",
			"message": err.Error(),
		})
		return
	}

	// Find and update the goal
	var updatedGoal *Goal
	for i := range goals.Goals {
		if goals.Goals[i].ID == id {
			if req.Status != "" {
				goals.Goals[i].Status = req.Status
				// Set completed_at if marking as completed
				if req.Status == "completed" && goals.Goals[i].CompletedAt == "" {
					goals.Goals[i].CompletedAt = time.Now().UTC().Format(time.RFC3339)
				}
			}
			if req.Title != "" {
				goals.Goals[i].Title = req.Title
			}
			if req.Description != "" {
				goals.Goals[i].Description = req.Description
			}
			updatedGoal = &goals.Goals[i]
			break
		}
	}

	if updatedGoal == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Goal not found",
			"message": "No goal found with ID: " + id,
		})
		return
	}

	// Save to file
	if err := saveGoals(goals.Goals); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save goals",
			"message": err.Error(),
		})
		return
	}

	log.Printf("Goal %s updated: status=%s", id, req.Status)
	c.JSON(http.StatusOK, updatedGoal)
}

// StartAgentForGoal starts an agent to work on a goal
// @Summary Start agent for a goal
// @Description Starts an agent to work on a specific goal with optional instructions
// @Tags goals
// @Accept json
// @Produce json
// @Param goal body StartAgentRequest true "Start agent request"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/goals/start-agent [post]
func StartAgentForGoal(c *gin.Context) {
	var req StartAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	if req.GoalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "goal_id is required",
			"message": "You must specify a goal_id",
		})
		return
	}

	// Load goals to find the target goal
	goals, err := loadGoalsFromFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to load goals",
			"message": err.Error(),
		})
		return
	}

	// Find the goal
	var targetGoal *Goal
	for i := range goals.Goals {
		if goals.Goals[i].ID == req.GoalID {
			targetGoal = &goals.Goals[i]
			break
		}
	}

	if targetGoal == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Goal not found",
			"message": "No goal found with ID: " + req.GoalID,
		})
		return
	}

	// Update goal status to 'execute'
	targetGoal.Status = "execute"

	// Find and update in goals list
	for i := range goals.Goals {
		if goals.Goals[i].ID == req.GoalID {
			goals.Goals[i] = *targetGoal
			break
		}
	}

	// Save updated status
	if err := saveGoals(goals.Goals); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update goal status",
			"message": err.Error(),
		})
		return
	}

	// Start agent in background (non-blocking)
	go startAgentOnGoal(targetGoal, req.Instructions)

	log.Printf("Agent started for goal: %s (%s)", targetGoal.Title, targetGoal.ID)
	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Agent started",
		"goal_id":    targetGoal.ID,
		"goal_title": targetGoal.Title,
		"status":     "execute",
	})
}

// startAgentOnGoal runs the agent to work on a goal (runs in background)
func startAgentOnGoal(goal *Goal, instructions string) {
	log.Printf("Background agent processing goal: %s", goal.Title)

	// TODO: Integrate with actual agent runtime
	// For now, this is a placeholder that demonstrates the architecture
	// The full implementation would:
	// 1. Create an AgentRuntime with goal-specific tools
	// 2. Build a system prompt from the goal description
	// 3. Run the agent loop with checkpoint-gate workflow
	// 4. Update goal status based on completion

	log.Printf("Agent completed goal: %s", goal.Title)
}
