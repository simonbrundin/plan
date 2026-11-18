package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teambition/rrule-go"
	"plan-api/internal/models"
	"plan-api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type CalendarEventHandler struct {
	repo        *repository.CalendarEventRepository
	calendars   *repository.CalendarRepository
	dbConnected bool
}

func NewCalendarEventHandler(dbConnected bool) *CalendarEventHandler {
	return &CalendarEventHandler{
		repo:        repository.NewCalendarEventRepository(),
		calendars:   repository.NewCalendarRepository(),
		dbConnected: dbConnected,
	}
}

func (h *CalendarEventHandler) checkCalendarDB(c *gin.Context) bool {
	if !h.dbConnected {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Message: "Database not connected."})
		return false
	}
	return true
}

func parseCalendarTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

// checkCalendarOwnership kontrollerar att angiven kalender finns och
// tillhör användaren. Värdet 0 betyder "ingen kalender" och är tillåtet,
// eftersom det används för att ta bort en kalenderkoppling.
func (h *CalendarEventHandler) checkCalendarOwnership(c *gin.Context, userID int64, calendarID *int64) bool {
	if calendarID == nil || *calendarID == 0 {
		return true
	}
	hasCalendar, err := h.repo.UserHasCalendar(c.Request.Context(), userID, *calendarID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return false
	}
	if !hasCalendar {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Calendar access denied"})
		return false
	}
	return true
}

func parseEventID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid calendar event ID"})
		return 0, false
	}
	return id, true
}

// validateIcon normaliserar en ikon (emoji eller Iconify-namn). En tom
// sträng är tillåten och betyder "ingen ikon".
func validateIcon(icon *string) (*string, bool) {
	if icon == nil {
		return nil, true
	}
	trimmed := strings.TrimSpace(*icon)
	if utf8.RuneCountInString(trimmed) > 100 {
		return nil, false
	}
	return &trimmed, true
}

func validateColor(color *string) (*string, bool) {
	if color == nil {
		return nil, true
	}
	trimmed := strings.TrimSpace(*color)
	if trimmed == "" {
		return &trimmed, true
	}
	if len(trimmed) != 7 || trimmed[0] != '#' {
		return nil, false
	}
	for _, character := range trimmed[1:] {
		if !strings.ContainsRune("0123456789abcdefABCDEF", character) {
			return nil, false
		}
	}
	return &trimmed, true
}

// List godoc
// @Summary List calendar events
// @Description List calendar events for the authenticated user. Without a time window the recurring series are returned unexpanded, so the client can expand them locally. Within a time window the recurring events are expanded into occurrences.
// @Tags calendar-events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param scope query string false "all returns every event without expanding recurring series"
// @Param from query string false "Start of the window (RFC3339), required unless scope is all"
// @Param to query string false "End of the window (RFC3339), required unless scope is all"
// @Success 200 {array} models.CalendarEvent
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/calendar-events [get]
func (h *CalendarEventHandler) List(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	switch c.Query("scope") {
	case "all":
		h.listAll(c, userID)
	case "":
		h.listInWindow(c, userID)
	default:
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "scope must be 'all'"})
	}
}

