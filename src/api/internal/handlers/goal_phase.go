package handlers

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"plan/api/internal/models"
	"plan/api/internal/phase"
	"plan/api/internal/phaseservice"
	"plan/api/internal/repository"

	"github.com/gin-gonic/gin"
)

// GoalPhaseHandler flyttar mål mellan faser.
//
// Övergångarna går genom fasmaskinen (package phase), som avgör om
// övergången är tillåten. Handlerens jobb är att hämta faktorna, kalla
// maskinen och skriva resultatet.
type GoalPhaseHandler struct {
	service *phaseservice.Service
	goals   *repository.GoalRepository
}

// NewGoalPhaseHandler ger en handler för fasövergångar.
func NewGoalPhaseHandler(dbConnected bool) *GoalPhaseHandler {
	return &GoalPhaseHandler{
		service: phaseservice.New(),
		goals:   repository.NewGoalRepository(),
	}
}

// TransitionPhase godoc
// @Summary Move a goal to another phase
// @Description Moves a goal between the phases define, plan, execute, verify and completed. The transition rules live in the phase machine - a missing precondition is answered with 409 and a stable reason code.
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Param request body models.PhaseTransitionRequest true "Target phase"
// @Success 200 {object} models.PhaseTransitionResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/phase [post]
func (h *GoalPhaseHandler) TransitionPhase(c *gin.Context) {
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

	hasAccess, _ := h.goals.UserHasAccess(c.Request.Context(), userID, goalID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	var req models.PhaseTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "to is required"})
		return
	}

	response, status, code, message := h.transition(c, userID, goalID, req.To, req.OverrideReason)
	if status != http.StatusOK {
		writePhaseError(c, status, code, message)
		return
	}

	c.JSON(http.StatusOK, response)
}

// CompleteGoal godoc
// @Summary Complete a goal directly
// @Description Completes a goal in one step, from any phase. This is the human path - the goal is completed as "deklarerat", or as "verifierat" if a verification has passed.
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Goal ID"
// @Param request body models.CompleteGoalRequest true "Reason"
// @Success 200 {object} models.PhaseTransitionResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/goals/{id}/complete [post]
func (h *GoalPhaseHandler) CompleteGoal(c *gin.Context) {
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

	hasAccess, _ := h.goals.UserHasAccess(c.Request.Context(), userID, goalID)
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Access denied"})
		return
	}

	var req models.CompleteGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "invalid body"})
		return
	}

	response, status, code, message := h.transition(c, userID, goalID, phase.Completed, req.Reason)
	if status != http.StatusOK {
		writePhaseError(c, status, code, message)
		return
	}

	c.JSON(http.StatusOK, response)
}

// transition gör själva övergången och svarar med statuskod och meddelande.
//
// Ett villkor som inte är uppfyllt ger 409 - då vet klienten att det
// finns något att fixa, inte bara att något gick fel.
func (h *GoalPhaseHandler) transition(
	c *gin.Context, userID, goalID int64, to, overrideReason string,
) (models.PhaseTransitionResponse, int, string, string) {
	ctx := c.Request.Context()

	current, err := h.goals.GetGoalStatus(ctx, goalID)
	if err != nil {
		return models.PhaseTransitionResponse{}, http.StatusInternalServerError, "", err.Error()
	}
	if current == nil {
		return models.PhaseTransitionResponse{}, http.StatusNotFound, "", "goal not found"
	}

	// En agent-nyckel får inte göra samma sak som en människa. Scopen
	// sätts av API-nyckelmiddleware, och en människas session har ingen
	// scope alls.
	isHuman := !isAgentRequest(c)

	outcome, err := h.service.Apply(ctx, goalID, to, isHuman, overrideReason)
	if err != nil {
		if errors.Is(err, phaseservice.ErrGoalNotFound) {
			return models.PhaseTransitionResponse{}, http.StatusNotFound, "", "goal not found"
		}
		return models.PhaseTransitionResponse{}, http.StatusInternalServerError, "", err.Error()
	}
	result := outcome.Result
	if !result.Allowed {
		status := http.StatusConflict
		if result.Code == phase.ReasonUnknownPhase {
			status = http.StatusBadRequest
		}
		return models.PhaseTransitionResponse{}, status, result.Code, result.Message
	}

	return models.PhaseTransitionResponse{
		GoalID:         goalID,
		From:           current.Name,
		To:             to,
		Forced:         result.Forced,
		CompletionKind: result.CompletionKind,
	}, http.StatusOK, "", ""
}

// writePhaseError svarar med en stabil skälkod, så att klienten kan
// visa vad som saknas - inte bara att något gick fel.
func writePhaseError(c *gin.Context, status int, code, message string) {
	body := gin.H{"error": http.StatusText(status), "message": message}
	if code != "" {
		body["code"] = code
	}
	c.JSON(status, body)
}

// isAgentRequest svarar på om anropet kommer från en agent-nyckel.
func isAgentRequest(c *gin.Context) bool {
	scope, exists := c.Get("apiKeyScope")
	if !exists {
		return false
	}
	value, ok := scope.(string)
	return ok && value == "agent"
}
