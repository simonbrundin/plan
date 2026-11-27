package models

// APIKeyScope bestämmer vad en API-nyckel får göra i Plan.
//
// Scopet är en teknisk gräns. En agent ska kunna arbeta med sina mål och
// körningar, men aldrig fatta ett human gate-beslut eller rensa bort något
// som en människa äger.
type APIKeyScope string

const (
	// ScopeFull ger nyckeln samma rättigheter som en inloggad användare.
	// Det är standard för nycklar som skapats av en människa.
	ScopeFull APIKeyScope = "full"

	// ScopeAgent begränsar nyckeln till agentens arbete med mål och
	// agentkörningar.
	ScopeAgent APIKeyScope = "agent"

	// ScopeReadonly ger nyckeln rätt att läsa, men inte att skriva.
	// Den används av integrationer som bara visar data - t.ex. en
	// widget på skrivbordet.
	ScopeReadonly APIKeyScope = "readonly"
)

// ValidAPIKeyScope svarar på om ett scope är känt.
func ValidAPIKeyScope(scope string) bool {
	switch APIKeyScope(scope) {
	case ScopeFull, ScopeAgent, ScopeReadonly:
		return true
	default:
		return false
	}
}
