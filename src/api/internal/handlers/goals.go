package handlers

import (
	"net/http"
	"strconv"

	"plan-api/internal/models"
	"plan-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type GoalHandler struct {
	repo        *repository.GoalRepository
	dbConnected bool
}

func NewGoalHandler(dbConnected bool) *GoalHandler {
	return &GoalHandler{
		repo:        repository.NewGoalRepository(),
		dbConnected: dbConnected,
	}
}

// checkDb returns error response if database is not connected
func (h *GoalHandler) checkDb(c *gin.Context) bool {
	if !h.dbConnected {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Message: "Database not connected. Set DATABASE_URL environment variable.",
		})
		return false
	}
	return true
}

// GetGoals godoc
// @Summary List all goals
// @Description Get all goals for the authenticated user
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Goal
// @Failure 401 {object} models.ErrorResponse
// @Router /api/goals [get]
func (h *GoalHandler) GetGoals(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	goals, err := h.repo.GetAllGoals(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	if goals == nil {
		goals = []models.Goal{}
	}
	c.JSON(http.StatusOK, goals)
}

// GetGoal godoc
// @Summary Get a goal by ID
// @Description Get a single goal with all its relations and dependencies
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Success 200 {object} models.GetGoalResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/goals/{id} [get]
func (h *GoalHandler) GetGoal(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid goal ID"})
		return
	}

	hasAccess, err := h.repo.UserHasAccess(c.Request.Context(), userID, goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	goal, err := h.repo.GetGoalByID(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if goal == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Goal not found"})
		return
	}

	allGoals, _ := h.repo.GetAllGoals(c.Request.Context(), userID)
	if allGoals == nil {
		allGoals = []models.Goal{}
	}

	childRelations, _ := h.repo.GetChildRelations(c.Request.Context(), goalID)
	var children []models.ChildWithDeps
	for _, cr := range childRelations {
		childGoal, _ := h.repo.GetGoalByID(c.Request.Context(), cr.ChildID)
		if childGoal != nil {
			dependsOn, _ := h.repo.GetDependencies(c.Request.Context(), cr.ChildID)
			var depGoals []models.Goal
			for _, d := range dependsOn {
				if g, _ := h.repo.GetGoalByID(c.Request.Context(), d.DependsOnID); g != nil {
					depGoals = append(depGoals, *g)
				}
			}
			blocking, _ := h.repo.GetBlockingGoals(c.Request.Context(), cr.ChildID)
			children = append(children, models.ChildWithDeps{
				Goal:      *childGoal,
				Order:     cr.Order,
				Weight:    cr.Weight,
				DependsOn: depGoals,
				Blocking:  blocking,
			})
		}
	}

	parentGoals, _ := h.repo.GetParentRelations(c.Request.Context(), goalID)
	if parentGoals == nil {
		parentGoals = []models.Goal{}
	}
	dependencies, _ := h.repo.GetDependencies(c.Request.Context(), goalID)
	var dependsOn []models.Goal
	for _, d := range dependencies {
		if g, _ := h.repo.GetGoalByID(c.Request.Context(), d.DependsOnID); g != nil {
			dependsOn = append(dependsOn, *g)
		}
	}
	blocking, _ := h.repo.GetBlockingGoals(c.Request.Context(), goalID)

	c.JSON(http.StatusOK, models.GetGoalResponse{
		Goal: models.GoalWithRelations{
			Goal:            *goal,
			ChildRelations:  childRelations,
			ParentRelations: parentGoals,
		},
		Children:     children,
		Parents:      parentGoals,
		AllGoals:     allGoals,
		Dependencies: dependencies,
		DependsOn:    dependsOn,
		Blocking:     blocking,
	})
}

// CreateGoal godoc
// @Summary Create a new goal
// @Description Create a new goal for the authenticated user
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param goal body models.CreateGoalRequest true "Goal data"
// @Success 201 {object} models.Goal
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/goals [post]
func (h *GoalHandler) CreateGoal(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Title is required"})
		return
	}

	icon := req.Icon
	if icon == "" {
		icon = "heroicons:star"
	}

	goal, err := h.repo.CreateGoal(c.Request.Context(), userID, req.Title, icon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, goal)
}

