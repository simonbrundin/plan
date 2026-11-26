package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"plan/agent/internal/core/workflow"
)

// WorkflowHandler handles workflow-related API requests
type WorkflowHandler struct {
	commands *workflow.WorkflowCommands
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(o *workflow.Orchestrator) *WorkflowHandler {
	return &WorkflowHandler{
		commands: workflow.NewWorkflowCommands(o),
	}
}

// SetOrchestrator sets the orchestrator for the handler (for deferred initialization)
func (h *WorkflowHandler) SetOrchestrator(o *workflow.Orchestrator) {
	h.commands = workflow.NewWorkflowCommands(o)
}

// StartWorkflowRequest is the request body for starting a workflow
type StartWorkflowRequest struct {
	Feature     string `json:"feature" binding:"required"`
	Description string `json:"description"`
}

// ApprovePlanRequest is the request body for approving a plan
type ApprovePlanRequest struct {
	ApprovedBy string `json:"approvedBy"`
}

// SubmitDecisionRequest is the request body for submitting a human decision
type SubmitDecisionRequest struct {
	CheckpointID string `json:"checkpointId" binding:"required"`
	Approved     bool   `json:"approved"`
	Feedback     string `json:"feedback"`
}

// StartWorkflow starts a new workflow
// @Summary Start a new workflow
// @Description Creates a checkpoint plan for a feature
// @Tags workflow
// @Accept json
// @Produce json
// @Param request body StartWorkflowRequest true "Workflow request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow [post]
func (h *WorkflowHandler) StartWorkflow(c *gin.Context) {
	var req StartWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan, err := h.commands.StartWorkflow(c.Request.Context(), req.Feature, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan":    plan,
		"message": "Plan created. Review and approve at /api/v1/workflow/" + plan.ID + "/approve",
	})
}

// GetPlan returns a plan by ID
// @Summary Get plan details
// @Description Returns details of a specific plan
// @Tags workflow
// @Produce json
// @Param id path string true "Plan ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/{id} [get]
func (h *WorkflowHandler) GetPlan(c *gin.Context) {
	planID := c.Param("id")

	plan, err := h.commands.GetPlanStatus(planID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// ApprovePlan approves a plan
// @Summary Approve a plan
// @Description Marks a plan as approved
// @Tags workflow
// @Produce json
// @Param id path string true "Plan ID"
// @Param request body ApprovePlanRequest false "Approval request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/{id}/approve [post]
func (h *WorkflowHandler) ApprovePlan(c *gin.Context) {
	planID := c.Param("id")

	var req ApprovePlanRequest
	// Kroppen är valfri - ett fel här betyder bara att den saknas.
	_ = c.ShouldBindJSON(&req)

	approvedBy := req.ApprovedBy
	if approvedBy == "" {
		approvedBy = "api"
	}

	if err := h.commands.ApprovePlan(planID, approvedBy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan approved"})
}

// RejectPlan rejects a plan
// @Summary Reject a plan
// @Description Marks a plan as rejected
// @Tags workflow
// @Produce json
// @Param id path string true "Plan ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/{id}/reject [post]
func (h *WorkflowHandler) RejectPlan(c *gin.Context) {
	planID := c.Param("id")

	if err := h.commands.RejectPlan(planID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan rejected"})
}

// GetCurrentCheckpoint returns the current checkpoint
// @Summary Get current checkpoint
// @Description Returns the currently executing checkpoint
// @Tags workflow
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/current/checkpoint [get]
func (h *WorkflowHandler) GetCurrentCheckpoint(c *gin.Context) {
	cp := h.commands.GetCurrentCheckpoint()
	if cp == nil {
		c.JSON(http.StatusOK, gin.H{"checkpoint": nil, "message": "No active checkpoint"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"checkpoint": cp})
}

// GetCheckpointProgress returns progress for the current checkpoint
// @Summary Get checkpoint progress
// @Description Returns gate progress for the current checkpoint
// @Tags workflow
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/current/progress [get]
func (h *WorkflowHandler) GetCheckpointProgress(c *gin.Context) {
	passed, total := h.commands.GetCheckpointProgress()

	c.JSON(http.StatusOK, gin.H{
		"passed":  passed,
		"total":   total,
		"percent": float64(passed) / float64(total) * 100,
	})
}

// SubmitHumanDecision submits a human decision
// @Summary Submit human decision
// @Description Submit approval or rejection for a checkpoint
// @Tags workflow
// @Accept json
// @Produce json
// @Param request body SubmitDecisionRequest true "Decision request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/decision [post]
func (h *WorkflowHandler) SubmitHumanDecision(c *gin.Context) {
	var req SubmitDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.commands.SubmitHumanDecision(req.CheckpointID, req.Approved, req.Feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := "Checkpoint rejected, changes requested"
	if req.Approved {
		message = "Checkpoint approved"
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}

// ListPlans returns all plans
// @Summary List all plans
// @Description Returns all checkpoint plans
// @Tags workflow
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/workflow/plans [get]
func (h *WorkflowHandler) ListPlans(c *gin.Context) {
	o := workflow.GetOrchestrator()
	if o == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Orchestrator not initialized"})
		return
	}

	store := o.StateStore()
	if store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "State store not available"})
		return
	}

	plans, err := store.GetAllPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}
