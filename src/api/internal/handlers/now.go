package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"plan/api/internal/models"
	"plan/api/internal/nowlist"

	"github.com/gin-gonic/gin"
)

// NowGoal är ett mål i "Gör nu"-listan.
type NowGoal struct {
	ID               int64           `json:"id"`
	Title            string          `json:"title"`
	Icon             string          `json:"icon"`
	Due              *models.DueDate `json:"due"`
	EstimatedMinutes *int            `json:"estimated_minutes"`
	// Depth är hur djupt målet ligger i hierarkin. 0 = toppnivå.
	Depth int `json:"depth"`
	// Number är målets nummer bland toppnivåmålen. 0 = barnmål.
	Number int `json:"number"`
	// ParentID är föräldern målet visas under. Null = toppnivå.
	ParentID *int64 `json:"parent_id,omitempty"`
	// Places är namnen på de platser målet gäller.
	Places []string `json:"places"`
	// Tags är etiketterna på målet.
	Tags []string `json:"tags"`
}

// NowListResponse är svaret på GET /api/v1/goals/now.
type NowListResponse struct {
	// Goals är listan i den ordning den visas.
	Goals []NowGoal `json:"goals"`
	// Count är antalet mål i listan.
	Count int `json:"count"`
	// Place är den plats listan filtrerades på. Null = ingen filtrering.
	Place *models.Place `json:"place"`
}

// GetNowGoals godoc
// @Summary Get the "Gör nu" list
// @Description Get the goals on the "Gör nu" page, in the order they are shown.
// @Description The list is filtered by the user's current place unless
// @Description "place" says otherwise.
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param place query string false "Place: current (default), none or a place id"
// @Success 200 {object} NowListResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/goals/now [get]
func (h *GoalHandler) GetNowGoals(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	ctx := c.Request.Context()

	goals, err := h.repo.GetPrioritizedGoals(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	// Platserna och taggarna hämtas i ett svep.
	plain := make([]models.Goal, len(goals))
	for i, g := range goals {
		plain[i] = g.Goal
	}
	h.attachPlaces(ctx, plain)
	h.attachTags(ctx, plain)
	for i := range goals {
		goals[i].Places = plain[i].Places
		goals[i].Tags = plain[i].Tags
	}

	place, placeID, err := h.resolveNowPlace(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	rows := nowlist.Build(nowGoalsFrom(goals), placeID)
	byID := prioritizedByID(goals)

	response := NowListResponse{
		Goals: make([]NowGoal, 0, len(rows)),
		Count: len(rows),
		Place: place,
	}
	for _, row := range rows {
		goal := row.Goal

		// Målet hämtas ur kartan för att få med platser och taggar.
		prioritized := byID[goal.ID]

		names := make([]string, 0, len(prioritized.Places))
		for _, p := range prioritized.Places {
			names = append(names, p.Name)
		}

		response.Goals = append(response.Goals, NowGoal{
			ID:               prioritized.ID,
			Title:            prioritized.Title,
			Icon:             prioritized.Icon,
			Due:              prioritized.Due,
			EstimatedMinutes: prioritized.EstimatedMinutes,
			Depth:            row.Depth,
			Number:           row.Number,
			ParentID:         row.ParentID,
			Places:           names,
			Tags:             prioritized.Tags,
		})
	}

	c.JSON(http.StatusOK, response)
}

// resolveNowPlace tar reda på vilken plats listan ska filtreras på.
//
// "current" (standard) använder användarens "jag är här"-plats. Har
// användaren inte valt någon plats filtreras ingenting bort - en
// platslös vy ska inte tömma listan. "none" stänger av filtreringen
// helt, och ett id anger en viss plats.
func (h *GoalHandler) resolveNowPlace(c *gin.Context, userID int64) (*models.Place, *int64, error) {
	raw := c.Query("place")
	if raw == "" {
		raw = "current"
	}

	switch raw {
	case "none", "all":
		return nil, nil, nil
	case "current":
		place, err := h.places.GetCurrentPlace(c.Request.Context(), userID)
		if err != nil {
			return nil, nil, err
		}
		if place == nil {
			return nil, nil, nil
		}
		return place, &place.ID, nil
	default:
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("okänt platsvärde: %q", raw)
		}

		place, err := h.places.Get(c.Request.Context(), userID, id)
		if err != nil {
			return nil, nil, err
		}
		if place == nil {
			return nil, nil, nil
		}
		return place, &place.ID, nil
	}
}

// nowGoalsFrom plockar ut de fält som listan behöver.
func nowGoalsFrom(goals []models.PrioritizedGoal) []nowlist.Goal {
	converted := make([]nowlist.Goal, 0, len(goals))
	for _, goal := range goals {
		parentIDs := make([]int64, 0, len(goal.Parents))
		for _, parent := range goal.Parents {
			parentIDs = append(parentIDs, parent.ID)
		}

		placeIDs := make([]int64, 0, len(goal.Places))
		for _, place := range goal.Places {
			placeIDs = append(placeIDs, place.ID)
		}

		converted = append(converted, nowlist.Goal{
			ID:        goal.ID,
			Now:       goal.Now,
			Started:   goal.Started,
			ParentIDs: parentIDs,
			PlaceIDs:  placeIDs,
		})
	}
	return converted
}

// prioritizedByID ger en sökbar karta över målen.
func prioritizedByID(goals []models.PrioritizedGoal) map[int64]models.PrioritizedGoal {
	byID := make(map[int64]models.PrioritizedGoal, len(goals))
	for _, goal := range goals {
		byID[goal.ID] = goal
	}
	return byID
}
