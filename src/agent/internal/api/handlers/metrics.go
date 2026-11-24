package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"plan-agent/internal/api/metrics"
)

// MetricsHandler handles metrics endpoints
type MetricsHandler struct{}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// PrometheusMetrics returns Prometheus-formatted metrics
// @Summary Prometheus metrics
// @Description Returns metrics in Prometheus format
// @Tags metrics
// @Produce text/plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics [get]
func (h *MetricsHandler) PrometheusMetrics(c *gin.Context) {
	m := metrics.Get()
	c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(m.PrometheusOutput()))
}

// MetricsJSON returns metrics in JSON format
// @Summary Metrics as JSON
// @Description Returns metrics in JSON format
// @Tags metrics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/metrics [get]
func (h *MetricsHandler) MetricsJSON(c *gin.Context) {
	m := metrics.Get()

	c.JSON(http.StatusOK, gin.H{
		"http": gin.H{
			"requests_in_flight": m.HTTPRequestsInFlight(),
		},
		"agent": gin.H{
			"turns_total":      0, // Would need atomic access for live stats
			"tool_calls_total": 0,
			"errors_total":     0,
		},
	})
}
