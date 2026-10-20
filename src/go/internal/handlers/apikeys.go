package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"plan-api/internal/database"
	"plan-api/internal/middleware"
)

// APIKey represents an API key (without the actual key)
type APIKey struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Created   time.Time  `json:"created"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
	Expires   *time.Time `json:"expires,omitempty"`
}

// APIKeyResponse is returned when creating a new API key (includes the raw key)
type APIKeyResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Key   string `json:"key"` // Only returned once!
}

// CreateAPIKeyRequest is the request body for creating an API key
type CreateAPIKeyRequest struct {
	Name    string `json:"name" binding:"required"`
	Expires string `json:"expires,omitempty"` // RFC3339 format
}

// APIKeyHandler handles API key management
type APIKeyHandler struct{}

func NewAPIKeyHandler() *APIKeyHandler {
	return &APIKeyHandler{}
}

// ListAPIKeys godoc
// @Summary List API keys
// @Description Get all API keys for the authenticated user
// @Tags api-keys
// @Produce json
// @Security BearerAuth
// @Success 200 {array} APIKey
// @Failure 401 {object} ErrorResponse
// @Router /api/api-keys [get]
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	pool := database.GetPool()
	if pool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not connected"})
		return
	}

	rows, err := pool.Query(context.Background(), `
		SELECT id, name, created, last_used, expires 
		FROM api_keys 
		WHERE user_id = $1 
		ORDER BY created DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var key APIKey
		if err := rows.Scan(&key.ID, &key.Name, &key.Created, &key.LastUsed, &key.Expires); err != nil {
			continue
		}
		keys = append(keys, key)
	}

	if keys == nil {
		keys = []APIKey{}
	}

	c.JSON(http.StatusOK, keys)
}

// CreateAPIKey godoc
// @Summary Create an API key
// @Description Create a new API key for programmatic access
// @Tags api-keys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key body CreateAPIKeyRequest true "API key data"
// @Success 201 {object} APIKeyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/api-keys [post]
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	pool := database.GetPool()
	if pool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not connected"})
		return
	}

	// Generate the API key using middleware helper
	rawKey, err := middleware.GenerateAPIKey(req.Name, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	// Get the ID
	var id int64
	err = pool.QueryRow(context.Background(), `
		SELECT id FROM api_keys WHERE name = $1 AND user_id = $2 
		ORDER BY created DESC LIMIT 1
	`, req.Name, userID).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, APIKeyResponse{
		ID:   id,
		Name: req.Name,
		Key:  rawKey, // Only returned once!
	})
}

// DeleteAPIKey godoc
// @Summary Delete an API key
// @Description Delete an API key
// @Tags api-keys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "API Key ID"
// @Success 200 {object} map[string]bool
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/api-keys/{id} [delete]
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	pool := database.GetPool()
	if pool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not connected"})
		return
	}

	result, err := pool.Exec(context.Background(), `
		DELETE FROM api_keys WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
