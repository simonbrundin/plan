package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRequireAPIKey kontrollerar att tjänsten kräver nyckel.
//
// Godkännanden fattas här, så utan nyckel kan någon besluta åt en människa.
func TestRequireAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("AGENT_API_KEY", "hemlighet")

	router := gin.New()
	router.Use(RequireAPIKey())
	router.POST("/api/v1/approvals/:goal_id/decision", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	tests := []struct {
		name   string
		method string
		path   string
		header string
		want   int
	}{
		{"utan nyckel", http.MethodPost, "/api/v1/approvals/42/decision", "", http.StatusUnauthorized},
		{"fel nyckel", http.MethodPost, "/api/v1/approvals/42/decision", "Bearer fel", http.StatusUnauthorized},
		{"rätt nyckel", http.MethodPost, "/api/v1/approvals/42/decision", "Bearer hemlighet", http.StatusOK},
		{"nyckel utan prefix", http.MethodPost, "/api/v1/approvals/42/decision", "hemlighet", http.StatusUnauthorized},
		{"hälsokontroll utan nyckel", http.MethodGet, "/api/v1/health", "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.header != "" {
				request.Header.Set(BearerHeader, tt.header)
			}
			router.ServeHTTP(recorder, request)
			if recorder.Code != tt.want {
				t.Errorf("%s %s gav HTTP %d, vill ha %d", tt.method, tt.path, recorder.Code, tt.want)
			}
		})
	}
}

// TestRequireAPIKeyUtanKonfigureradNyckel kontrollerar att en tjänst utan
// nyckel inte låser ut sig själv.
func TestRequireAPIKeyUtanKonfigureradNyckel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("AGENT_API_KEY", "")

	router := gin.New()
	router.Use(RequireAPIKey())
	router.GET("/api/v1/conversations", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/conversations", nil)
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("anrop utan konfigurerad nyckel gav HTTP %d, vill ha %d", recorder.Code, http.StatusOK)
	}
}
