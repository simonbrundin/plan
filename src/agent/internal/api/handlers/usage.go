package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"plan/agent/internal/core/usage"
)

// UsageHandler handles token usage endpoints
type UsageHandler struct {
	tracker *usage.Tracker
}

// NewUsageHandler creates a new usage handler
func NewUsageHandler(tracker *usage.Tracker) *UsageHandler {
	return &UsageHandler{tracker: tracker}
}

// UsageStats represents aggregated usage statistics
// @Summary Usage statistics
// @Description Returns aggregated token usage statistics
// @Tags usage
// @Produce json
// @Success 200 {object} usage.Stats
// @Router /api/v1/usage/stats [get]
func (h *UsageHandler) UsageStats(c *gin.Context) {
	if h.tracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "usage tracking not available",
		})
		return
	}

	stats, err := h.tracker.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ConversationUsage returns usage for a specific conversation
// @Summary Conversation usage
// @Description Returns token usage for a specific conversation
// @Tags usage
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {array} usage.TokenUsage
// @Router /api/v1/usage/conversation/{id} [get]
func (h *UsageHandler) ConversationUsage(c *gin.Context) {
	if h.tracker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "usage tracking not available",
		})
		return
	}

	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "conversation id required",
		})
		return
	}

	records, err := h.tracker.GetByConversation(c.Request.Context(), conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversationId": conversationID,
		"usage":          records,
	})
}

// UsageCost represents cost configuration
// @Summary Cost configuration
// @Description Returns the cost configuration per provider/model
// @Tags usage
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/usage/costs [get]
func (h *UsageHandler) CostConfig(c *gin.Context) {
	costs := usage.DefaultCosts()
	result := make([]gin.H, len(costs))
	for i, c := range costs {
		result[i] = gin.H{
			"provider":        c.Provider,
			"model":           c.Model,
			"inputCostPer1M":  c.InputCostPer1M,
			"outputCostPer1M": c.OutputCostPer1M,
		}
	}
	c.JSON(http.StatusOK, result)
}