// UpdateGoal godoc
// @Summary Update a goal
// @Description Update a goal's title, icon, started, or finished status
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Param goal body models.UpdateGoalRequest true "Update data"
// @Success 200 {object} models.Goal
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/goals/{id} [patch]
func (h *GoalHandler) UpdateGoal(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid goal ID"})
		return
	}

	hasAccess, _ := h.repo.UserHasAccess(c.Request.Context(), userID, goalID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	var req models.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "No fields to update"})
		return
	}

	goal, err := h.repo.UpdateGoal(c.Request.Context(), goalID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, goal)
}

// DeleteGoal godoc
// @Summary Delete a goal
// @Description Delete a goal and all its relations
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/goals/{id} [delete]
func (h *GoalHandler) DeleteGoal(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid goal ID"})
		return
	}

	hasAccess, _ := h.repo.UserHasAccess(c.Request.Context(), userID, goalID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	if err := h.repo.DeleteGoal(c.Request.Context(), goalID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "deleted": goalID})
}

// GetPrioritizedGoals godoc
// @Summary Get prioritized goals
// @Description Get all active goals sorted by priority
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.PrioritizedGoal
// @Failure 401 {object} models.ErrorResponse
// @Router /api/goals/prioritized [get]
func (h *GoalHandler) GetPrioritizedGoals(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	goals, err := h.repo.GetPrioritizedGoals(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	if goals == nil {
		goals = []models.PrioritizedGoal{}
	}
	c.JSON(http.StatusOK, goals)
}

// CreateRelation godoc
// @Summary Create a parent-child relation
// @Description Create a relation between two goals
// @Tags relations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param relation body models.CreateRelationRequest true "Relation data"
// @Success 200 {object} models.ChildRelation
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/goals/relations [post]
func (h *GoalHandler) CreateRelation(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.CreateRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "childId and parentId are required"})
		return
	}

	hasAccess1, _ := h.repo.UserHasAccess(c.Request.Context(), userID, req.ChildID)
	hasAccess2, _ := h.repo.UserHasAccess(c.Request.Context(), userID, req.ParentID)
	if !hasAccess1 || !hasAccess2 {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	rel, err := h.repo.CreateRelation(c.Request.Context(), req.ParentID, req.ChildID, req.Order, req.Weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, rel)
}

// UpdateRelation godoc
// @Summary Update a relation
// @Description Update a relation's order and weight
// @Tags relations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param relation body models.UpdateRelationRequest true "Update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/goals/relations [patch]
func (h *GoalHandler) UpdateRelation(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.UpdateRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "childId and parentId are required"})
		return
	}

	hasAccess, _ := h.repo.UserHasAccess(c.Request.Context(), userID, req.ParentID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	if err := h.repo.UpdateRelation(c.Request.Context(), req.ParentID, req.ChildID, req.Order, req.Weight); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": true})
}

// DeleteRelation godoc
// @Summary Delete a relation
// @Description Delete a parent-child relation
// @Tags relations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param relation body models.DeleteRelationRequest true "Relation data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/goals/relations [delete]
func (h *GoalHandler) DeleteRelation(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.DeleteRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "childId and parentId are required"})
		return
	}

	hasAccess, _ := h.repo.UserHasAccess(c.Request.Context(), userID, req.ParentID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	if err := h.repo.DeleteRelation(c.Request.Context(), req.ParentID, req.ChildID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// CreateDependency godoc
// @Summary Create a dependency
// @Description Create a dependency between two goals
// @Tags dependencies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param dependency body models.CreateDependencyRequest true "Dependency data"
// @Success 200 {object} models.GoalDependency
// @Failure 400 {object} models.ErrorResponse "Invalid request or cycle detected"
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/goals/dependencies [post]
func (h *GoalHandler) CreateDependency(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.CreateDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "goalId and dependsOnId are required"})
		return
	}

	if req.GoalID == req.DependsOnID {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "A goal cannot depend on itself"})
		return
	}

	hasAccess1, _ := h.repo.UserHasAccess(c.Request.Context(), userID, req.GoalID)
	hasAccess2, _ := h.repo.UserHasAccess(c.Request.Context(), userID, req.DependsOnID)
	if !hasAccess1 || !hasAccess2 {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	hasCycle, path, _ := h.repo.CheckForCycle(c.Request.Context(), req.GoalID, req.DependsOnID)
	if hasCycle {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: "Cykel upptäckt: Detta skulle skapa en cykel genom: " + formatPath(path),
		})
		return
	}

	dep, err := h.repo.CreateDependency(c.Request.Context(), req.GoalID, req.DependsOnID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dep)
}

