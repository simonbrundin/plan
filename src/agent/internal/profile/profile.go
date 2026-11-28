// Package profile beskriver agentprofilernas policy.
//
// Profilen är inte ett stycke text i prompten - den är tekniska gränser
// som gäller oavsett vad modellen säger. En kodprofil får arbeta i en
// isolerad kopia av repot och får skicka brancher; en livsprofil får inte
// röra kod alls.
package profile

import (
	"fmt"
	"strings"
)

// Namnen på profilerna.
const (
	// Coding är profilen för kodändringar.
	Coding = "coding"
	// Life är profilen för agentarbete som inte rör kod.
	Life = "life"
)

// Åtgärdsprofilerna - vad agenten får göra i en viss körning.
//
// De följer uppgiften, inte målets ämne: ett livsmål kan kräva kod och ett
// kodmål kan kräva research. Profilen föreslås av agenten i definitionen
// och godkänns av användaren - därför är standarden den strängaste.
const (
	// ActionPlanering får agenten prata, bryta ned, föreslå och fråga.
	// Den skriver ingenting.
	ActionPlanering = "planering"
	// ActionResearch får agenten läsa och samla underlag. Den skriver
	// ingenting.
	ActionResearch = "research"
	// ActionArbete får agenten arbeta i en isolerad kopia, skriva kod och
	// föreslå ändringar.
	ActionArbete = "arbete"
)

// Policy beskriver vad en profil får göra.
type Policy struct {
	// Name är profilens namn.
	Name string
	// Action är åtgärdsprofilen - vad agenten får göra i körningen.
	//
	// Den är tom för profiler som inte är kopplade till en körning.
	Action string
	// MayChangeCode säger om profilen får ändra filer i repot.
	MayChangeCode bool
	// NeedsRepo säger om jobbet behöver klona repot.
	NeedsRepo bool
	// NeedsHumanGate säger om körningen alltid kräver ett mänskligt
	// beslut innan något skrivs.
	NeedsHumanGate bool
	// CPULimit och MemoryLimit sätter gränserna för jobbet.
	CPULimit    string
	MemoryLimit string
	// MaxTurns begränsar hur många rundor modellen får köra.
	MaxTurns int
}

// PolicyFor ger policyn för en profil.
//
// En okänd profil ger coding - den är den strängaste, och en okänd profil
// ska inte ge mer rättigheter än vad som är avsett.
func PolicyFor(name string) (Policy, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", Coding:
		return CodingPolicy(), nil
	case Life:
		return LifePolicy(), nil
	default:
		return Policy{}, fmt.Errorf("unknown agent profile %q", name)
	}
}

// ActionPolicy ger policyn för en åtgärdsprofil.
//
// Åtgärdsprofilen är det som styr - den tekniska profilen är bara ett
// medel. Planering och research mappar mot livsprofilens rättigheter,
// arbete mot kodprofilens.
//
// En okänd profil ger den strängaste profilen - planering - så att en
// bugg inte ger agenten mer rättigheter än avsett.
func ActionPolicy(action string) (Policy, error) {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "", ActionPlanering:
		policy := LifePolicy()
		policy.Action = ActionPlanering
		return policy, nil
	case ActionResearch:
		policy := LifePolicy()
		policy.Action = ActionResearch
		// Research får läsa mer än planering - men fortfarande inte skriva.
		policy.MaxTurns = 16
		return policy, nil
	case ActionArbete:
		policy := CodingPolicy()
		policy.Action = ActionArbete
		return policy, nil
	default:
		return Policy{}, fmt.Errorf("unknown action profile %q", action)
	}
}

// ActionKnown berättar om namnet är en åtgärdsprofil.
func ActionKnown(action string) bool {
	_, err := ActionPolicy(action)
	return err == nil
}

// ActionNames ger de åtgärdsprofiler som stöds.
func ActionNames() []string {
	return []string{ActionPlanering, ActionResearch, ActionArbete}
}

// CodingPolicy ger policyn för kodändringar.
func CodingPolicy() Policy {
	return Policy{
		Name:           Coding,
		MayChangeCode:  true,
		NeedsRepo:      true,
		NeedsHumanGate: true,
		CPULimit:       "1000m",
		MemoryLimit:    "2Gi",
		MaxTurns:       20,
	}
}

// LifePolicy ger policyn för agentarbete utan kodändringar.
func LifePolicy() Policy {
	return Policy{
		Name:           Life,
		MayChangeCode:  false,
		NeedsRepo:      false,
		NeedsHumanGate: true,
		CPULimit:       "500m",
		MemoryLimit:    "1Gi",
		MaxTurns:       12,
	}
}

// Names ger de profiler som stöds.
func Names() []string {
	return []string{Coding, Life}
}
