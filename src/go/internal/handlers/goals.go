package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"plan-api/internal/models"
	"plan-api/internal/recurrence"
	"plan-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type GoalHandler struct {
	repo        *repository.GoalRepository
	places      *repository.PlaceRepository
	dbConnected bool
}

func NewGoalHandler(dbConnected bool) *GoalHandler {
	return &GoalHandler{
		repo:        repository.NewGoalRepository(),
		places:      repository.NewPlaceRepository(),
		dbConnected: dbConnected,
	}
}

// attachPlaces fyller i målets platser.
//
// Ett fel här ska inte förstöra svaret - platserna är extra information,
// så ett fel hoppas över i stället för att returneras.
func (h *GoalHandler) attachPlaces(ctx context.Context, goals []models.Goal) {
	if len(goals) == 0 {
		return
	}

	ids := make([]int64, len(goals))
	for i, g := range goals {
		ids[i] = g.ID
	}

	byGoal, err := h.places.GetPlacesForGoals(ctx, ids)
	if err != nil {
		return
	}

	for i := range goals {
		goals[i].Places = placesOrEmpty(byGoal[goals[i].ID])
	}
}

// placesOrEmpty gör att en tom lista alltid skickas ut som [] och inte null.
func placesOrEmpty(places []models.Place) []models.Place {
	if places == nil {
		return []models.Place{}
	}
	return places
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
	h.attachPlaces(c.Request.Context(), goals)
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

	// Hela grafen hämtas med fyra frågor i stället för en fråga per mål och
	// relation. Det är den skillnad som gör målsidan snabb: en målsida med
	// många undermål kostade tidigare 30+ rundturer till databasen.
	graph, err := h.loadGraph(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	goal, ok := graph.goal(goalID)
	if !ok {
		// Målet tillhör inte användaren - det finns inte att hämta.
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Goal not found"})
		return
	}

	children := graph.children(goalID)
	parentGoals := graph.parents(goalID)
	dependencies := graph.dependenciesOf(goalID)
	dependsOn := graph.dependsOnOf(goalID)
	blocking := graph.blockingOf(goalID)

	c.JSON(http.StatusOK, models.GetGoalResponse{
		Goal: models.GoalWithRelations{
			Goal:            goal,
			ChildRelations:  graph.childRelations(goalID),
			ParentRelations: parentGoals,
		},
		Children:     children,
		Parents:      parentGoals,
		AllGoals:     graph.goals,
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

	// Upprepningen valideras innan målet skapas, så att en ogiltig regel
	// inte lämnar kvar ett tomt mål.
	recurrenceRule := ""
	recurrenceMode := recurrence.ModeSchedule
	recurrenceStart := time.Now().UTC()

	if req.RecurrenceRule != nil {
		recurrenceRule = strings.TrimSpace(*req.RecurrenceRule)

		if err := recurrence.Validate(recurrenceRule); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: recurrenceErrorMessage(err)})
			return
		}

		if req.RecurrenceMode != nil && *req.RecurrenceMode != "" {
			mode := recurrence.Mode(*req.RecurrenceMode)
			if err := recurrence.ValidateMode(mode); err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: recurrenceErrorMessage(err)})
				return
			}
			recurrenceMode = mode
		}

		if req.StartDate != nil && *req.StartDate != "" {
			parsed, err := time.Parse(models.DueDateFormat, *req.StartDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "start_date måste vara på formatet YYYY-MM-DD"})
				return
			}
			recurrenceStart = parsed
		}

		if _, err := recurrence.FirstOccurrence(recurrenceRule, recurrenceStart); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: recurrenceErrorMessage(err)})
			return
		}
	}

	goal, err := h.repo.CreateGoal(c.Request.Context(), userID, req.Title, icon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	if recurrenceRule != "" {
		goal, err = h.repo.SetRecurrence(c.Request.Context(), goal.ID, recurrenceRule, recurrenceMode, recurrenceStart)
		if err != nil {
			// Målet skapades men fick ingen upprepning. Städa bort det
			// igen, så att inget tomt mål blir kvar i listan.
			if cleanupErr := h.repo.DeleteGoal(c.Request.Context(), goal.ID); cleanupErr != nil {
				fmt.Printf("failed to clean up goal %d after recurrence error: %v\n", goal.ID, cleanupErr)
			}
			c.JSON(recurrenceErrorStatus(err), models.ErrorResponse{Message: recurrenceErrorMessage(err)})
			return
		}
	}

	c.JSON(http.StatusCreated, goal)
}

