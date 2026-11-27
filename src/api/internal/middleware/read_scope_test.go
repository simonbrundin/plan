package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"plan/api/internal/models"

	"github.com/gin-gonic/gin"
)

// TestRestrictToReadScopeStopparSkrivning kontrollerar att en läsnyckel
// inte kan ändra något.
func TestRestrictToReadScopeStopparSkrivning(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("apiKeyScope", string(models.ScopeReadonly))
		c.Next()
	})
	router.Use(RestrictToReadScope())
	router.GET("/api/v1/goals/now", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	router.PATCH("/api/v1/goals/1", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	router.DELETE("/api/v1/goals/1", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// Läsning går igenom.
	read := httptest.NewRecorder()
	router.ServeHTTP(read, httptest.NewRequest(http.MethodGet, "/api/v1/goals/now", nil))
	if read.Code != http.StatusOK {
		t.Fatalf("GET gav %d, vill ha 200", read.Code)
	}

	// Skrivning stoppas.
	for _, method := range []string{http.MethodPatch, http.MethodDelete, http.MethodPost, http.MethodPut} {
		write := httptest.NewRecorder()
		router.ServeHTTP(write, httptest.NewRequest(method, "/api/v1/goals/1", nil))
		if write.Code != http.StatusForbidden {
			t.Errorf("%s gav %d, vill ha 403", method, write.Code)
		}
	}
}

// TestRestrictToReadScopeSläpperIgenAndraScopes kontrollerar att
// middleware inte påverkar nycklar med andra scopes.
func TestRestrictToReadScopeSläpperIgenAndraScopes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, scope := range []models.APIKeyScope{models.ScopeFull, models.ScopeAgent} {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("apiKeyScope", string(scope))
			c.Next()
		})
		router.Use(RestrictToReadScope())
		router.PATCH("/api/v1/goals/1", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

		write := httptest.NewRecorder()
		router.ServeHTTP(write, httptest.NewRequest(http.MethodPatch, "/api/v1/goals/1", nil))
		if write.Code != http.StatusOK {
			t.Errorf("scope %s: PATCH gav %d, vill ha 200", scope, write.Code)
		}
	}
}