// DeleteDependency godoc
// @Summary Delete a dependency
// @Description Delete a dependency between two goals
// @Tags dependencies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param dependency body models.DeleteDependencyRequest true "Dependency data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/goals/dependencies [delete]
func (h *GoalHandler) DeleteDependency(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.DeleteDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "goalId and dependsOnId are required"})
		return
	}

	hasAccess, _ := h.repo.UserHasAccess(c.Request.Context(), userID, req.GoalID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	if err := h.repo.DeleteDependency(c.Request.Context(), req.GoalID, req.DependsOnID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// GetStatuses godoc
// @Summary List all statuses
// @Description Get all available goal statuses
// @Tags statuses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Status
// @Failure 401 {object} models.ErrorResponse
// @Router /api/statuses [get]
func (h *GoalHandler) GetStatuses(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	statuses, err := h.repo.GetAllStatuses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	if statuses == nil {
		statuses = []models.Status{}
	}
	c.JSON(http.StatusOK, statuses)
}

// GetGoalStatus godoc
// @Summary Get a goal's current status
// @Description Get the current status of a specific goal
// @Tags statuses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Success 200 {object} models.Status
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/goals/{id}/status [get]
func (h *GoalHandler) GetGoalStatus(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid goal ID"})
		return
	}

	hasAccess, _ := h.repo.UserHasAccess(c.Request.Context(), userID, goalID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	status, err := h.repo.GetGoalStatus(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if status == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Goal not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// UpdateGoalStatus godoc
// @Summary Update a goal's status
// @Description Update a goal's status and log the change in history
// @Tags statuses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Param status body models.UpdateStatusRequest true "New status"
// @Success 200 {object} models.StatusUpdate
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/goals/{id}/status [patch]
func (h *GoalHandler) UpdateGoalStatus(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid goal ID"})
		return
	}

	hasAccess, _ := h.repo.UserHasAccess(c.Request.Context(), userID, goalID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	var req models.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "status_id is required"})
		return
	}

	update, err := h.repo.UpdateGoalStatus(c.Request.Context(), goalID, int64(req.StatusID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, update)
}

// GetStatusHistory godoc
// @Summary Get status change history for a goal
// @Description Get all status changes for a specific goal
// @Tags statuses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Success 200 {array} models.StatusUpdate
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/goals/{id}/status/history [get]
func (h *GoalHandler) GetStatusHistory(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	goalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid goal ID"})
		return
	}

	hasAccess, _ := h.repo.UserHasAccess(c.Request.Context(), userID, goalID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	history, err := h.repo.GetStatusHistory(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	if history == nil {
		history = []models.StatusUpdate{}
	}
	c.JSON(http.StatusOK, history)
}

// Helper function to get user ID from context
func getUserID(c *gin.Context) int64 {
	if v, exists := c.Get("userID"); exists {
		return v.(int64)
	}
	return 0
}

// Helper to format cycle path for error message
func formatPath(path []int64) string {
	result := ""
	for i, id := range path {
		if i > 0 {
			result += " → "
		}
		result += strconv.FormatInt(id, 10)
	}
	return result
}
