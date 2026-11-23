package handlers

import (
	"net/http"
	"strconv"

	"plan-api/internal/models"
	"plan-api/internal/repository"

	"github.com/gin-gonic/gin"
)

// AgentRunHandler hanterar agentkörningar - kopplingen mellan ett mål
// och det arbete en agent gör med det.
type AgentRunHandler struct {
	repo *repository.AgentRunRepository
}

func NewAgentRunHandler(dbConnected bool) *AgentRunHandler {
	return &AgentRunHandler{
		repo: repository.NewAgentRunRepository(),
	}
}

// ListAgentRuns godoc
// @Summary List agent runs
// @Description Get agent runs for the authenticated user, optionally filtered by goal
// @Tags agent-runs
// @Produce json
// @Security BearerAuth
// @Param goal_id query int false "Filter by goal ID"
// @Success 200 {array} models.AgentRun
// @Failure 401 {object} models.ErrorResponse
// @Router /api/agent-runs [get]
func (h *AgentRunHandler) ListAgentRuns(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var goalID *int64
	if raw := c.Query("goal_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid goal_id"})
			return
		}
		goalID = &parsed
	}

	runs, err := h.repo.List(c.Request.Context(), userID, goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, runs)
}

// CreateAgentRun godoc
// @Summary Start an agent run
// @Description Register that an agent started working on a goal
// @Tags agent-runs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateAgentRunRequest true "Agent run"
// @Success 201 {object} models.AgentRun
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/agent-runs [post]
func (h *AgentRunHandler) CreateAgentRun(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.CreateAgentRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "goal_id is required"})
		return
	}

	if req.Phase != "" {
		if err := models.ValidateAgentRunPhase(req.Phase); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
	}
	if req.Status != "" {
		if err := models.ValidateAgentRunStatus(req.Status); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
	}

	run, err := h.repo.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, run)
}

// UpdateAgentRun godoc
// @Summary Update an agent run
// @Description Update phase, status, summary, cost or artifact of an agent run
// @Tags agent-runs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Agent run ID"
// @Param request body models.UpdateAgentRunRequest true "Fields to update"
// @Success 200 {object} models.AgentRun
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/agent-runs/{id} [patch]
func (h *AgentRunHandler) UpdateAgentRun(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	runID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid agent run ID"})
		return
	}

	var req models.UpdateAgentRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid request body"})
		return
	}

	if req.Phase != nil {
		if err := models.ValidateAgentRunPhase(*req.Phase); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
	}
	if req.Status != nil {
		if err := models.ValidateAgentRunStatus(*req.Status); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
	}

	run, err := h.repo.Update(c.Request.Context(), userID, runID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Agent run not found"})
		return
	}

	c.JSON(http.StatusOK, run)
}
