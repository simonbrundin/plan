package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the Bearer token and extracts user ID
// For now, we use a simple JWT-like validation. This should be replaced
// with proper Zitadel/OIDC validation in production.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Development mode: skip authentication
		if os.Getenv("AUTH_DISABLED") == "true" {
			// Default to user 4 (simonbrundin@gmail.com) who has test data
			c.Set("userID", int64(4))
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization header required"})
			c.Abort()
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]

		// For development/testing, accept tokens in format "user_<id>"
		// Example: "user_123" → userID = 123
		if strings.HasPrefix(token, "user_") {
			var userID int64
			for _, ch := range token[5:] {
				if ch < '0' || ch > '9' {
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user token"})
					c.Abort()
					return
				}
				userID = userID*10 + int64(ch-'0')
			}
			c.Set("userID", userID)
			c.Next()
			return
		}

		// TODO: Implement proper JWT validation for Zitadel
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
		c.Abort()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOrigin == "" {
			allowedOrigin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
