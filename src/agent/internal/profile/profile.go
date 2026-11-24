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

// Policy beskriver vad en profil får göra.
type Policy struct {
	// Name är profilens namn.
	Name string
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