// recurrenceErrorStatus översätter ett upprepningsfel till en HTTP-status.
func recurrenceErrorStatus(err error) int {
	switch {
	case errors.Is(err, recurrence.ErrInvalidRule),
		errors.Is(err, recurrence.ErrInvalidMode),
		errors.Is(err, recurrence.ErrNoOccurrence):
		return http.StatusBadRequest
	case errors.Is(err, repository.ErrNotRecurring):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// recurrenceErrorMessage ger ett läsbart felmeddelande för ett upprepningsfel.
func recurrenceErrorMessage(err error) string {
	switch {
	case errors.Is(err, recurrence.ErrInvalidRule):
		return err.Error()
	case errors.Is(err, recurrence.ErrInvalidMode):
		return "Ogiltig schemaläggningsstrategi. Använd 'schedule' eller 'push'."
	case errors.Is(err, recurrence.ErrNoOccurrence):
		return "Regeln ger inga förekomster."
	case errors.Is(err, repository.ErrNotRecurring):
		return "Målet är inte återkommande."
	default:
		return err.Error()
	}
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

	// Ett evigt mål kan inte slutföras - det pågår hela livet.
	if req.FinishesGoal() {
		existing, err := h.repo.GetGoalByID(c.Request.Context(), goalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		if existing != nil && existing.Eternal {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Ett evigt mål kan inte slutföras"})
			return
		}
	}

	goal, err := h.repo.UpdateGoal(c.Request.Context(), goalID, req)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidTimestamp) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
		if errors.Is(err, repository.ErrInvalidDueDate) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "due-datumet måste vara på formatet YYYY-MM-DD"})
			return
		}
		if errors.Is(err, repository.ErrInvalidEstimate) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "tidsåtgången måste vara ett positivt antal minuter"})
			return
		}
		if errors.Is(err, recurrence.ErrInvalidRule) || errors.Is(err, recurrence.ErrInvalidMode) ||
			errors.Is(err, recurrence.ErrNoOccurrence) || errors.Is(err, repository.ErrNotRecurring) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: recurrenceErrorMessage(err)})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, goal)
}

// CompleteGoalOccurrence godoc
// @Summary Complete the current occurrence of a recurring goal
// @Description Marks the current occurrence of a recurring goal as completed and moves on to the next one
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Param completion body models.CompleteOccurrenceRequest true "Completion data"
// @Success 200 {object} models.Goal
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/goals/{id}/completions [post]
func (h *GoalHandler) CompleteGoalOccurrence(c *gin.Context) {
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

	var req models.CompleteOccurrenceRequest
	_ = c.ShouldBindJSON(&req)

	// Förekomsten som slutförs. Utan uppgift slutförs målets aktuella
	// förekomst.
	var occurrenceDate *time.Time
	if req.OccurrenceDate != "" {
		parsed, err := time.Parse(models.DueDateFormat, req.OccurrenceDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "occurrence_date måste vara på formatet YYYY-MM-DD"})
			return
		}
		occurrenceDate = &parsed
	}

	completedOn := time.Now().UTC()
	if req.CompletedOn != "" {
		parsed, err := time.Parse(models.DueDateFormat, req.CompletedOn)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "completed_on måste vara på formatet YYYY-MM-DD"})
			return
		}
		completedOn = parsed
	}

	goal, err := h.repo.CompleteOccurrence(c.Request.Context(), goalID, occurrenceDate, completedOn)
	if err != nil {
		if errors.Is(err, repository.ErrNotRecurring) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Målet är inte återkommande"})
			return
		}
		if errors.Is(err, recurrence.ErrNoOccurrence) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Målet har ingen aktiv förekomst"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if goal == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Goal not found"})
		return
	}

	c.JSON(http.StatusOK, goal)
}

// UndoGoalOccurrence godoc
// @Summary Undo the last completed occurrence of a recurring goal
// @Description Removes the latest completion of a recurring goal and restores that occurrence
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Success 200 {object} models.Goal
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/goals/{id}/completions [delete]
func (h *GoalHandler) UndoGoalOccurrence(c *gin.Context) {
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

	goal, err := h.repo.UndoOccurrence(c.Request.Context(), goalID)
	if err != nil {
		if errors.Is(err, repository.ErrNotRecurring) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Målet är inte återkommande"})
			return
		}
		if errors.Is(err, repository.ErrNoCompletion) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Målet har ingen slutförd förekomst"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if goal == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Goal not found"})
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

	// Prioriterade mål är en lista av mål - de platser som hör till dem
	// hämtas i ett svep.
	plain := make([]models.Goal, len(goals))
	for i, g := range goals {
		plain[i] = g.Goal
	}
	h.attachPlaces(c.Request.Context(), plain)
	for i := range goals {
		goals[i].Places = plain[i].Places
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

// ReorderRelations godoc
// @Summary Reorder the children of a goal
// @Description Sets a new order for all children of a goal in one transaction
// @Tags relations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param relation body models.ReorderRelationsRequest true "New order of the children"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/goals/relations/reorder [patch]
func (h *GoalHandler) ReorderRelations(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.ReorderRelationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "parentId and childIds are required"})
		return
	}

	if len(req.ChildIDs) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "childIds must contain at least one child"})
		return
	}

	hasAccess, _ := h.repo.UserHasAccess(c.Request.Context(), userID, req.ParentID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	if err := h.repo.ReorderRelations(c.Request.Context(), req.ParentID, req.ChildIDs); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": true})
}

// ReorderGoals godoc
// @Summary Reorder the root goals
// @Description Sets a new order for the root goals (goals without a parent) of the current user
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param goals body models.ReorderGoalsRequest true "New order of the root goals"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/goals/reorder [patch]
func (h *GoalHandler) ReorderGoals(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.ReorderGoalsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "goalIds is required"})
		return
	}

	if len(req.GoalIDs) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "goalIds must contain at least one goal"})
		return
	}

	if err := h.repo.ReorderRootGoals(c.Request.Context(), userID, req.GoalIDs); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": true})
}

// ReorderPriorityGoals godoc
// @Summary Reorder the priority list
// @Description Sets a new order for the active goals in the priority list
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param goals body models.ReorderPriorityGoalsRequest true "New priority order"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/goals/priority/reorder [patch]
func (h *GoalHandler) ReorderPriorityGoals(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.ReorderPriorityGoalsRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.GoalIDs) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "goalIds must contain at least one goal"})
		return
	}

	if err := h.repo.ReorderPriorityGoals(c.Request.Context(), userID, req.GoalIDs); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": true})
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
