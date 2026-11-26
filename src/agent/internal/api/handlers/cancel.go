package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"plan/agent/internal/core/tools"
)

// CancelRequest represents a request to cancel a tool execution
// @Description Request to cancel a running tool execution
type CancelRequest struct {
	ToolCallID string `json:"toolCallId" binding:"required"`
}

// CancelResponse represents the response from a cancel request
// @Description Response from a cancel request
type CancelResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ToolCallID string `json:"toolCallId"`
}

// CancelTool cancels a running tool execution
// @Summary Cancel tool execution
// @Description Cancels a running tool execution by tool call ID
// @Tags tools
// @Accept json
// @Produce json
// @Param request body CancelRequest true "Cancel request"
// @Success 200 {object} CancelResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/tools/cancel [post]
func CancelTool(c *gin.Context) {
	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "toolCallId is required",
		})
		return
	}

	manager := tools.GetGlobalCancellationManager()

	// Check if the tool call exists
	if !manager.IsCancelled(req.ToolCallID) && manager.Count() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "tool execution not found",
			"message": "No running tool execution found with this ID",
		})
		return
	}

	// Cancel the tool execution
	manager.Cancel(req.ToolCallID)

	c.JSON(http.StatusOK, CancelResponse{
		Success:    true,
		Message:    "Cancellation signal sent",
		ToolCallID: req.ToolCallID,
	})
}

// CancelAllTools cancels all running tool executions
// @Summary Cancel all tool executions
// @Description Cancels all running tool executions
// @Tags tools
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tools/cancel-all [post]
func CancelAllTools(c *gin.Context) {
	manager := tools.GetGlobalCancellationManager()
	count := manager.Count()

	manager.CancelAll()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "Cancellation signals sent to all running tools",
		"cancelledCount": count,
	})
}

// ToolStatus provides information about running tools
// @Summary Tool execution status
// @Description Returns the number of running tool executions
// @Tags tools
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tools/status [get]
func ToolStatus(c *gin.Context) {
	manager := tools.GetGlobalCancellationManager()

	c.JSON(http.StatusOK, gin.H{
		"runningCount": manager.Count(),
	})
}
