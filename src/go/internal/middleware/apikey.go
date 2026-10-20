package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"plan-api/internal/database"
)

const (
	// APIKeyHeader is the header name for API key authentication
	APIKeyHeader = "X-API-Key"
	
	// APIKeyQueryParam is the query parameter for API key authentication
	APIKeyQueryParam = "api_key"
)

// APIKeyAuth returns middleware that validates API keys
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for API key in header
		apiKey := c.GetHeader(APIKeyHeader)
		
		// Fallback to query parameter
		if apiKey == "" {
			apiKey = c.Query(APIKeyQueryParam)
		}
		
		if apiKey == "" {
			// No API key provided, continue to other auth methods
			c.Next()
			return
		}
		
		// Validate the API key
		userID, err := validateAPIKey(apiKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid API key",
				"message": err.Error(),
			})
			return
		}
		
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			return
		}
		
		// Set user ID in context (similar to session auth)
		c.Set("userID", userID)
		c.Set("authMethod", "api_key")
		
		c.Next()
	}
}

// validateAPIKey validates an API key and returns the associated user ID
func validateAPIKey(apiKey string) (int64, error) {
	// Check for bypass key (development only)
	bypassKey := os.Getenv("API_KEY_BYPASS")
	if bypassKey != "" && apiKey == bypassKey {
		// Return user ID 1 for bypass key
		return 1, nil
	}
	
	// Hash the provided key
	keyHash := hashAPIKey(apiKey)
	
	// Look up in database
	pool := database.GetPool()
	if pool == nil {
		return 0, fmt.Errorf("database not connected")
	}
	
	var userID int64
	err := pool.QueryRow(context.Background(), `
		SELECT COALESCE(user_id, 1) FROM api_keys 
		WHERE key_hash = $1 
		AND (expires IS NULL OR expires > NOW())
	`, keyHash).Scan(&userID)
	if err != nil {
		return 0, nil // Key not found
	}
	
	return userID, nil
}

// hashAPIKey creates a SHA-256 hash of the API key
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// GenerateAPIKey creates a new API key and returns the raw key (only time it's visible)
func GenerateAPIKey(name string, userID int64) (string, error) {
	pool := database.GetPool()
	if pool == nil {
		return "", fmt.Errorf("database not connected")
	}
	
	// Generate a random key
	rawKey := generateRandomKey(32)
	keyHash := hashAPIKey(rawKey)
	
	// Store the hash
	_, err := pool.Exec(context.Background(), `
		INSERT INTO api_keys (name, key_hash, user_id) 
		VALUES ($1, $2, $3)
	`, name, keyHash, userID)
	if err != nil {
		return "", err
	}
	
	return rawKey, nil
}

// generateRandomKey generates a random key string
func generateRandomKey(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	rand.Read(bytes)
	for i := range bytes {
		bytes[i] = charset[int(bytes[i])%len(charset)]
	}
	return string(bytes)
}

// RequireAPIKey returns middleware that requires API key authentication
func RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if already authenticated via API key
		if authMethod, exists := c.Get("authMethod"); exists {
			if authMethod == "api_key" {
				c.Next()
				return
			}
		}
		
		// Check for API key in header
		apiKey := c.GetHeader(APIKeyHeader)
		if apiKey == "" {
			apiKey = c.Query(APIKeyQueryParam)
		}
		
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
			})
			return
		}
		
		userID, err := validateAPIKey(apiKey)
		if err != nil || userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			return
		}
		
		c.Set("userID", userID)
		c.Set("authMethod", "api_key")
		c.Next()
	}
}