// listAll svarar med alla händelser, där återkommande serier inte
// expanderas. Klienten expanderar dem själv.
func (h *CalendarEventHandler) listAll(c *gin.Context, userID int64) {
	events, err := h.repo.ListAll(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// listInWindow svarar med de händelser som ligger i tidsfönstret, där
// återkommande serier expanderas till förekomster.
func (h *CalendarEventHandler) listInWindow(c *gin.Context, userID int64) {
	from, err := parseCalendarTime(c.Query("from"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "from must be an RFC3339 timestamp"})
		return
	}
	to, err := parseCalendarTime(c.Query("to"))
	if err != nil || !to.After(from) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "to must be an RFC3339 timestamp after from"})
		return
	}

	events, err := h.repo.List(c.Request.Context(), userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// Get godoc
// @Summary Get a calendar event
// @Description Get a single calendar event by ID
// @Tags calendar-events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calendar event ID"
// @Success 200 {object} models.CalendarEvent
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/calendar-events/{id} [get]
func (h *CalendarEventHandler) Get(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}
	eventID, ok := parseEventID(c)
	if !ok {
		return
	}

	event, err := h.repo.Get(c.Request.Context(), userID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Calendar event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

// Create godoc
// @Summary Create a calendar event
// @Description Create a new calendar event, optionally recurring and optionally linked to a goal
// @Tags calendar-events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param event body models.CreateCalendarEventRequest true "Calendar event"
// @Success 201 {object} models.CalendarEvent
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/calendar-events [post]
func (h *CalendarEventHandler) Create(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.CreateCalendarEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "title, start_at and end_at are required"})
		return
	}
	startAt, err := parseCalendarTime(req.StartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "start_at must be an RFC3339 timestamp"})
		return
	}
	endAt, err := parseCalendarTime(req.EndAt)
	if err != nil || !endAt.After(startAt) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "end_at must be an RFC3339 timestamp after start_at"})
		return
	}

	if req.GoalID != nil {
		hasGoal, err := h.repo.UserHasGoal(c.Request.Context(), userID, *req.GoalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		if !hasGoal {
			c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Goal access denied"})
			return
		}
	}
	if !h.checkCalendarOwnership(c, userID, req.CalendarID) {
		return
	}
	if req.RecurrenceRule != nil {
		if _, err := rrule.StrToRRule(*req.RecurrenceRule); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid recurrence_rule"})
			return
		}
	}
	icon, ok := validateIcon(req.Icon)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "icon must be at most 100 characters"})
		return
	}
	req.Icon = icon
	color, ok := validateColor(req.Color)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "color must be a hexadecimal color, for example #6366f1"})
		return
	}
	req.Color = color

	if req.CalendarID != nil && *req.CalendarID == 0 {
		// 0 betyder uttryckligen "ingen kalender".
		req.CalendarID = nil
	} else if req.CalendarID == nil {
		// Ingen kalender angiven - händelsen hamnar i standardkalendern
		// om användaren har satt någon.
		defaultCalendarID, err := h.calendars.DefaultCalendarID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		req.CalendarID = defaultCalendarID
	}

	event, err := h.repo.Create(c.Request.Context(), userID, req, startAt, endAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, event)
}

// Update godoc
// @Summary Update a calendar event
// @Description Update a calendar event. The icon field has three states: omitted keeps the current icon, an empty string clears it and a value sets it.
// @Tags calendar-events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calendar event ID"
// @Param event body models.UpdateCalendarEventRequest true "Fields to update"
// @Success 200 {object} models.CalendarEvent
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/calendar-events/{id} [patch]
func (h *CalendarEventHandler) Update(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}
	eventID, ok := parseEventID(c)
	if !ok {
		return
	}

	var req models.UpdateCalendarEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid calendar event"})
		return
	}
	var startAt, endAt *time.Time
	if req.StartAt != nil {
		parsed, err := parseCalendarTime(*req.StartAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "start_at must be an RFC3339 timestamp"})
			return
		}
		startAt = &parsed
	}
	if req.EndAt != nil {
		parsed, err := parseCalendarTime(*req.EndAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "end_at must be an RFC3339 timestamp"})
			return
		}
		endAt = &parsed
	}
	if startAt != nil && endAt != nil && !endAt.After(*startAt) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "end_at must be after start_at"})
		return
	}
	if startAt != nil || endAt != nil {
		existing, err := h.repo.Get(c.Request.Context(), userID, eventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		if existing == nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Calendar event not found"})
			return
		}
		if startAt == nil {
			startAt = &existing.StartAt
		}
		if endAt == nil {
			endAt = &existing.EndAt
		}
		if !endAt.After(*startAt) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "end_at must be after start_at"})
			return
		}
	}

	if req.GoalID != nil {
		hasGoal, err := h.repo.UserHasGoal(c.Request.Context(), userID, *req.GoalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		if !hasGoal {
			c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Goal access denied"})
			return
		}
	}
	if !h.checkCalendarOwnership(c, userID, req.CalendarID) {
		return
	}
	if req.RecurrenceRule != nil {
		if _, err := rrule.StrToRRule(*req.RecurrenceRule); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid recurrence_rule"})
			return
		}
	}
	icon, ok := validateIcon(req.Icon)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "icon must be at most 100 characters"})
		return
	}
	req.Icon = icon
	color, ok := validateColor(req.Color)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "color must be a hexadecimal color, for example #6366f1"})
		return
	}
	req.Color = color

	event, err := h.repo.Update(c.Request.Context(), userID, eventID, req, startAt, endAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Calendar event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

// CreateOccurrenceOverride godoc
// @Summary Override a single occurrence
// @Description Create an override for one occurrence in a recurring series. Unspecified fields are inherited from the parent event.
// @Tags calendar-events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Parent calendar event ID"
// @Param occurrence body models.CreateCalendarOccurrenceRequest true "Occurrence override"
// @Success 201 {object} models.CalendarEvent
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/calendar-events/{id}/occurrences [post]
func (h *CalendarEventHandler) CreateOccurrenceOverride(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}
	parentID, ok := parseEventID(c)
	if !ok {
		return
	}

	var req models.CreateCalendarOccurrenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "original_start_at, start_at and end_at are required"})
		return
	}
	originalStart, err := parseCalendarTime(req.OriginalStartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "original_start_at must be an RFC3339 timestamp"})
		return
	}
	startAt, err := parseCalendarTime(req.StartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "start_at must be an RFC3339 timestamp"})
		return
	}
	endAt, err := parseCalendarTime(req.EndAt)
	if err != nil || !endAt.After(startAt) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "end_at must be after start_at"})
		return
	}
	if req.Status != "" && req.Status != "confirmed" && req.Status != "tentative" && req.Status != "cancelled" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid status"})
		return
	}
	if req.GoalID != nil {
		hasGoal, err := h.repo.UserHasGoal(c.Request.Context(), userID, *req.GoalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		if !hasGoal {
			c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Goal access denied"})
			return
		}
	}
	if !h.checkCalendarOwnership(c, userID, req.CalendarID) {
		return
	}
	icon, ok := validateIcon(req.Icon)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "icon must be at most 100 characters"})
		return
	}
	req.Icon = icon
	color, ok := validateColor(req.Color)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "color must be a hexadecimal color, for example #6366f1"})
		return
	}
	req.Color = color

	event, err := h.repo.CreateOverride(c.Request.Context(), userID, parentID, req, originalStart, startAt, endAt)
	if err == repository.ErrInvalidTimestamp {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}
	if err == nil {
		c.JSON(http.StatusCreated, event)
		return
	}
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Recurring calendar event not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
}

