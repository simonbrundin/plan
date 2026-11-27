package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AgentScopeAllows svarar på om en agent-nyckel får göra anropet.
//
// Allowlisten är medvetet liten. Agenten får arbeta med mål och
// agentkörningar, men får aldrig besluta om ett godkännande - det är
// human gate, och den ska alltid göras av en människa. Den får inte heller
// radera något, eftersom allt som en människa har skapat ska överleva.
func AgentScopeAllows(method, path string) bool {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) < 3 || segments[0] != "api" || segments[1] != "v1" {
		return false
	}

	resource := segments[2]
	rest := segments[3:]

	switch resource {
	case "statuses":
		return method == http.MethodGet

	case "places":
		// Agenten behöver känna till användarens platser för att kunna
		// föreslå platser på ett mål. Den får inte skapa eller radera
		// platser.
		return method == http.MethodGet

	case "goals":
		return AgentScopeAllowsGoals(method, rest)

	case "agent-runs":
		return AgentScopeAllowsAgentRuns(method, rest)

	case "agent-approvals":
		// Läsa ett godkännande är ok, att besluta om det är det inte.
		return method == http.MethodGet && len(rest) == 1

	case "verification-results":
		// Ett verifieringsomdöme godkänns alltid av en människa - en
		// agent får inte godkänna sig själv.
		return false
	}

	return false
}

// AgentScopeAllowsGoals gäller målen.
func AgentScopeAllowsGoals(method string, rest []string) bool {
	switch len(rest) {
	case 0:
		// Lista och skapa mål - undermål skapas genom Plan, inte genom
		// filsystemet.
		return method == http.MethodGet || method == http.MethodPost
	case 1:
		// "relations" är en samling, inte ett mål-id. Dit lägger agenten
		// undermål - ett skrivande anrop som hör ihop med målen, så det
		// tillåts.
		if rest[0] == "relations" {
			return method == http.MethodPost
		}
		return method == http.MethodGet || method == http.MethodPatch
	case 2:
		switch rest[1] {
		case "status":
			// Statusen på ett mål.
			return method == http.MethodGet || method == http.MethodPatch
		case "phase":
			// Fasmaskinen. Agenten får begära en övergång - villkoren
			// kontrolleras i API:t, precis som för en människa. Men
			// agenten får inte avsluta ett mål själv, så "complete"
			// (som ligger under målen, inte under ett mål) är inte
			// tillåtet.
			return method == http.MethodPost
		case "verification":
			// Verifieringsplanen och körningen av den. Agenten får
			// föreslå en plan och samla bevis - men inte godkänna ett
			// omdöme.
			return method == http.MethodGet || method == http.MethodPut || method == http.MethodPost
		case "evidence":
			// Bevis på målet.
			return method == http.MethodPost
		case "tags", "places":
			// Agenten får sätta taggar och platser på ett mål - de är
			// en del av måldefinitionen. Den får inte radera dem.
			return method == http.MethodPut
		}
	case 3:
		// "/verification/run" - köra verifieringen. Agenten får göra det,
		// men omdömet blir först giltigt när en människa godkänner det.
		if rest[1] == "verification" && rest[2] == "run" {
			return method == http.MethodPost
		}
	}

	return false
}

// AgentScopeAllowsAgentRuns gäller agentkörningarna och deras händelser
// och godkännanden.
func AgentScopeAllowsAgentRuns(method string, rest []string) bool {
	switch len(rest) {
	case 0:
		return method == http.MethodGet || method == http.MethodPost
	case 1:
		return method == http.MethodGet || method == http.MethodPatch
	case 2:
		switch rest[1] {
		case "events":
			return method == http.MethodGet || method == http.MethodPost
		case "approvals":
			return method == http.MethodGet || method == http.MethodPost
		case "usage":
			// Agenten rapporterar sin användning. Den får inte radera
			// eller ändra den i efterhand.
			return method == http.MethodGet || method == http.MethodPost
		}
	}

	return false
}

// RestrictToAgentScope begränsar en agent-nyckel till agentens arbete.
//
// Nycklar med fullt scope passerar ohindrad. Nycklar med agent-scope får
// bara de anrop som agenten faktiskt behöver - allt annat stoppas här, i
// API:t, och inte i en prompt.
func RestrictToAgentScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		scope, exists := c.Get("apiKeyScope")
		if !exists || scope != "agent" {
			c.Next()
			return
		}

		if !AgentScopeAllows(c.Request.Method, c.Request.URL.Path) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Agent API-nyckeln får inte göra det här anropet",
			})
			return
		}

		c.Next()
	}
}
