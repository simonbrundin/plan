package middleware

import (
	"net/http"

	"plan/api/internal/models"

	"github.com/gin-gonic/gin"
)

// RestrictToReadScope begränsar en läsnyckel till GET-anrop.
//
// Scopet är en teknisk gräns. En widget eller ett skript som bara ska
// visa data ska inte kunna ändra något, även om nyckeln hamnar fel.
func RestrictToReadScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		scope, exists := c.Get("apiKeyScope")
		if !exists || scope != string(models.ScopeReadonly) {
			c.Next()
			return
		}

		if c.Request.Method != http.MethodGet {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Läsnyckeln får bara göra GET-anrop",
			})
			return
		}

		c.Next()
	}
}