// SplitSeries godoc
// @Summary Split a recurring series
// @Description End a recurring series before a chosen occurrence and continue it as a new series
// @Tags calendar-events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Parent calendar event ID"
// @Param split body models.SplitCalendarEventRequest true "Split request"
// @Success 201 {object} models.CalendarEvent
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/calendar-events/{id}/split [post]
func (h *CalendarEventHandler) SplitSeries(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}
	parentID, ok := parseEventID(c)
	if !ok {
		return
	}
	var req models.SplitCalendarEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "original_start_at, start_at, end_at and recurrence_rule are required"})
		return
	}
	originalStart, err := parseCalendarTime(req.OriginalStartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "original_start_at must be an RFC3339 timestamp"})
		return
	}
	startAt, err := parseCalendarTime(req.StartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "start_at must be an RFC3339 timestamp"})
		return
	}
	endAt, err := parseCalendarTime(req.EndAt)
	if err != nil || !endAt.After(startAt) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "end_at must be after start_at"})
		return
	}
	if _, err := rrule.StrToRRule(req.RecurrenceRule); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid recurrence_rule"})
		return
	}
	if req.GoalID != nil {
		hasGoal, err := h.repo.UserHasGoal(c.Request.Context(), userID, *req.GoalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		if !hasGoal {
			c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Goal access denied"})
			return
		}
	}
	if !h.checkCalendarOwnership(c, userID, req.CalendarID) {
		return
	}
	icon, ok := validateIcon(req.Icon)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "icon must be at most 100 characters"})
		return
	}
	req.Icon = icon
	color, ok := validateColor(req.Color)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "color must be a hexadecimal color, for example #6366f1"})
		return
	}
	req.Color = color
	created, err := h.repo.SplitSeries(c.Request.Context(), userID, parentID, req, originalStart, startAt, endAt)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Recurring calendar event not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// Delete godoc
// @Summary Delete a calendar event
// @Description Delete a calendar event
// @Tags calendar-events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calendar event ID"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/calendar-events/{id} [delete]
func (h *CalendarEventHandler) Delete(c *gin.Context) {
	if !h.checkCalendarDB(c) {
		return
	}
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}
	eventID, ok := parseEventID(c)
	if !ok {
		return
	}
	deleted, err := h.repo.Delete(c.Request.Context(), userID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Calendar event not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
