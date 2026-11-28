package middleware

import (
	"strings"
	"testing"

	"plan/api/internal/contracts"
)

// TestAgentScopeTillaterKontraktetsAnrop kontrollerar att Plan tillåter
// varje anrop agenten gör.
//
// Agenten autentiserar med sin egen API-nyckel, som i prod har scope
// 'agent'. Då gäller allowlistan i agent_scope.go - inte 'full'. Ett anrop
// som kontraktet säger att agenten gör, men som allowlistan inte tillåter,
// ger 403 i prod. I dev har nyckeln scope 'full', så felet syns inte där.
//
// Testet fångar den typen av gap: kontraktet och allowlistan kan inte
// glida isär.
func TestAgentScopeTillaterKontraktetsAnrop(t *testing.T) {
	contract, err := contracts.Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	agentens := contract.AgentEndpoints()
	if len(agentens) == 0 {
		t.Fatal("kontraktet ska markera vilka anrop agenten gör")
	}

	for _, endpoint := range agentens {
		// {id} i kontraktet blir ett konkret id i ett anrop.
		sökväg := strings.ReplaceAll(endpoint.Path, "{id}", "1")

		if !AgentScopeAllows(endpoint.Method, sökväg) {
			t.Errorf("agenten gör %s %s, men agent-scope tillåter det inte", endpoint.Method, sökväg)
		}
	}
}
