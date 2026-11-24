package handlers

import (
	"net/http"
	"strconv"

	"plan-api/internal/models"
	"plan-api/internal/repository"

	"github.com/gin-gonic/gin"
)

// AgentApprovalHandler hanterar mänskliga beslut i agentkörningar.
//
// Agenten får föreslå, men beslutet fattas alltid av en människa. Därför
// ligger beslutet i Plan och inte i agentens eget minne.
type AgentApprovalHandler struct {
	repo      *repository.AgentApprovalRepository
	eventRepo *repository.AgentRunEventRepository
}

func NewAgentApprovalHandler() *AgentApprovalHandler {
	return &AgentApprovalHandler{
		repo:      repository.NewAgentApprovalRepository(),
		eventRepo: repository.NewAgentRunEventRepository(),
	}
}

// ListAgentApprovals godoc
// @Summary List approvals for an agent run
// @Description Get the human decisions waiting on an agent run
// @Tags agent-runs
// @Produce json
// @Security BearerAuth
// @Param id path int true "Agent run ID"
// @Success 200 {array} models.AgentApproval
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/agent-runs/{id}/approvals [get]
func (h *AgentApprovalHandler) ListAgentApprovals(c *gin.Context) {
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

	approvals, err := h.repo.List(c.Request.Context(), userID, runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, approvals)
}

// ListPendingAgentApprovals godoc
// @Summary List approvals waiting for a human decision
// @Description Get every proposal that waits for a human, optionally for one goal
// @Tags agent-runs
// @Produce json
// @Security BearerAuth
// @Param goal_id query int false "Only approvals for this goal"
// @Success 200 {array} models.AgentApproval
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/agent-approvals [get]
func (h *AgentApprovalHandler) ListPendingAgentApprovals(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	ctx := c.Request.Context()

	// Ett mål kan ha flera godkännanden - då visas just det målets.
	if raw := c.Query("goal_id"); raw != "" {
		goalID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid goal ID"})
			return
		}

		approvals, err := h.repo.ListByGoal(ctx, userID, goalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
			return
		}

		c.JSON(http.StatusOK, approvals)
		return
	}

	approvals, err := h.repo.ListPending(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, approvals)
}

// CreateAgentApproval godoc
// @Summary Propose a change for human approval
// @Description Register a proposal that needs a human decision before it is applied
// @Tags agent-runs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Agent run ID"
// @Param request body models.CreateAgentApprovalRequest true "Approval"
// @Success 201 {object} models.AgentApproval
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/agent-runs/{id}/approvals [post]
func (h *AgentApprovalHandler) CreateAgentApproval(c *gin.Context) {
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

	var req models.CreateAgentApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "title is required"})
		return
	}
	req.AgentRunID = runID

	approval, err := h.repo.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if approval == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Agent run not found"})
		return
	}

	c.JSON(http.StatusCreated, approval)
}

// DecideAgentApproval godoc
// @Summary Decide an approval
// @Description Approve or reject a proposal. Only a person can make this decision.
// @Tags agent-runs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Approval ID"
// @Param request body models.DecideAgentApprovalRequest true "Decision"
// @Success 200 {object} models.AgentApproval
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/agent-approvals/{id}/decision [post]
func (h *AgentApprovalHandler) DecideAgentApproval(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	approvalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid approval ID"})
		return
	}

	var req models.DecideAgentApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "decision is required"})
		return
	}
	if err := models.ValidateAgentApprovalDecision(req.Decision); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: err.Error()})
		return
	}

	approval, err := h.repo.Decide(c.Request.Context(), userID, approvalID, req.Decision, req.Feedback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if approval == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Approval not found or already decided"})
		return
	}

	// Beslutet skrivs till körningens historik så att det går att se
	// att en människa fattade det.
	h.recordDecision(c, approval)

	c.JSON(http.StatusOK, approval)
}

// recordDecision skriver beslutet till körningens händelsehistorik.
func (h *AgentApprovalHandler) recordDecision(c *gin.Context, approval *models.AgentApproval) {
	payload := map[string]any{
		"decision":    approval.Status,
		"approval_id": approval.ID,
	}
	if approval.Feedback != "" {
		payload["feedback"] = approval.Feedback
	}

	_, err := h.eventRepo.Create(c.Request.Context(), getUserID(c), approval.AgentRunID, models.CreateAgentRunEventRequest{
		Type:    models.AgentRunEventApprovalDecided,
		Payload: payload,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
	}
}
