package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestAgentScopeAllows kontrollerar vilka anrop en agent-nyckel får göra.
//
// Allowlisten är agentens behörighet i Plan. Human gate-beslut och
// raderingar ska alltid ligga utanför agentens räckvidd.
func TestAgentScopeAllows(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		{"läsa mål", http.MethodGet, "/api/v1/goals", true},
		{"skapa undermål", http.MethodPost, "/api/v1/goals", true},
		{"läsa ett mål", http.MethodGet, "/api/v1/goals/42", true},
		{"uppdatera ett mål", http.MethodPatch, "/api/v1/goals/42", true},
		{"radera ett mål", http.MethodDelete, "/api/v1/goals/42", false},
		{"sätta målstatus", http.MethodPatch, "/api/v1/goals/42/status", true},
		{"läsa målstatus", http.MethodGet, "/api/v1/goals/42/status", true},
		{"radera målstatus", http.MethodDelete, "/api/v1/goals/42/status", false},
		{"begära fasövergång", http.MethodPost, "/api/v1/goals/42/phase", true},
		{"avsluta mål", http.MethodPost, "/api/v1/goals/42/complete", false},
		{"läsa verifieringsplan", http.MethodGet, "/api/v1/goals/42/verification", true},
		{"sätta verifieringsplan", http.MethodPut, "/api/v1/goals/42/verification", true},
		{"köra verifiering", http.MethodPost, "/api/v1/goals/42/verification/run", true},
		{"lägga bevis", http.MethodPost, "/api/v1/goals/42/evidence", true},
		{"godkänna omdöme", http.MethodPost, "/api/v1/verification-results/3/approve", false},
		{"läsa statusar", http.MethodGet, "/api/v1/statuses", true},
		{"skapa körning", http.MethodPost, "/api/v1/agent-runs", true},
		{"läsa körning", http.MethodGet, "/api/v1/agent-runs/7", true},
		{"uppdatera körning", http.MethodPatch, "/api/v1/agent-runs/7", true},
		{"radera körning", http.MethodDelete, "/api/v1/agent-runs/7", false},
		{"läsa händelser", http.MethodGet, "/api/v1/agent-runs/7/events", true},
		{"skriva händelser", http.MethodPost, "/api/v1/agent-runs/7/events", true},
		{"läsa godkännanden", http.MethodGet, "/api/v1/agent-runs/7/approvals", true},
		{"skapa godkännande", http.MethodPost, "/api/v1/agent-runs/7/approvals", true},
		{"läsa ett godkännande", http.MethodGet, "/api/v1/agent-approvals/3", true},
		{"besluta om godkännande", http.MethodPost, "/api/v1/agent-approvals/3/decision", false},
		{"ta bort kalendrar", http.MethodDelete, "/api/v1/calendars/1", false},
		{"hantera api-nycklar", http.MethodGet, "/api/v1/api-keys", false},
		{"allt annat", http.MethodGet, "/api/v1/users", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AgentScopeAllows(tt.method, tt.path); got != tt.want {
				t.Errorf("AgentScopeAllows(%q, %q) = %v, vill ha %v", tt.method, tt.path, got, tt.want)
			}
		})
	}
}

// TestRestrictToAgentScopeStopparFörbjudnaAnrop kontrollerar att
// middleware faktiskt stoppar anropet.
func TestRestrictToAgentScopeStopparFörbjudnaAnrop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("apiKeyScope", "agent")
		c.Next()
	})
	router.Use(RestrictToAgentScope())
	router.POST("/api/v1/agent-approvals/:id/decision", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	router.GET("/api/v1/goals", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	forbidden := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/agent-approvals/3/decision", nil)
	router.ServeHTTP(forbidden, request)
	if forbidden.Code != http.StatusForbidden {
		t.Errorf("beslut via agent-nyckel gav HTTP %d, vill ha %d", forbidden.Code, http.StatusForbidden)
	}

	allowed := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/v1/goals", nil)
	router.ServeHTTP(allowed, request)
	if allowed.Code != http.StatusOK {
		t.Errorf("läsning av mål gav HTTP %d, vill ha %d", allowed.Code, http.StatusOK)
	}
}

// TestRestrictToAgentScopeSläpperFörbiFulltScope kontrollerar att en vanlig
// nyckel inte påverkas.
func TestRestrictToAgentScopeSläpperFörbiFulltScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("apiKeyScope", "full")
		c.Next()
	})
	router.Use(RestrictToAgentScope())
	router.POST("/api/v1/agent-approvals/:id/decision", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/agent-approvals/3/decision", nil)
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("beslut med fullt scope gav HTTP %d, vill ha %d", recorder.Code, http.StatusOK)
	}
}
