package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"plan/api/internal/models"
	"plan/api/internal/repository"

	"github.com/gin-gonic/gin"
)

type PlaceHandler struct {
	repo    *repository.PlaceRepository
	goals   *repository.GoalRepository
	events  *repository.CalendarEventRepository
	dbReady bool
}

func NewPlaceHandler(dbConnected bool) *PlaceHandler {
	return &PlaceHandler{
		repo:    repository.NewPlaceRepository(),
		goals:   repository.NewGoalRepository(),
		events:  repository.NewCalendarEventRepository(),
		dbReady: dbConnected,
	}
}

// checkDb returns error response if database is not connected
func (h *PlaceHandler) checkDb(c *gin.Context) bool {
	if !h.dbReady {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Message: "Database not connected. Set DATABASE_URL environment variable.",
		})
		return false
	}
	return true
}

// List godoc
// @Summary List all places
// @Description Get all places for the authenticated user
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Place
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/places [get]
func (h *PlaceHandler) List(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	places, err := h.repo.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if places == nil {
		places = []models.Place{}
	}

	c.JSON(http.StatusOK, places)
}

// Create godoc
// @Summary Create a place
// @Description Create a place for the authenticated user
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param place body models.CreatePlaceRequest true "Place data"
// @Success 201 {object} models.Place
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/places [post]
func (h *PlaceHandler) Create(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.CreatePlaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Name, latitude and longitude are required"})
		return
	}

	radius := models.DefaultPlaceRadiusMeters
	if req.RadiusMeters != nil {
		radius = *req.RadiusMeters
	}
	icon := ""
	if req.Icon != nil {
		icon = *req.Icon
	}

	place := models.Place{
		Name:         req.Name,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		RadiusMeters: radius,
		Icon:         icon,
	}
	if err := place.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	created, err := h.repo.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// Update godoc
// @Summary Update a place
// @Description Update a place's name, coordinates or radius
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Place ID"
// @Param place body models.UpdatePlaceRequest true "Update data"
// @Success 200 {object} models.Place
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/places/{id} [patch]
func (h *PlaceHandler) Update(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	placeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid place ID"})
		return
	}

	var req models.UpdatePlaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "No fields to update"})
		return
	}

	existing, err := h.repo.Get(c.Request.Context(), userID, placeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Place not found"})
		return
	}

	// Validera det som faktiskt ändras - oförändrade värden tas från
	// den befintliga platsen.
	updated := *existing
	if req.Name != nil {
		updated.Name = *req.Name
	}
	if req.Latitude != nil {
		updated.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		updated.Longitude = *req.Longitude
	}
	if req.RadiusMeters != nil {
		updated.RadiusMeters = *req.RadiusMeters
	}
	if req.Icon != nil {
		updated.Icon = *req.Icon
	}
	if err := updated.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	place, err := h.repo.Update(c.Request.Context(), userID, placeID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if place == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Place not found"})
		return
	}

	c.JSON(http.StatusOK, place)
}

// Delete godoc
// @Summary Delete a place
// @Description Delete a place and its links to goals and events
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Place ID"
// @Success 200 {object} map[string]bool
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/places/{id} [delete]
func (h *PlaceHandler) Delete(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	placeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid place ID"})
		return
	}

	deleted, err := h.repo.Delete(c.Request.Context(), userID, placeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Place not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "deleted": placeID})
}

// Nearby godoc
// @Summary Find places near a position
// @Description Get the places whose radius contains the given position
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param lat query number true "Latitude"
// @Param lon query number true "Longitude"
// @Success 200 {array} models.Place
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/places/nearby [get]
func (h *PlaceHandler) Nearby(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "lat must be a number"})
		return
	}
	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "lon must be a number"})
		return
	}

	places, err := h.repo.GetNearbyPlaces(c.Request.Context(), userID, lat, lon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if places == nil {
		places = []models.Place{}
	}

	c.JSON(http.StatusOK, places)
}

// GetGoalPlaces godoc
// @Summary List a goal's places
// @Description Get the places linked to a goal
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Success 200 {array} models.Place
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/places [get]
func (h *PlaceHandler) GetGoalPlaces(c *gin.Context) {
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

	places, err := h.repo.GetGoalPlaces(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if places == nil {
		places = []models.Place{}
	}

	c.JSON(http.StatusOK, places)
}

// SetGoalPlaces godoc
// @Summary Replace a goal's places
// @Description Replace the places linked to a goal. An empty list removes all places.
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Param places body models.SetPlacesRequest true "Place IDs"
// @Success 200 {array} models.Place
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/places [put]
func (h *PlaceHandler) SetGoalPlaces(c *gin.Context) {
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

	var req models.SetPlacesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "place_ids is required"})
		return
	}

	places, err := h.repo.SetGoalPlaces(c.Request.Context(), userID, goalID, req.PlaceIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if places == nil {
		places = []models.Place{}
	}

	c.JSON(http.StatusOK, places)
}

// GetEventPlaces godoc
// @Summary List a calendar event's places
// @Description Get the places linked to a calendar event
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Success 200 {array} models.Place
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/calendar-events/{id}/places [get]
func (h *PlaceHandler) GetEventPlaces(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid event ID"})
		return
	}

	event, err := h.events.Get(c.Request.Context(), userID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Event not found"})
		return
	}

	places, err := h.repo.GetEventPlaces(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if places == nil {
		places = []models.Place{}
	}

	c.JSON(http.StatusOK, places)
}

// SetEventPlaces godoc
// @Summary Replace a calendar event's places
// @Description Replace the places linked to a calendar event. An empty list removes all places.
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Event ID"
// @Param places body models.SetPlacesRequest true "Place IDs"
// @Success 200 {array} models.Place
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/calendar-events/{id}/places [put]
func (h *PlaceHandler) SetEventPlaces(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid event ID"})
		return
	}

	event, err := h.events.Get(c.Request.Context(), userID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Event not found"})
		return
	}

	var req models.SetPlacesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "place_ids is required"})
		return
	}

	places, err := h.repo.SetEventPlaces(c.Request.Context(), userID, eventID, req.PlaceIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if places == nil {
		places = []models.Place{}
	}

	c.JSON(http.StatusOK, places)
}

// GetCurrentPlace godoc
// @Summary Get the current place
// @Description Get the place the user has chosen as "I am here"
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.Place
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/me/current-place [get]
func (h *PlaceHandler) GetCurrentPlace(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	place, err := h.repo.GetCurrentPlace(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if place == nil {
		c.JSON(http.StatusOK, gin.H{"place": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"place": place})
}

// SetCurrentPlace godoc
// @Summary Set the current place
// @Description Choose the place the user is at, or clear it with null
// @Tags places
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param place body models.SetCurrentPlaceRequest true "Place ID or null"
// @Success 200 {object} models.Place
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/me/current-place [put]
func (h *PlaceHandler) SetCurrentPlace(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.SetCurrentPlaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "place_id is required"})
		return
	}

	if req.PlaceID != nil {
		place, err := h.repo.Get(c.Request.Context(), userID, *req.PlaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		if place == nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Place not found"})
			return
		}
	}

	if err := h.repo.SetCurrentPlace(c.Request.Context(), userID, req.PlaceID); err != nil {
		if errors.Is(err, repository.ErrPlaceNotFound) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Place not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	place, err := h.repo.GetCurrentPlace(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if place == nil {
		c.JSON(http.StatusOK, gin.H{"place": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"place": place})
}
