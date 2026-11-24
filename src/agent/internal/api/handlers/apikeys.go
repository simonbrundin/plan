package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"plan-agent/internal/api/db"
)

// ListAPIKeys returns all configured API keys (without the actual key values)
// @Summary List API keys
// @Tags api-keys
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/api-keys [get]
func ListAPIKeys(c *gin.Context) {
	ctx := context.Background()
	keys, err := db.GetAPIKeys(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return keys with masked values
	type maskedKey struct {
		ID       int    `json:"id"`
		Provider string `json:"provider"`
		HasKey   bool   `json:"has_key"`
		Model    string `json:"model"`
	}
	masked := make([]maskedKey, len(keys))
	for i, k := range keys {
		masked[i] = maskedKey{
			ID:       k.ID,
			Provider: k.Provider,
			HasKey:   k.APIKey != "",
			Model:    k.Model,
		}
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": masked})
}

// SetAPIKeyRequest is the request body for setting an API key
type SetAPIKeyRequest struct {
	Provider string `json:"provider" binding:"required"`
	APIKey   string `json:"api_key" binding:"required"`
	Model    string `json:"model"`
}

// SetAPIKey stores or updates an API key
// @Summary Set an API key
// @Tags api-keys
// @Accept json
// @Produce json
// @Param request body SetAPIKeyRequest true "API key request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/api-keys [post]
func SetAPIKey(c *gin.Context) {
	var req SetAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	if err := db.SetAPIKey(ctx, req.Provider, req.APIKey, req.Model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DeleteAPIKey removes an API key
// @Summary Delete an API key
// @Tags api-keys
// @Produce json
// @Param provider path string true "Provider name"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/api-keys/{provider} [delete]
func DeleteAPIKey(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	ctx := context.Background()
	if err := db.DeleteAPIKey(ctx, provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
