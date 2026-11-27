package contracts

import (
	"strings"
	"testing"

	"plan/api/internal/middleware"
	"plan/api/internal/phase"
)

// TestFasovergangarnaMatcharMaskinen kontrollerar att kontraktets
// beskrivning av fasövergångarna stämmer med maskinen som faktiskt
// avgör.
//
// Kontraktet är det agenten och människan läser. Om det säger en annan
// skälkod än API:t svarar med blir felet nästan omöjligt att hitta - och
// agenten kan inte veta om den ska be om hjälp eller göra om något.
func TestFasovergangarnaMatcharMaskinen(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	if len(contract.PhaseTransitions) == 0 {
		t.Fatal("kontraktet ska beskriva fasövergångarna")
	}

	for _, övergång := range contract.PhaseTransitions {
		if !phase.Known(övergång.From) {
			t.Errorf("kontraktet har en okänd fas %q", övergång.From)
		}
		if !phase.Known(övergång.To) {
			t.Errorf("kontraktet har en okänd fas %q", övergång.To)
		}

		// Utan förutsättningarna ska maskinen stoppa övergången med
		// exakt den skälkod kontraktet anger.
		resultat := phase.Evaluate(övergång.From, övergång.To, phase.Facts{IsHuman: false})
		if resultat.Allowed {
			t.Errorf("övergången %s -> %s ska kräva %s", övergång.From, övergång.To, övergång.Requires)
			continue
		}
		if resultat.Code != övergång.Reason {
			t.Errorf("övergången %s -> %s ger skälet %q i API:t, men kontraktet säger %q",
				övergång.From, övergång.To, resultat.Code, övergång.Reason)
		}
	}
}

// TestAgentensAnropAllowsForAgentScope kontrollerar att varje anrop
// agenten gör enligt kontraktet går att göra med agent-scope.
//
// Det är den viktigaste kontrollen i kontraktet: i prod har agentens nyckel
// scope 'agent', och ett anrop som inte är tillåtet ger 403 - ett fel
// som bara syns i prod.
func TestAgentensAnropAllowsForAgentScope(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	for _, endpoint := range contract.AgentEndpoints() {
		if !middleware.AgentScopeAllows(endpoint.Method, endpoint.Path) {
			t.Errorf("kontraktet markerar %s %s som agentens anrop, men agent-scope stoppar det",
				endpoint.Method, endpoint.Path)
		}
	}
}

// TestManskligaAnropInteAgentens kontrollerar att de anrop som ska
// avgöras av en människa inte är markerade som agentens egna.
//
// Att avsluta ett mål och att godkänna ett verifieringsomdöme är
// människans beslut. Markeras de som agentens anrop öppnas human gate
// utan att någon märker det.
func TestManskligaAnropInteAgentens(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	manskligaAnrop := []string{
		"POST /api/v1/goals/{id}/complete",
		"POST /api/v1/verification-results/{id}/approve",
	}

	for _, anrop := range manskligaAnrop {
		delar := strings.SplitN(anrop, " ", 2)
		if contract.HasEndpoint(delar[0], delar[1]) {
			for _, endpoint := range contract.Endpoints {
				if endpoint.Method == delar[0] && endpoint.Path == delar[1] && endpoint.Agent {
					t.Errorf("%s ska avgöras av en människa, men är markerat som agentens anrop", anrop)
				}
			}
		}
	}
}
