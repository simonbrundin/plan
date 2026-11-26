package handlers

import (
	"net/http"
	"strconv"
	"time"

	"plan/api/internal/models"
	"plan/api/internal/repository"

	"github.com/gin-gonic/gin"
)

// AgentUsageHandler hanterar kostnadsdata och budgetgränser för
// agentkörningar.
//
// Plan äger kostnaderna. Agenten rapporterar bara användningen - räknandet
// görs här, med en versionsstyrd prislista, så att kostnaden kan
// reproduceras.
type AgentUsageHandler struct {
	usageRepo  *repository.AgentUsageRepository
	budgetRepo *repository.AgentBudgetRepository
}

// NewAgentUsageHandler skapar en handler för kostnad och budgetar.
func NewAgentUsageHandler() *AgentUsageHandler {
	return &AgentUsageHandler{
		usageRepo:  repository.NewAgentUsageRepository(),
		budgetRepo: repository.NewAgentBudgetRepository(),
	}
}

// RecordUsage godoc
// @Summary Record usage for an agent run
// @Description Store the token usage of an agent run. The cost is calculated in Plan from a versioned price list.
// @Tags agent-usage
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Agent run ID"
// @Param request body models.CreateAgentRunUsageRequest true "Usage to record"
// @Success 200 {object} models.AgentRunUsage
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/agent-runs/{id}/usage [post]
func (h *AgentUsageHandler) RecordUsage(c *gin.Context) {
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

	var req models.CreateAgentRunUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid request body"})
		return
	}

	usage, err := h.usageRepo.Record(c.Request.Context(), userID, runID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if usage == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Agent run not found"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// GetUsage godoc
// @Summary Get usage for an agent run
// @Tags agent-usage
// @Produce json
// @Security BearerAuth
// @Param id path int true "Agent run ID"
// @Success 200 {object} models.AgentRunUsage
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/agent-runs/{id}/usage [get]
func (h *AgentUsageHandler) GetUsage(c *gin.Context) {
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

	usage, err := h.usageRepo.GetByRun(c.Request.Context(), userID, runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if usage == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "No usage recorded for this run"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// ListUsage godoc
// @Summary List and aggregate agent usage
// @Description List agent usage rows and an aggregated summary, filtered by agent, model, tool or period.
// @Tags agent-usage
// @Produce json
// @Security BearerAuth
// @Param agent_id query string false "Filter by agent"
// @Param model query string false "Filter by model"
// @Param tool query string false "Filter by tool name"
// @Param from query string false "Period start (RFC3339)"
// @Param to query string false "Period end (RFC3339)"
// @Success 200 {object} models.AgentUsageSummary
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/agent-usage [get]
func (h *AgentUsageHandler) ListUsage(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	filter, err := usageFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	rows, err := h.usageRepo.List(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	summary, err := h.usageRepo.Summary(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
		"rows":    rows,
	})
}

// ListBudgets godoc
// @Summary List agent budgets
// @Tags agent-budgets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.AgentBudget
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/agent-budgets [get]
func (h *AgentUsageHandler) ListBudgets(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	budgets, err := h.budgetRepo.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, budgets)
}

// CreateBudget godoc
// @Summary Create an agent budget
// @Description Create a cost limit for agent runs over a period. A hard limit blocks new runs, a warning limit only warns.
// @Tags agent-budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateAgentBudgetRequest true "Budget to create"
// @Success 201 {object} models.AgentBudget
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/agent-budgets [post]
func (h *AgentUsageHandler) CreateBudget(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	var req models.CreateAgentBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid request body"})
		return
	}
	if req.Period == "" {
		req.Period = models.BudgetPeriodMonth
	}
	if err := models.ValidateAgentBudgetPeriod(req.Period); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}
	if req.HardLimitUSD <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "hard_limit_usd måste vara större än noll"})
		return
	}

	budget, err := h.budgetRepo.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, budget)
}

// UpdateBudget godoc
// @Summary Update an agent budget
// @Description Change the limits of an agent budget or pause it.
// @Tags agent-budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Budget ID"
// @Param request body models.UpdateAgentBudgetRequest true "Changes to apply"
// @Success 200 {object} models.AgentBudget
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/agent-budgets/{id} [patch]
func (h *AgentUsageHandler) UpdateBudget(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	budgetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid budget ID"})
		return
	}

	var req models.UpdateAgentBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid request body"})
		return
	}
	if req.HardLimitUSD != nil && *req.HardLimitUSD <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "hard_limit_usd måste vara större än noll"})
		return
	}

	budget, err := h.budgetRepo.Update(c.Request.Context(), userID, budgetID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, budget)
}

// DeleteBudget godoc
// @Summary Delete an agent budget
// @Tags agent-budgets
// @Produce json
// @Security BearerAuth
// @Param id path int true "Budget ID"
// @Success 204
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/agent-budgets/{id} [delete]
func (h *AgentUsageHandler) DeleteBudget(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	budgetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid budget ID"})
		return
	}

	if err := h.budgetRepo.Delete(c.Request.Context(), userID, budgetID); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// BudgetStatus godoc
// @Summary Show the status of an agent budget
// @Tags agent-budgets
// @Produce json
// @Security BearerAuth
// @Param id path int true "Budget ID"
// @Success 200 {object} models.BudgetStatus
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/agent-budgets/{id}/status [get]
func (h *AgentUsageHandler) BudgetStatus(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	budgetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid budget ID"})
		return
	}

	budgets, err := h.budgetRepo.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	for _, budget := range budgets {
		if budget.ID != budgetID {
			continue
		}
		status, err := h.budgetRepo.Status(c.Request.Context(), userID, budget)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, status)
		return
	}

	c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Budget not found"})
}

// usageFilter läser filtren ur frågan.
func usageFilter(c *gin.Context) (models.AgentUsageFilter, error) {
	filter := models.AgentUsageFilter{
		AgentID: c.Query("agent_id"),
		Model:   c.Query("model"),
		Tool:    c.Query("tool"),
	}

	if raw := c.Query("from"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return filter, err
		}
		filter.From = &parsed
	}
	if raw := c.Query("to"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return filter, err
		}
		filter.To = &parsed
	}

	return filter, nil
}
