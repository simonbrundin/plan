package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"plan-api/internal/models"
	"plan-api/internal/repository"

	"github.com/gin-gonic/gin"
)

// AgentRunHandler hanterar agentkörningar - kopplingen mellan ett mål
// och det arbete en agent gör med det.
type AgentRunHandler struct {
	repo      *repository.AgentRunRepository
	eventRepo *repository.AgentRunEventRepository
	usageRepo *repository.AgentUsageRepository
}

func NewAgentRunHandler(dbConnected bool) *AgentRunHandler {
	return &AgentRunHandler{
		repo:      repository.NewAgentRunRepository(),
		eventRepo: repository.NewAgentRunEventRepository(),
		usageRepo: repository.NewAgentUsageRepository(),
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
// @Router /api/v1/agent-runs [get]
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
// @Router /api/v1/agent-runs [post]
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
	if req.Trigger != "" {
		if err := models.ValidateAgentRunTrigger(req.Trigger); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
	}
	if req.AgentProfile != "" {
		if err := models.ValidateAgentProfile(req.AgentProfile); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
	}

	run, warning, err := h.repo.Create(c.Request.Context(), userID, req)
	if err != nil {
		// En budget som är slut stoppar körningen. Då ska skälet fram,
		// inte bara ett felmeddelande.
		var budgetErr *models.BudgetExceededError
		if errors.As(err, &budgetErr) {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Message: budgetErr.Error(),
				Details: budgetErr.Status,
			})
			return
		}
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, agentRunResponse{AgentRun: run, BudgetWarning: warning})
}

// agentRunResponse är svaret när en körning skapas.
//
// BudgetWarning finns med när varningsgränsen är nådd - körningen får
// starta, men människan behöver veta att budgeten är nära slut.
type agentRunResponse struct {
	*models.AgentRun
	BudgetWarning *models.BudgetStatus `json:"budget_warning,omitempty"`
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
// @Router /api/v1/agent-runs/{id} [patch]
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
	if req.Trigger != nil {
		if err := models.ValidateAgentRunTrigger(*req.Trigger); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
	}
	if req.AgentProfile != nil {
		if err := models.ValidateAgentProfile(*req.AgentProfile); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
	}

	current, err := h.repo.Get(c.Request.Context(), userID, runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if current == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Agent run not found"})
		return
	}
	if err := models.ValidateAgentRunTransition(*current, req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
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
	if models.IsTerminalAgentRunStatus(run.Status) {
		h.ensureUsageRow(c, userID, runID)
	}

	c.JSON(http.StatusOK, run)
}

// ensureUsageRow skapar en kostnadsrad om körningen saknar en.
//
// Varje avslutad körning ska ha en rad - annars ser det ut som att
// agenten aldrig kostade något. Raden markeras som ofullständig när
// agenten inte hann rapportera användningen.
func (h *AgentRunHandler) ensureUsageRow(c *gin.Context, userID, runID int64) {
	if err := h.usageRepo.EnsureRow(c.Request.Context(), userID, runID); err != nil {
		// Körningen är sparad. Att kostnadsraden saknas får inte stoppa
		// svaret - det loggas i stället.
		log.Printf("Failed to ensure usage row for run %d: %v", runID, err)
	}
}

// CancelAgentRun godoc
// @Summary Cancel an agent run
// @Description Cancel a running or queued agent run. The agent stops and the cluster resources are cleaned up.
// @Tags agent-runs
// @Produce json
// @Security BearerAuth
// @Param id path int true "Agent run ID"
// @Success 200 {object} models.AgentRun
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/agent-runs/{id}/cancel [post]
func (h *AgentRunHandler) CancelAgentRun(c *gin.Context) {
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

	current, err := h.repo.Get(c.Request.Context(), userID, runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if current == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Agent run not found"})
		return
	}

	status := models.AgentRunStatusCancelled
	summary := "Körningen avbröts av en människa"
	result := "Körningen avbröts innan den blev klar"
	finishedAt := time.Now().UTC()

	// En avslutad körning kan inte avbrytas - avbrytning är ett sätt att
	// stanna, inte ett sätt att skriva om historiken. Samma regel gäller
	// som när status uppdateras med PATCH.
	if err := models.ValidateAgentRunTransition(*current, models.UpdateAgentRunRequest{Status: &status}); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	run, err := h.repo.Update(c.Request.Context(), userID, runID, models.UpdateAgentRunRequest{
		Status:     &status,
		Summary:    &summary,
		Result:     &result,
		FinishedAt: &finishedAt,
	})
	if err != nil {
		if errors.Is(err, models.ErrInvalidAgentRun) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Agent run not found"})
		return
	}

	h.ensureUsageRow(c, userID, runID)

	// Händelsen gör det synligt i historiken att det var en människa som
	// avbröt - inte ett fel.
	if _, err := h.eventRepo.Create(c.Request.Context(), userID, runID, models.CreateAgentRunEventRequest{
		Type:  models.AgentRunEventRunCancelled,
		Phase: run.Phase,
		Payload: map[string]any{
			"cancelled_by": "user",
		},
	}); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, run)
}

// ListAgentRunEvents godoc
// @Summary List AgentRun events
// @Description Get the chronological event history for an agent run
// @Tags agent-runs
// @Produce json
// @Security BearerAuth
// @Param id path int true "Agent run ID"
// @Success 200 {array} models.AgentRunEvent
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/agent-runs/{id}/events [get]
func (h *AgentRunHandler) ListAgentRunEvents(c *gin.Context) {
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

	events, err := h.eventRepo.List(c.Request.Context(), userID, runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if events == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Agent run not found"})
		return
	}
	c.JSON(http.StatusOK, events)
}

// CreateAgentRunEvent godoc
// @Summary Record an AgentRun event
// @Description Append a structured event to an agent run
// @Tags agent-runs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Agent run ID"
// @Param request body models.CreateAgentRunEventRequest true "Agent run event"
// @Success 201 {object} models.AgentRunEvent
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/agent-runs/{id}/events [post]
func (h *AgentRunHandler) CreateAgentRunEvent(c *gin.Context) {
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

	var req models.CreateAgentRunEventRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Type == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "type is required"})
		return
	}
	if err := models.ValidateAgentRunEventType(req.Type); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}
	if req.Phase != "" {
		if err := models.ValidateAgentRunPhase(req.Phase); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
			return
		}
	}

	event, err := h.eventRepo.Create(c.Request.Context(), userID, runID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Agent run not found"})
		return
	}
	c.JSON(http.StatusCreated, event)
}
