package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"plan-api/internal/models"
	"plan-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type CalendarHandler struct {
	repo        *repository.CalendarRepository
	dbConnected bool
}

func NewCalendarHandler(dbConnected bool) *CalendarHandler {
	return &CalendarHandler{
		repo:        repository.NewCalendarRepository(),
		dbConnected: dbConnected,
	}
}

func (h *CalendarHandler) checkCalendarDB(c *gin.Context) bool {
	if !h.dbConnected {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Message: "Database not connected."})
		return false
	}
	return true
}

func parseCalendarID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid calendar ID"})
		return 0, false
	}
	return id, true
}

// validateCalendarName kontrollerar att namnet är användbart.
func validateCalendarName(name *string) (*string, bool) {
	if name == nil {
		return nil, true
	}
	trimmed := strings.TrimSpace(*name)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 64 {
		return nil, false
	}
	return &trimmed, true
}

// List godoc
// @Summary List calendars
// @Description List the calendars owned by the authenticated user.
// @Tags calendars
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Calendar
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/calendars [get]
func (h *CalendarHandler) List(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	calendars, err := h.repo.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, calendars)
}

// Create godoc
// @Summary Create calendar
// @Description Create a new calendar for the authenticated user.
// @Tags calendars
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} models.Calendar
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/calendars [post]
func (h *CalendarHandler) Create(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	name, ok := validateCalendarName(&req.Name)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Calendar name must be between 1 and 64 characters"})
		return
	}
	req.Name = *name

	color, ok := validateColor(req.Color)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid color"})
		return
	}
	req.Color = color

	icon, ok := validateIcon(req.Icon)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid icon"})
		return
	}
	req.Icon = icon

	calendar, err := h.repo.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, calendar)
}

// Update godoc
// @Summary Update calendar
// @Description Update the name, color, icon or order of a calendar.
// @Tags calendars
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.Calendar
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/calendars/{id} [patch]
func (h *CalendarHandler) Update(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}
	calendarID, ok := parseCalendarID(c)
	if !ok {
		return
	}

	var req models.UpdateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	name, ok := validateCalendarName(req.Name)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Calendar name must be between 1 and 64 characters"})
		return
	}
	req.Name = name

	color, ok := validateColor(req.Color)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid color"})
		return
	}
	req.Color = color

	icon, ok := validateIcon(req.Icon)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid icon"})
		return
	}
	req.Icon = icon

	calendar, err := h.repo.Update(c.Request.Context(), userID, calendarID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if calendar == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Calendar not found"})
		return
	}
	c.JSON(http.StatusOK, calendar)
}

// Delete godoc
// @Summary Delete calendar
// @Description Delete a calendar. Events in it are kept but lose their calendar.
// @Tags calendars
// @Produce json
// @Security BearerAuth
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/calendars/{id} [delete]
func (h *CalendarHandler) Delete(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}
	calendarID, ok := parseCalendarID(c)
	if !ok {
		return
	}

	deleted, err := h.repo.Delete(c.Request.Context(), userID, calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Calendar not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
