package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"plan-agent/internal/approvals"
	"plan-agent/internal/planclient"
)

// ApprovalHandler exposes the pending agent proposals over REST so that
// Plan can show them and let the user approve or reject them.
//
// Slack is not used: the decision is taken in Plan, where the goal lives.
type ApprovalHandler struct {
	approvalStore *approvals.Store
	planClient    *planclient.Client
}

// NewApprovalHandler creates a new approval handler
func NewApprovalHandler(store *approvals.Store, planClient *planclient.Client) *ApprovalHandler {
	return &ApprovalHandler{
		approvalStore: store,
		planClient:    planClient,
	}
}

// DecisionRequest is the body of an approval decision
type DecisionRequest struct {
	// Decision is "approve" or "reject"
	Decision string `json:"decision" binding:"required"`
	// Feedback is optional - what the agent should change
	Feedback string `json:"feedback,omitempty"`
}

// ListApprovals returns the proposals waiting for a decision
// @Summary List pending approvals
// @Tags approvals
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/approvals [get]
func (h *ApprovalHandler) ListApprovals(c *gin.Context) {
	if h.approvalStore == nil {
		c.JSON(http.StatusOK, gin.H{"approvals": []any{}})
		return
	}

	all, err := h.approvalStore.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pending := make([]*approvals.PendingApproval, 0, len(all))
	for _, approval := range all {
		if approval.Status == "" || approval.Status == approvals.StatusPending {
			pending = append(pending, approval)
		}
	}

	c.JSON(http.StatusOK, gin.H{"approvals": pending})
}

// GetApproval returns one proposal
// @Summary Get one pending approval
// @Tags approvals
// @Produce json
// @Param goal_id path int true "Goal ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/v1/approvals/{goal_id} [get]
func (h *ApprovalHandler) GetApproval(c *gin.Context) {
	goalID, ok := parseGoalID(c)
	if !ok {
		return
	}

	if h.approvalStore == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No pending approval found"})
		return
	}

	approval, err := h.approvalStore.Get(goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if approval == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No pending approval found"})
		return
	}

	c.JSON(http.StatusOK, approval)
}

// DecideApproval approves or rejects a proposal
// @Summary Approve or reject a proposal
// @Tags approvals
// @Accept json
// @Produce json
// @Param goal_id path int true "Goal ID"
// @Param decision body DecisionRequest true "Decision"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/v1/approvals/{goal_id}/decision [post]
func (h *ApprovalHandler) DecideApproval(c *gin.Context) {
	goalID, ok := parseGoalID(c)
	if !ok {
		return
	}

	if h.approvalStore == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No pending approval found"})
		return
	}

	var req DecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	approval, err := h.approvalStore.Get(goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if approval == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No pending approval found"})
		return
	}

	switch req.Decision {
	case "approve":
		h.approve(c, goalID, approval)
	case "reject":
		h.reject(c, goalID, approval, req.Feedback)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown decision"})
	}
}

// approve writes the proposed definition to Plan and moves the goal to plan
func (h *ApprovalHandler) approve(c *gin.Context, goalID int64, approval *approvals.PendingApproval) {
	if h.planClient != nil {
		title := approval.Title
		description := approval.Description
		if err := h.planClient.UpdateGoalWithPhase(c.Request.Context(), goalID, &title, &description, "plan"); err != nil {
			log.Printf("Failed to update goal %d in Plan: %v", goalID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update goal in Plan"})
			return
		}
		log.Printf("Goal %d approved, status -> plan", goalID)

		// Mark the agent run as done so Plan shows the decision.
		h.markRun(c, goalID, approval, "plan", "done")
	}

	h.recordDecision(c, goalID, "approve", feedbackOf(approval))

	if err := h.approvalStore.Delete(goalID); err != nil {
		log.Printf("Failed to delete approval for goal %d: %v", goalID, err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "approved", "goal_id": goalID})
}

// reject keeps the feedback so the agent can define the goal again
func (h *ApprovalHandler) reject(c *gin.Context, goalID int64, approval *approvals.PendingApproval, feedback string) {
	approval.Status = approvals.StatusRejected
	approval.Feedback = feedback
	if err := h.approvalStore.Save(approval); err != nil {
		log.Printf("Failed to save rejection for goal %d: %v", goalID, err)
	}

	// The agent picks the goal up again on the next poll and defines it
	// again with the feedback.
	h.markRun(c, goalID, approval, "", "queued")

	h.recordDecision(c, goalID, "reject", feedback)

	c.JSON(http.StatusOK, gin.H{"status": "rejected", "goal_id": goalID, "feedback": feedback})
}

// recordDecision writes the human decision to the run history.
//
// Beslutet fattas av en människa - agenten får aldrig godkänna sig själv,
// så det är här historiken visar att en människa var med.
func (h *ApprovalHandler) recordDecision(c *gin.Context, goalID int64, decision, feedback string) {
	if h.planClient == nil {
		return
	}

	runs, err := h.planClient.GetAgentRuns(c.Request.Context(), &goalID)
	if err != nil {
		log.Printf("Failed to list agent runs for goal %d: %v", goalID, err)
		return
	}
	if len(runs) == 0 {
		return
	}

	payload := map[string]any{"decision": decision}
	if feedback != "" {
		payload["feedback"] = feedback
	}
	err = h.planClient.CreateAgentRunEvent(c.Request.Context(), runs[0].ID, planclient.CreateAgentRunEventRequest{
		Type:    planclient.EventApprovalDecided,
		Payload: payload,
	})
	if err != nil {
		log.Printf("Failed to record approval decision for goal %d: %v", goalID, err)
	}
}

// feedbackOf returns the feedback on an approval, om det finns.
func feedbackOf(approval *approvals.PendingApproval) string {
	if approval == nil {
		return ""
	}
	return approval.Feedback
}

// markRun reflects the decision in the agent run in Plan
func (h *ApprovalHandler) markRun(c *gin.Context, goalID int64, approval *approvals.PendingApproval, phase, status string) {
	if h.planClient == nil {
		return
	}

	runs, err := h.planClient.GetAgentRuns(c.Request.Context(), &goalID)
	if err != nil {
		log.Printf("Failed to list agent runs for goal %d: %v", goalID, err)
		return
	}
	if len(runs) == 0 {
		return
	}

	run := runs[0]
	summary := run.Summary
	if summary == "" {
		summary = approval.Title
	}
	req := planclient.UpdateAgentRunRequest{Status: &status, Summary: &summary}
	if phase != "" {
		req.Phase = &phase
	}
	if _, err := h.planClient.UpdateAgentRun(c.Request.Context(), run.ID, req); err != nil {
		log.Printf("Failed to update agent run %d: %v", run.ID, err)
	}
}

// parseGoalID reads and validates the goal id from the path
func parseGoalID(c *gin.Context) (int64, bool) {
	goalID, err := strconv.ParseInt(c.Param("goal_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid goal ID: %v", err)})
		return 0, false
	}
	return goalID, true
}
