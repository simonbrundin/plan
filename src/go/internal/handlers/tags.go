package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"plan-api/internal/models"
	"plan-api/internal/repository"

	"github.com/gin-gonic/gin"
)

// TagHandler hanterar taggar och målens tagglistor.
type TagHandler struct {
	repo    *repository.TagRepository
	goals   *repository.GoalRepository
	dbReady bool
}

func NewTagHandler(dbConnected bool) *TagHandler {
	return &TagHandler{
		repo:    repository.NewTagRepository(),
		goals:   repository.NewGoalRepository(),
		dbReady: dbConnected,
	}
}

// checkDb returns error response if database is not connected
func (h *TagHandler) checkDb(c *gin.Context) bool {
	if !h.dbReady {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Message: "Database not connected. Set DATABASE_URL environment variable.",
		})
		return false
	}
	return true
}

// List godoc
// @Summary List all tags
// @Description Get all tags for the authenticated user with the number of goals using them
// @Tags tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Tag
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/tags [get]
func (h *TagHandler) List(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	tags, err := h.repo.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if tags == nil {
		tags = []models.Tag{}
	}

	c.JSON(http.StatusOK, tags)
}

// GetGoalTags godoc
// @Summary List a goal's tags
// @Description Get the tags linked to a goal
// @Tags tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Success 200 {array} string
// @Failure 403 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/tags [get]
func (h *TagHandler) GetGoalTags(c *gin.Context) {
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

	hasAccess, err := h.goals.UserHasAccess(c.Request.Context(), userID, goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	tags, err := h.repo.GetGoalTags(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if tags == nil {
		tags = []string{}
	}

	c.JSON(http.StatusOK, tags)
}

// SetGoalTags godoc
// @Summary Replace a goal's tags
// @Description Replace the tags linked to a goal. Tags that don't exist yet are created.
// @Tags tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Param tags body models.SetGoalTagsRequest true "Tag names"
// @Success 200 {array} string
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/tags [put]
func (h *TagHandler) SetGoalTags(c *gin.Context) {
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

	hasAccess, err := h.goals.UserHasAccess(c.Request.Context(), userID, goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	var req models.SetGoalTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "tags is required"})
		return
	}

	tags, err := h.repo.SetGoalTags(c.Request.Context(), userID, goalID, req.Tags)
	if err != nil {
		if errors.Is(err, models.ErrInvalidTagName) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if tags == nil {
		tags = []string{}
	}

	c.JSON(http.StatusOK, tags)
}

// DeleteGoalTag godoc
// @Summary Remove a tag from a goal
// @Description Remove one tag from a goal. The tag itself is kept for other goals.
// @Tags tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Param name path string true "Tag name"
// @Success 200 {array} string
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/tags/{name} [delete]
func (h *TagHandler) DeleteGoalTag(c *gin.Context) {
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

	hasAccess, err := h.goals.UserHasAccess(c.Request.Context(), userID, goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	tags, err := h.repo.DeleteGoalTag(c.Request.Context(), userID, goalID, c.Param("name"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidTagName) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if tags == nil {
		tags = []string{}
	}

	c.JSON(http.StatusOK, tags)
}
